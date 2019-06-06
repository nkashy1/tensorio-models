import evaluator
import argparse
import numpy as np
import json
import os
import tensorflow as tf


def create_evaluator_argument_parser():
    parser = argparse.ArgumentParser(description='Checkpoint Evaluator')
    parser.add_argument(
        '--ckpt-path',
        type=str,
        required=True,
        help='Input path to checkpoint to be evaluated.'
    )
    parser.add_argument(
        '--npy-file-0class',
        type=str,
        required=True,
        help='Input path to NPY file of shape [#samples, width, height, 3] for 0-class'
    )
    parser.add_argument(
        '--npy-file-1class',
        type=str,
        required=True,
        help='Input path to NPY file of shape [#samples, width, height, 3] for 1-class'
    )
    parser.add_argument(
        '--output-file',
        type=str,
        required=True,
        help='Output path to store evaluation metrics.'
    )

    # Optional parameters - may be needed for training, etc, models

    parser.add_argument(
        '--debug',
        type=bool,
        default=False,
        help='Set to True for debugging.'
    )
    parser.add_argument(
        '--tag-set',
        type=str,
        default='serve',
        help="Typically 'serve' or 'train'."
    )
    parser.add_argument(
        '--signature-def-key',
        type=str,
        default='serving_default',
        help="Typically 'serving_default' or 'train'."
    )
    parser.add_argument(
        '--input-tensor-label',
        type=str,
        default='input',
        help="Typically 'input' - run with debug=True to see."
    )
    parser.add_argument(
        '--output-tensor-label',
        type=str,
        default='output',
        help="Typically 'outout' or 'predictions' - run with debug=True to see."
    )
    parser.add_argument(
        '--eval-threshold',
        type=float,
        default=0.5,
        help="Boundary between 0-class and 1-class"
    )
    return parser


if __name__ == '__main__':
    args = create_evaluator_argument_parser().parse_args()
    eval = evaluator.Evaluator(
        debug=args.debug, checkpoint_basename=args.ckpt_path)
    eval.tag_set = args.tag_set
    eval.signature_def_key = args.signature_def_key
    eval.threshold = args.eval_threshold
    print("Loading 0-class file...")
    with tf.gfile.GFile(args.npy_file_0class, 'rb') as f:
        zero_nparr = np.load(f)
    print("Loading 1-class file...")
    with tf.gfile.GFile(args.npy_file_1class, 'rb') as f:
        one_nparr = np.load(f)
    stats = eval.evaluate(zero_nparr, one_nparr,
                          args.input_tensor_label, args.output_tensor_label)
    json_stats = json.dumps(stats, indent=2, sort_keys=True)
    print(json_stats)
    with tf.gfile.GFile(args.output_file, 'w') as f:
        f.write(json_stats)
