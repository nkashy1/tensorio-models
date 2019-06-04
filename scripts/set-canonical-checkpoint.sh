#/usr/bin/env sh

REPOSITORY_API_URL=${REPOSITORY_API_URL:-localhost:8081}
REPOSITORY_API_TOKEN=${REPOSITORY_API_TOKEN:-WriterToken}
MODEL=$1
HYPERPARAMETERS=$2
CANONICAL_CHECKPOINT=$3

if [ -z $MODEL ]; then
    echo "ERROR: Please pass model name as the first argument to this script"
    exit 1
fi

if [ -z $HYPERPARAMETERS ]; then
    echo "ERROR: Please pass hyperparameters name as the second argument to this script"
    exit 1
fi

if [ -z $CANONICAL_CHECKPOINT ]; then
    echo "ERROR: Please pass canonical checkpoint name as the third argument to this script"
    exit 1
fi

set -e

echo "Setting canonical checkpoint -- model: $MODEL, hyperparameters: $HYPERPARAMETERS, checkpoint: $CANONICAL_CHECKPOINT"
curl -X PUT \
    -H "Authorization: Bearer $REPOSITORY_API_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"canonicalCheckpoint\": \"$CANONICAL_CHECKPOINT\"}}" \
    $REPOSITORY_API_URL/v1/repository/models/$MODEL/hyperparameters/$HYPERPARAMETERS

