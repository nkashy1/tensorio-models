import os
import tensorflow as tf



def find_trained_checkpoints(model_directory_name):
    checkpoints = []
    for f in tf.gfile.ListDirectory(model_directory_name):
        if f.endswith(".index") and "initial" not in f: #  Find index files (jointly made with .data files)
            path = os.path.join(model_directory_name, f)
            checkpoints.append(path.split(".")[0]) # Remove extension
    return checkpoints if len(checkpoints) > 0 else None


def check_or_create_dir(output_path):
    if not tf.gfile.IsDirectory(output_path): # exists?
        tf.gfile.MkDir(output_path)


