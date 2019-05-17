# imports?



def _aggregate_cumulative_moving_average(current, next, ckpt_meta):
    index = ckpt_meta['index'] # n = index in CWA formula
    update = (next - current) / float(index + 1)
    return current + update


def _aggregate_weighted_cumulative_moving_average(current, next, ckpt_meta):
    steps = ckpt_meta['steps']
    cumulative_steps = ckpt_meta['cumulative_steps']
    update = (next - current) / float(cumulative_steps + steps)
    return current + update