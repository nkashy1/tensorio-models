# Evaluator

## Overview

A Python (3.6.8) TensorFlow Checkpoint evaluator.

## Environment setup

Create a environment variables `.env` as provided in `sample.env`, or just
copy and fill in the correct values. The environment variables are defined as
follows:

1. `GOOGLE_APPLICATION_CREDENTIALS` - Google Cloud service account credential file json.
1. `EVALUATION_CKPT_PATH` - Path to a local or GCS checkpoint to be evaluated
1. `EVALUATION_0CLASS_NPY` - Path to a local or GCS numpy.save() file with array of pre-processed images for 0-class.
1. `EVALUATION_1CLASS_NPY` - Path to a local of GCS numpy.save() file with array of pre-processed images for 1-class.
1. `EVALUATION_OUTPUT_JSON` - Output (GCS) path to save evaluation metrics

## Run Container

To run container, simply run from root-dir:

```
make run-eval
```

Or to run locally:
```
run.sh
```

Then to see output:
```
gsutil cat $EVALUATION_OUTPUT_JSON
```