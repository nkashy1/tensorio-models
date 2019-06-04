#/usr/bin/env sh

FLEA_API_URL=${FLEA_API_URL:-localhost:8083}
FLEA_API_TOKEN=${FLEA_API_TOKEN:-TaskGenToken}
MODEL=$1
HYPERPARAMETERS=$2
CHECKPOINT=$3
TASK=$4
LINK=$5

if [ -z $MODEL ]; then
    echo "ERROR: Please pass model name as the first argument to this script"
    exit 1
fi

if [ -z $HYPERPARAMETERS ]; then
    echo "ERROR: Please pass hyperparameters name as the second argument to this script"
    exit 1
fi

if [ -z $CHECKPOINT ]; then
    echo "ERROR: Please pass checkpoint name as the third argument to this script"
    exit 1
fi

if [ -z $TASK ]; then
    echo "ERROR: Please pass task name as the fourth argument to this script"
    exit 1
fi

if [ -z $LINK ]; then
    echo "ERROR: Please pass checkpoint link as the fifth argument to this script"
    exit 1
fi

set -e

echo "Creating task -- model: $MODEL, hyperparameters: $HYPERPARAMETERS, checkpoint: $CHECKPOINT, task: $TASK, link: $LINK"
curl -s -X POST -H "Authorization: Bearer $FLEA_API_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"modelId\": \"$MODEL\", \"hyperparametersId\": \"$HYPERPARAMETERS\", \"checkpointId\": \"$CHECKPOINT\", \"taskId\": \"$TASK\", \"deadline\": \"2019-07-31T23:59:59.002Z\", \"active\": true, \"link\": \"$LINK\"}" \
    $FLEA_API_URL/v1/flea/create_task
