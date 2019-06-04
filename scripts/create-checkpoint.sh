#/usr/bin/env sh

REPOSITORY_API_URL=${REPOSITORY_API_URL:-localhost:8081}
REPOSITORY_API_TOKEN=${REPOSITORY_API_TOKEN:-WriterToken}
MODEL=$1
HYPERPARAMETERS=$2
CHECKPOINT=$3
LINK=$4

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

if [ -z $LINK ]; then
    echo "ERROR: Please pass checkpoint link as the fourth argument to this script"
    exit 1
fi

set -e

echo "Creating checkpoint -- model: $MODEL, hyperparameters: $HYPERPARAMETERS, checkpoint: $CHECKPOINT, link: $LINK"
curl -X POST \
    -H "Authorization: Bearer $REPOSITORY_API_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"checkpointId\": \"$CHECKPOINT\", \"link\": \"$LINK\"}" \
    $REPOSITORY_API_URL/v1/repository/models/$MODEL/hyperparameters/$HYPERPARAMETERS/checkpoints

