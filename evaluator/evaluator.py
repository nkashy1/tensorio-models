import os
from tensorflow.core.example import example_pb2
from tensorflow.core.framework import types_pb2
from tensorflow.python.client import session
from tensorflow.python.saved_model import loader
from tensorflow.python.tools import saved_model_utils
from tensorflow.python.tools import saved_model_cli
from tensorflow.python.framework import ops as ops_lib
import tensorflow as tf
import numpy as np


class EvaluatorNotFoundException(Exception):
    pass


class Evaluator(object):

    def __init__(self, checkpoint_basename, debug=False):
        self.debug = debug
        self.checkpoint_basename = checkpoint_basename
        self.saved_model_dir = checkpoint_basename
        self.meta_graph_depth = None
        self.input_tensor_info = None
        self.output_tensor_info = None
        self.tag_set = 'serve'
        self.signature_def_key = 'serving_default'
        self.threshold = 0.5
        self.zero_preds = []
        self.one_preds = []

    def populate_graph_and_tensor_maps(self):
        print("Loading graph metadata...")
        self.meta_graph_def = saved_model_utils.get_meta_graph_def(self.saved_model_dir,
                                                                   self.tag_set)
        self.input_tensor_info = saved_model_cli._get_inputs_tensor_info_from_meta_graph_def(
            self.meta_graph_def, self.signature_def_key)
        self.output_tensor_info = saved_model_cli._get_outputs_tensor_info_from_meta_graph_def(
            self.meta_graph_def, self.signature_def_key)

    def compute_stats(self, zero_preds, one_preds):
        results = {}
        zcorrect = 0
        sum_err = 0.0
        sum_err_sq = 0.0
        for zpred in zero_preds:
            sum_err += zpred[0]
            sum_err_sq += (zpred[0] * zpred[0])
            if zpred[0] < self.threshold:
                zcorrect += 1
        num_zeros = zero_preds.shape[0]
        num_ones = one_preds.shape[0]
        results["num_neg_samples"] = num_zeros
        results["num_pos_samples"] = num_ones
        results["true_neg_pct"] = 100.0 * zcorrect / num_zeros
        results["false_pos_pct"] = 100.0 * (num_zeros - zcorrect) / num_zeros
        ocorrect = 0
        for opred in one_preds:
            sum_err += (1 - opred[0])
            sum_err_sq += (1 - opred[0]) * (1 - opred[0])
            if opred[0] >= self.threshold:
                ocorrect += 1
        results["true_pos_pct"] = 100.0 * ocorrect / num_ones
        results["false_neg_pct"] = 100.0 * (num_ones - ocorrect) / num_ones
        results["mean_error"] = sum_err / (num_zeros + num_ones)
        results["mesn_error_sq"] = sum_err_sq / (num_zeros + num_ones)
        results["total_eval_samples"] = num_zeros + num_ones
        return results

    def evaluate(self, zero_class_nparr, one_class_nparr,
                 input_name='input', output_name='output'):
        self.populate_graph_and_tensor_maps()
        if self.debug:
            print("Input Tensor Map:", self.input_tensor_info)
            print("Output Tensor Map:", self.output_tensor_info)

        input_tensor_name = self.input_tensor_info[input_name].name
        if input_tensor_name is None:
            raise EvaluatorNotFoundException
        else:
            print("Mapped", input_name, "to", input_tensor_name)

        output_tensor_name = self.output_tensor_info[output_name].name
        if output_tensor_name is None:
            raise EvaluatorNotFoundException
        else:
            print("Mapped", output_name, "to", output_tensor_name)

        print("Loading model...")
        with session.Session(None, graph=ops_lib.Graph()) as sess:
            loader.load(sess, self.tag_set.split(','), self.saved_model_dir)

            print("Starting evaluation...")
            num_zero_records = zero_class_nparr.shape[0]
            feed_dict = {input_tensor_name: zero_class_nparr}
            if self.signature_def_key == "train":
                feed_dict["label:0"] = np.zeros(
                    [num_zero_records, 1], dtype=np.int32)
            self.zero_preds = sess.run(output_tensor_name, feed_dict=feed_dict)
            print("Evaluated", num_zero_records, "zero samples")

            num_one_records = one_class_nparr.shape[0]
            feed_dict = {input_tensor_name: one_class_nparr}
            if self.signature_def_key == "train":
                feed_dict["label:0"] = np.ones(
                    [num_one_records, 1], dtype=np.int32)
            self.one_preds = sess.run(output_tensor_name, feed_dict=feed_dict)
            print("Evaluated", num_one_records, "one samples")
        return self.compute_stats(self.zero_preds, self.one_preds)
