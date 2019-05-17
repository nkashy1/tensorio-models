# aggregator

## Overview

A Python (3.6.8) TensorFlow Checkpoint aggregator that
is necessary for Federated Learning. Specifically, this container will take a
file containing a column of checkpoint paths and will dump the aggregated output
to the specified output path. Currently there are two forms of aggregation, namely,
cumulative running average (CWA) and weighted cumulative running average (WCMA).

## Environment setup

Create a environment variables `.env` as provided in `sample.env`, or just
copy and fill in the correct values. The environment variables are defined as
follows:

1. `GOOGLE_CREDENTIALS_FILE` - Google Cloud service account credential file
(json or P12) with the appropriate permissions to read/write to a project's
GCS buckets.
1. `AGGREGATION_CKPTS_FILELIST` - Path to a local or GCS single-column csv/tsv/txt
file containing a list of checkpoint base paths to be aggregated
1. `AGGREGATION_TYPE` - Type of running aggregation to be performed.
  1. `cma`: cumulative moving average
  1. `wcma`: weighted cumulative moving average
1. `AGGREGATION_OUTPUT_PATH` - Output (GCS) path to save aggregated model checkpoint

## Run Container

To run container, simply run:

```
docker-compose up --build
```

Then check the appropriate GCS bucket (corresponding to `AGGREGATION_OUTPUT_PATH`)
to see if checkpoints were aggregated correctly.