import json
import os
import six
import numpy as np
import tensorflow as tf
from six.moves import zip  # pylint: disable=redefined-builtin
from utils import aggregation_fn
from collections import OrderedDict


def find_trained_checkpoints(model_directory_name):
    checkpoints = []
    for f in tf.gfile.ListDirectory(model_directory_name):
        if f.endswith(".index") and "initial" not in f: #  Find index files (jointly made with .data files)
            path = os.path.join(model_directory_name, f)
            checkpoints.append(path.split(".")[0]) # Remove extension
    return checkpoints if len(checkpoints) > 0 else None


STEP_FILE_BASENAME = "model.steps"
def get_ckpts(model_directory_names, debug=False):
    # # Get checkpoints to merge
    # ckpt = tf.train.get_checkpoint_state(model_directory_name)
    # ckpt_merge_list = ckpt.all_model_checkpoint_paths

    # Get checkpoint to merge - this version avoids TF API issues from above
    # TODO: revisit^
    ckpt_merge_dict = OrderedDict() # Necessary for cumulative_steps to work
    cumulative_steps = 0
    for index, model_dir_name in enumerate(model_directory_names):
        key = find_trained_checkpoints(model_dir_name)[0]
        step_filename = os.path.join(model_dir_name, STEP_FILE_BASENAME)
        # NOTE: Parallelize this I/O?
        with tf.gfile.Open(step_filename, 'r') as step_fp:
            steps = int(step_fp.read()) # NOTE: For now, assume single integer txt file for simplicity
        ckpt_merge_dict[key] = {
            'steps': steps,
            'index': index,
            'cumulative_steps': cumulative_steps
        }
        # Update sumulative AFTER, does not include CURRENT steps
        cumulative_steps += steps

    # Error check
    if not ckpt_merge_dict or len(ckpt_merge_dict) == 0:
        raise ValueError("No checkpoints provided for averaging.")
    if debug:
        print ("Reading variables and staging to aggregate checkpoints: ")
        for c in ckpt_merge_dict:
            print (json.dumps(c, indent=4))
    return ckpt_merge_dict, cumulative_steps


def aggregate_ckpts(ckpt_merge_dict, aggregation_function=aggregation_fn._aggregate_cumulative_moving_average, debug=False):
    # Grab variables/weights from one of the ckpts
    var_list = tf.contrib.framework.list_variables(list(ckpt_merge_dict.keys())[0])
    var_values, var_dtypes = {}, {}

    # Iterate through all checkpoints and add variables/weights values
    for c, ckpt_meta in ckpt_merge_dict.items():
        # Load checkpoint
        reader = tf.contrib.framework.load_checkpoint(c)
        # Aggregate variables/weights
        for (name, shape) in var_list:
            if name not in var_values.keys() and name not in var_dtypes.keys(): # 2nd condition as sanity
                var_values[name] = np.zeros(shape)
            tensor = reader.get_tensor(name)
            var_dtypes[name] = tensor.dtype
            var_values[name] = aggregation_function(var_values[name], tensor, ckpt_meta)

        if debug:
            print ("Checkpoint has been aggregated:  {0}".format(c))
    return var_values, var_dtypes


def check_or_create_dir(output_path):
    if not tf.gfile.IsDirectory(output_path): # exists?
        tf.gfile.MkDir(output_path)


# TODO: update "model.steps"
CHECKPOINT_BASENAME = "checkpoint"
def apply_aggregation(output_path, var_values, var_dtypes, total_steps, debug=False):
    # Create output checkpoint/steps path
    check_or_create_dir(output_path)
    aggregated_checkpoint_path = os.path.join(output_path, CHECKPOINT_BASENAME)
    aggregated_steps_path = os.path.join(output_path, STEP_FILE_BASENAME)
    # Get/Create original Tensorflow variables
    tf.reset_default_graph()
    tf_vars = [tf.get_variable(v, shape=var_values[v].shape, dtype=var_dtypes[v]) for v in var_values]
    # Create respective placeholders to feed in Numpy values
    placeholders = [tf.placeholder(v.dtype, shape=v.shape) for v in tf_vars]
    # Assign fed values to original Tensorflow variables
    assign_ops = [tf.assign(v, p) for (v, p) in zip(tf_vars, placeholders)]
    # Create Saver to execute sync
    saver = tf.train.Saver(tf.global_variables(), max_to_keep=1)
    # Build a model consisting only of variables and set to average values
    with tf.Session() as session:
        # Initialize computational graph
        session.run(tf.global_variables_initializer())
        # Run built Tensorflow computational graph to assign variables/weights
        for p, assign_op, (name, value) in zip(placeholders, assign_ops, six.iteritems(var_values)):
            session.run(assign_op, {p: value})

        # Save the averaged checkpoint: write_meta_graph=True
        save_path = saver.save(session, aggregated_checkpoint_path, global_step=0, write_meta_graph=False)
        if debug:
            print("Averaged model saved in file: {0}".format(save_path))
    # Save total steps
    with tf.gfile.Open(aggregated_steps_path, "w") as steps_fp:
        steps_fp.write(str(total_steps))


