import aggregator
import argparse
import pandas as pd
import tensorflow as tf



def create_aggregator_argument_parser():
    parser = argparse.ArgumentParser(description='Checkpoint Aggregator')
    parser.add_argument(
        '--aggregation-type',
        type=str,
        required=True,
        choices=['CumulativeMovingAverage', 'WeightedCumulativeMovingAverage'], # TODO: read from aggregator function dict
        help=(
            'Type of running aggregation to be performed.',
            '   CumulativeMovingAverage',
            '   WeightedCumulativeMovingAverage'
        )
    )
    parser.add_argument(
        '--ckpt-paths-file',
        type=str,
        required=True,
        help='Single column csv/tsv/txt containing checkpoint directory paths to aggregate.'
    )
    parser.add_argument(
        '--resource-path',
        type=str,
        required=True,
        help='Resouce path to be used to fetch the right model metadata from FLEA and tensorio-models.'
    )
    parser.add_argument(
        '--output-path',
        type=str,
        required=True,
        help='Output (GCS) path to save aggregated model.'
    )
    parser.add_argument(
        '--output-resource-path',
        type=str,
        required=True,
        help='TensorIO output resource path, should mostly match resource path.'
    )
    parser.add_argument(
        '--export-type',
        type=str,
        required=True,
        choices=['checkpoint', 'saved_model', 'bundle'],
        help='Export output type: (ckpt, pb, bundle)'
    )
    parser.add_argument(
        '--repository',
        type=str,
        required=True,
        help='REPOSITORY to query base model from tensorio-models repository'
    )
    parser.add_argument(
        '--token',
        type=str,
        required=True,
        help='Tensorio-models repository Authorization token.'
    )
    parser.add_argument(
        '--debug',
        type=bool,
        default=False,
        help='Set to True for debugging.'
    )
    return parser


def parse_ckpt_paths_file(ckpt_filelist):
    with tf.gfile.Open(ckpt_filelist, 'r') as ckpt_filelist_fp:
        df = pd.read_csv(ckpt_filelist_fp, header=None, index_col=False)
        return df[0].tolist()


if __name__ == '__main__':
    aggregator_argument_parser = create_aggregator_argument_parser()
    aggregator_args = aggregator_argument_parser.parse_args()
    ckpt_filelist = parse_ckpt_paths_file(aggregator_args.ckpt_paths_file)
    agrgtr = aggregator.Aggregator(
        aggregator_args.resource_path,
        aggregator_args.output_resource_path,
        aggregator_args.repository,
        aggregator_args.token,
        aggregator_args.export_type,
        aggregator_args.aggregation_type
    )
    agrgtr.aggregate(ckpt_filelist, aggregator_args.output_path)

