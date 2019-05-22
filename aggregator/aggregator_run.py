import aggregator
import argparse
import pandas as pd



def create_aggregator_argument_parser():
    parser = argparse.ArgumentParser(description='Checkpoint Aggregator')
    parser.add_argument(
        '--aggregation-type',
        type=str,
        required=True,
        choices=list(aggregator.AGGREGATOR_FUNCTIONS.keys()),
        help=(
            'Type of running aggregation to be performed.',
            '   cma: cumulative moving average',
            '   wcma: weighted cumulative moving average'
        )
    )
    parser.add_argument(
        '--ckpt-paths-file',
        type=str,
        required=True,
        help='Single column csv/tsv/txt containing checkpoint directory paths to aggregate.'
    )
    parser.add_argument(
        '--output-path',
        type=str,
        required=True,
        help='Output (GCS) path to save aggregated model.'
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


# python aggregator_run.py \
#   --aggregation-type cma \
#   --ckpt-paths-file ./sample-ckpts-filelist.txt \
#   --output-path gs://doc-ai-models/aggregator-tests/aggregated-cma-test
if __name__ == '__main__':
    aggregator_argument_parser = create_aggregator_argument_parser()
    aggregator_args = aggregator_argument_parser.parse_args()
    ckpt_filelist = parse_ckpt_paths_file(aggregator_args.ckpt_paths_file)
    print (ckpt_filelist)
    agrgtr = aggregator.Aggregator(aggregator_args.aggregation_type, debug=aggregator_args.debug)
    agrgtr.aggregate(ckpt_filelist, aggregator_args.output_path)