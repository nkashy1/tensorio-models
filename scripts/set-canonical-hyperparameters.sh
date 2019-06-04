#/usr/bin/env sh

REPOSITORY_API_URL=${REPOSITORY_API_URL:-localhost:8081}
REPOSITORY_API_TOKEN=${REPOSITORY_API_TOKEN:-WriterToken}
MODEL=$1
CANONICAL_HYPERPARAMETERS=$2

if [ -z $MODEL ]; then
    echo "ERROR: Please pass model name as the first argument to this script"
    exit 1
fi

if [ -z $CANONICAL_HYPERPARAMETERS ]; then
    echo "ERROR: Please pass canonical hyperparameters name as the second argument to this script"
    exit 1
fi

set -e

echo "Setting canonical hyperparameters -- model: $MODEL, hyperparameters: $CANONICAL_HYPERPARAMETERS"
curl -X PUT \
    -H "Authorization: Bearer $REPOSITORY_API_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"model\": {\"canonicalHyperparameters\": \"$CANONICAL_HYPERPARAMETERS\"}}" \
    $REPOSITORY_API_URL/v1/repository/models/$MODEL

