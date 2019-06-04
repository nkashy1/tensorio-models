#/usr/bin/env sh

REPOSITORY_API_URL=${REPOSITORY_API_URL:-localhost:8081}
REPOSITORY_API_TOKEN=${REPOSITORY_API_TOKEN:-WriterToken}
MODEL=$1
HYPERPARAMETERS=$2

if [ -z $MODEL ]; then
    echo "ERROR: Please pass model name as the first argument to this script"
    exit 1
fi

if [ -z $HYPERPARAMETERS ]; then
    echo "ERROR: Please pass hyperparameters name as the second argument to this script"
    exit 1
fi

set -e

echo "Creating hyperparameters -- model: $MODEL, hyperparameters: $HYPERPARAMETERS"
curl -X POST \
    -H "Authorization: Bearer $REPOSITORY_API_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"hyperparametersId\": \"$HYPERPARAMETERS\"}" \
    $REPOSITORY_API_URL/v1/repository/models/$MODEL/hyperparameters
