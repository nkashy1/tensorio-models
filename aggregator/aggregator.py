import json
import jsonschema
import os
import six
import numpy as np
import tempfile
import tensorflow as tf
import zipfile
from six.moves import zip  # pylint: disable=redefined-builtin
from utils import ckpts
from collections import OrderedDict, defaultdict
from utils import ckpts
from tensorio_bundler import bundler
import time
import requests
import wget
from tensorflow.python.saved_model import loader
import logging
import copy


CHECKPOINT_BASENAME = "checkpoint"
CHECKPOINT_PB_BASENAME = "variables"
BUNDLE_CHECKPOINT_DIRECTORY = "checkpoints"
RESULT_JSON_BASENAME = "result.json"
SAVED_MODEL_BASENAME = "saved_model.pb"
MODEL_JSON_BASENAME = "model.json"
TIO_PREFIX = "tio://"


class AggregatorNotFoundException(Exception):
    pass


class SchemaFailedException(Exception):
    pass


class TensorIOModelsRepositoryException(Exception):
    pass


class Aggregator(object):

    def __init__(
        self,
        resource_path,
        output_resource_path,
        repository,
        token,
        export_type,
        aggregator_type,
        temp_dir="/tmp/",
        schema_path="./schemas/schema.json"
    ):
        self.resource_path = resource_path
        self.output_resource_path = output_resource_path
        self.repository = repository
        self.token = token
        self.checkpoint_basename = CHECKPOINT_BASENAME if export_type == 'checkpoint' else CHECKPOINT_PB_BASENAME
        self.temp_dir = temp_dir

        # Initialize logger
        self.logger = logging.getLogger('aggregator')
        self.logger.setLevel(logging.DEBUG) # TODO: env var or cli to set this
        self.main_handler = logging.StreamHandler()
        self.main_formatter = logging.Formatter('%(levelname)s - %(asctime)s - %(name)s: %(message)s')
        self.main_handler.setFormatter(self.main_formatter)
        self.logger.addHandler(self.main_handler)

        self.logger.info("Configured to talk to repository: {0}".format(repository))
        self.logger.info("Checkpoint will be READ from: {0}".format(resource_path))
        self.logger.info("Checkpoint will be WRITTEN to: {0}".format(output_resource_path))
        self.logger.debug("Using temporary directory: {0}".format(temp_dir))
        self.logger.debug("Using schema: {0}".format(schema_path))

        # Cumulative Moving Average, Weighted Cumulative Moving Average
        self.aggregator_step_functions = {
            "CumulativeMovingAverage": self._cumulative_moving_average_steps,
            "WeightedCumulativeMovingAverage": self._weighted_cumulative_moving_average_steps
        }
        # Aggregated model export type functions
        self.aggregator_export_functions = {
            "checkpoint": self._output_checkpoint,
            "saved_model": self._output_saved_model,
            "bundle": self._output_bundle,
        }

        # Get canonical checkpoint and set canonical variables
        self._get_cannonical()

        if aggregator_type not in self.aggregator_step_functions.keys():
            raise AggregatorNotFoundException("Aggregator of type \"{0}\" not found.".format(aggregator_type))
        else:
            self.aggregator_step_fn = self.aggregator_step_functions[aggregator_type]

        if export_type not in self.aggregator_export_functions.keys():
            raise AggregatorNotFoundException("Aggregator export type \"{0}\" not found.".format(export_type))
        else:
            self.export_fn = self.aggregator_export_functions[export_type]

        with tf.gfile.Open(schema_path, 'r') as schema_fp:
            self.schema = json.load(schema_fp)


    def aggregate(self, bundle_paths, output_path):
        update_var_values, update_var_dtypes = self._aggregate_ckpts(bundle_paths)
        self._apply_aggregation(output_path, update_var_values, update_var_dtypes)


    @staticmethod
    def validate_json_schema(result, schema):
        try:
            jsonschema.validate(instance=result, schema=schema)
        except Exception as e:
            raise SchemaFailedException(str(e))


    @staticmethod
    def _cumulative_moving_average_steps(steps):
        return 1


    @staticmethod
    def _weighted_cumulative_moving_average_steps(steps):
        return steps


    @staticmethod
    def _update_function(current, next_, current_steps, steps):
        if current_steps == 0:
            return next_
        update = ((current * current_steps) + (next_ * steps)) / (current_steps + steps)
        return update


    def _get_cannonical(self):
        # Initialize: Read canonical json, saved_model, respective variables names/shape, set values to zero
        # Opening in temporary directory, and storing relevant files in memory
        # This way we only need to get canonical files once instead of twice,
        # without changing class structure
        self.canonical_var_values, self.canonical_var_dtypes = {}, {}
        self.canonical_var_set = set([])

        with tempfile.TemporaryDirectory(dir=self.temp_dir) as canonical_temp:
            self.canonical_bundle_temppath = self._get_cannonical_bunde(canonical_temp)
            model_json_temppath = os.path.join(self.canonical_bundle_temppath, MODEL_JSON_BASENAME)

            # TODO: validate model.json
            with tf.gfile.Open(model_json_temppath, 'r') as model_json_input:
                self.model_json = json.load(model_json_input)
                model_json_spec = self.model_json.get('model', {})
                self.model_json_mode = model_json_spec.get('file')

            saved_model_pb_temppath = os.path.join(self.canonical_bundle_temppath, self.model_json_mode, SAVED_MODEL_BASENAME)
            with tf.gfile.Open(saved_model_pb_temppath, 'rb') as saved_model_fp:
                self.saved_model_binary = saved_model_fp.read()

            canonical_variables_temppath = os.path.join(self.canonical_bundle_temppath, self.model_json_mode, CHECKPOINT_PB_BASENAME)
            canonical_ckpt = ckpts.find_trained_checkpoints(canonical_variables_temppath)[0]
            # Load checkpoint
            canonical_reader = tf.contrib.framework.load_checkpoint(canonical_ckpt)
            self.canonical_var_list = tf.contrib.framework.list_variables(canonical_ckpt)
            # Initialize variables/weights
            for (canonical_name, canonical_shape) in self.canonical_var_list:
                self.canonical_var_values[canonical_name] = np.zeros(canonical_shape)
                canonical_tensor = canonical_reader.get_tensor(canonical_name)
                self.canonical_var_dtypes[canonical_name] = canonical_tensor.dtype
                self.canonical_var_set.add(canonical_name)
            self.logger.info("Initializing aggregated variables from: {0}".format(canonical_ckpt))


    def _get_cannonical_bunde(self, tempfile_temp_dir):
        try:
            url = "{0}{1}".format(self.repository, self.resource_path)
            headers = {
                "Authorization": "Bearer {0}".format(self.token)
            }
            response = requests.request("GET", url, data="", headers=headers)
            response_json = response.json()
            self.logger.info("Cannonical checkpoint JSON response {0}".format(response_json))
            url2 = response_json['link']
            bundle_download_filename = wget.download(url2, out=tempfile_temp_dir)
        except Exception as e:
            raise TensorIOModelsRepositoryException(
                "Failed to get tensorio-models cannonical bundle, error: {0}".format(str(e))
            )

        # Unzip bundle
        bundle_basename = os.path.basename(bundle_download_filename)
        with tf.gfile.Open(bundle_download_filename, 'rb') as bundle_fp:
            with zipfile.ZipFile(bundle_fp) as zf:
                zf.extractall(tempfile_temp_dir)
        os.remove(bundle_download_filename)

        bundle_dirname = ".".join(bundle_basename.split(".")[:-1])
        unzipped_bundle_dirs = [dir_name for dir_name in tf.gfile.ListDirectory(tempfile_temp_dir) if dir_name.split(".")[-1] != "zip"]
        assert len(unzipped_bundle_dirs) == 1
        bundle_path = os.path.join(tempfile_temp_dir, unzipped_bundle_dirs[0])
        return bundle_path


    def _aggregate_ckpts(self, bundle_paths):
        current_steps = 0
        update_var_values = copy.deepcopy(self.canonical_var_values)
        update_var_dtypes = copy.deepcopy(self.canonical_var_dtypes)

        # Load result/update checkpoints and aggregate
        for i, bundle_path in enumerate(bundle_paths):
            var_values, var_dtypes = {}, {}
            self.logger.info("Aggregating bundle: {0}".format(bundle_path))
            self.logger.info("Number of steps seen thus far: {0}".format(current_steps))
            with tempfile.TemporaryDirectory(dir=self.temp_dir) as temp:
                bundle_basename = os.path.basename(bundle_path)
                with tf.gfile.Open(bundle_path, 'rb') as bundle_fp:
                    with zipfile.ZipFile(bundle_fp) as zf:
                        zf.extractall(temp)
                        self.logger.info("Unzipping succesful to directory: {0}".format(temp))
                bundle_dirname = ".".join(bundle_basename.split(".")[:-1])
                model_directory_name = os.path.join(temp, bundle_dirname)
                checkpoint_dir_name = os.path.join(model_directory_name, BUNDLE_CHECKPOINT_DIRECTORY)
                current_ckpt = ckpts.find_trained_checkpoints(checkpoint_dir_name)[0]

                # Get result/update checkpoint
                result_filename = os.path.join(model_directory_name, RESULT_JSON_BASENAME)
                with tf.gfile.Open(result_filename, 'r') as result_fp:
                    results_dict = json.load(result_fp)
                    Aggregator.validate_json_schema(results_dict, self.schema)
                    checkpoint_steps = results_dict['numSamples']
                    self.logger.debug("Number of samples for update {0}: {1}".format(bundle_path, checkpoint_steps))
                    steps = self.aggregator_step_fn(checkpoint_steps)
                    self.logger.debug("Processed number of steps based on aggregation type: {0}".format(steps))

                # Load checkpoint
                reader = tf.contrib.framework.load_checkpoint(current_ckpt)
                var_list = tf.contrib.framework.list_variables(current_ckpt)
                var_set = set([name for (name, _ ) in var_list])
                var_dict = dict(var_list)

                # Verify set differences
                missing_canonical_vars = self.canonical_var_set - var_set
                missing_update_vars = var_set - self.canonical_var_set
                if len(missing_canonical_vars) > 0:
                    self.logger.debug("Canonical checkpoint variables NOT in Update: {0}".format(missing_canonical_vars))
                if len(missing_update_vars) > 0:
                    self.logger.debug("Update checkpoint variables NOT in Canonical: {0}".format(missing_update_vars))

                # Aggregate variables/weights
                for (name, shape) in self.canonical_var_list:
                    tensor = reader.get_tensor(name)
                    var_values[name] = tensor
                    var_dtypes[name] = tensor.dtype

                    # Do not update variables that are not in result/update checkpoint
                    if name not in var_values.keys() and name not in var_dtypes.keys():
                        continue
                    if shape != var_dict[name]:
                        self.logger.info("Canonical variable of shape: {0}, Update variable of shape: {1}".format(shape, var_list_dict[name]))
                    # Check dtype match
                    if var_dtypes[name] != self.canonical_var_dtypes[name]:
                        self.logger.info("Canonical variable of type: {0}, Update variable of type: {1}".format(canonical_var_dtypes[name], var_dtypes[name]))

                    update_var_values[name] = Aggregator._update_function(update_var_values[name], tensor, current_steps, steps)
                self.logger.info("Checkpoint aggregation successful for: {0}".format(current_ckpt))
            current_steps += steps

        return update_var_values, update_var_dtypes


    def _apply_aggregation(self, output_path, var_values, var_dtypes):
        # Get/Create original Tensorflow variables
        self.logger.info("Applying aggregation to export bundle...")
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
            # Export aggregation
            self.export_fn(saver, session, output_path)


    def _output_checkpoint(self, saver, session, output_path):
        ckpts.check_or_create_dir(output_path)
        aggregated_checkpoint_path = os.path.join(output_path, self.checkpoint_basename)
        save_path = saver.save(session, aggregated_checkpoint_path, write_meta_graph=True)
        self.logger.info("Averaged model chackpoints saved in: {0}".format(save_path))


    def _output_saved_model(self, saver, session, output_path):
        # Create necessary bundle structure for aggregation
        ckpts.check_or_create_dir(output_path)
        aggregated_checkpoint_directory = os.path.join(output_path, self.checkpoint_basename)
        ckpts.check_or_create_dir(aggregated_checkpoint_directory)
        # `aggregated_checkpoint_path` is not a directory, but basename for variables/ files
        aggregated_checkpoint_path = os.path.join(aggregated_checkpoint_directory, self.checkpoint_basename)

        # Unzip and copy necessary files
        with tempfile.TemporaryDirectory(dir=self.temp_dir) as temp:
            bundle_path = self._get_cannonical_bunde(temp)
            saved_model_pb_temppath = os.path.join(bundle_path, self.model_json_mode, SAVED_MODEL_BASENAME)
            saved_model_pb_newpath = os.path.join(output_path, SAVED_MODEL_BASENAME)
            with tf.gfile.Open(saved_model_pb_newpath, 'wb') as saved_model_output:
                saved_model_output.write(self.saved_model_binary)

        # No metagraph needed for pb exports
        save_path = saver.save(session, aggregated_checkpoint_path, write_meta_graph=False)
        self.logger.info("Averaged saved_model ProtoBuf saved in: {0}".format(save_path))


    def _output_bundle(self, saver, session, output_path):
        with tempfile.TemporaryDirectory(dir=self.temp_dir) as temp_bundle:
            # Create necessary bundle structure for aggregation
            ckpts.check_or_create_dir(temp_bundle)
            aggregated_pb_directory = os.path.join(temp_bundle, self.model_json_mode)
            ckpts.check_or_create_dir(aggregated_pb_directory)
            aggregated_checkpoint_directory = os.path.join(aggregated_pb_directory, self.checkpoint_basename)
            ckpts.check_or_create_dir(aggregated_checkpoint_directory)
            # `aggregated_checkpoint_path` is not a directory, but basename for variables/ files
            aggregated_checkpoint_path = os.path.join(aggregated_checkpoint_directory, self.checkpoint_basename)

            # No metagraph needed for pb exports
            save_path = saver.save(session, aggregated_checkpoint_path, write_meta_graph=False)
            self.logger.info("Averaged saved_model ProtoBuf saved in: {0}".format(save_path))

            # Unzip and copy necessary files
            tio_output_path = "{0}{1}".format(TIO_PREFIX, self.output_resource_path)
            saved_model_pb_newpath = os.path.join(aggregated_pb_directory, SAVED_MODEL_BASENAME)
            with tf.gfile.Open(saved_model_pb_newpath, 'wb') as saved_model_output:
                saved_model_output.write(self.saved_model_binary)

            model_json_newpath = os.path.join(temp_bundle, MODEL_JSON_BASENAME)
            self.model_json['id'] = tio_output_path
            with tf.gfile.Open(model_json_newpath, 'w') as model_json_output:
                json.dump(self.model_json, model_json_output, indent=4)

            # Create bundle
            bundle_basename = os.path.basename(self.canonical_bundle_temppath)
            bundle_name = ".".join(bundle_basename.split(".")[:-1])
            bundle_extension = bundle_basename.split(".")[-1]
            # example bundle name: manna-train-aggregated-1520170716.tiobundle.zip
            tiobundle_zip_name = "{0}-aggregated-{1}.{2}.zip".format(bundle_name, int(time.time()), bundle_extension)
            tiobundle_output_path = os.path.join(output_path, tiobundle_zip_name)
            bundle_output_path = bundler.tiobundle_build(
                aggregated_pb_directory, # bundler only needs saved_model directory
                model_json_newpath,
                None,
                bundle_basename, # unzipped bundle directory must match canonical
                tiobundle_output_path
            )
            self.logger.info("Output bundle stored: {0}".format(tiobundle_output_path))
            assert tiobundle_output_path == bundle_output_path

            # Register bundle
            registration = bundler.register_bundle(bundle_output_path, self.output_resource_path)
            self.logger.info('Bundle registered against repository: {}'.format(registration))



