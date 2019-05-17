from utils import aggregation_fn, ckpts


# NOTE: TensorIO outputs checkpoints, and can read in checkpoints, but depends on
#       the original .pb file for seved_model functionality (receiver fns, etc.)

# Cumulative Moving Average, Weighted Cumulative Moving Average
AGGREGATOR_FUNCTIONS = {
    "cma": aggregation_fn._aggregate_cumulative_moving_average,
    "wcma": aggregation_fn._aggregate_weighted_cumulative_moving_average
}

class AggregatorNotFoundException(Exception):
    pass

class Aggregator(object):

    def __init__(self, aggregator_type="cma", debug=False):
        self.num_aggregations = 0
        self.debug = debug
        if aggregator_type not in AGGREGATOR_FUNCTIONS.keys():
            raise AggregatorNotFoundException("Aggregator of type \"{0}\" not found.".format(aggregator_type))
        else:
            self.aggregator_fn = AGGREGATOR_FUNCTIONS[aggregator_type]


    def aggregate(self, model_directory_names, output_path):
        ckpt_merge_dict, total_steps = ckpts.get_ckpts(model_directory_names, debug=self.debug)
        var_values, var_dtypes = ckpts.aggregate_ckpts(ckpt_merge_dict, aggregation_function=self.aggregator_fn, debug=self.debug)
        ckpts.apply_aggregation(output_path, var_values, var_dtypes, total_steps, debug=self.debug)
        self.num_aggregations += 1




