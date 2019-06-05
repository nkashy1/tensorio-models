#/usr/bin/env sh

# To test a local deployment if you don't plan to download any dummy bundles:
# $ ./setup.sh
#
# To test a local deployment if you do plan to download dummy bundles hosted at, for example,
# localhost:8000:
# $ FILE_SERVER_URL=localhost:8000 ./setup.sh
#
# To test a remote deployment if you don't plan to download any dummy bundles:
# $ API_URL=<remote URL> ./setup.sh
#
# To test a remote deployment if you do plan to download dummy bundles hosted at, for example,
# localhost:8000:
# $ API_URL=<remote URL> FILE_SERVER_URL=localhost:8000 ./setup.sh

API_URL=${API_URL:-localhost:8081}
API_TOKEN=${API_TOKEN:-WriterToken}
FILE_SERVER_URL=${FILE_SERVER_URL:-http://example.com}

TIMESTAMP=$(date -u +%s)
MODEL="TestModel-$TIMESTAMP"

set -e

echo "Setting up API instance at: $API_URL"

echo "Creating model: $MODEL ..."
curl -X POST \
    -H "Authorization: Bearer $API_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"model\": {\"modelId\": \"$MODEL\", \"details\": \"This model is a test\", \"canonicalHyperparameters\": \"hyperparameters-2\"}}" \
    $API_URL/v1/repository/models

echo

echo "Creating hyperparameters-1..."
curl -X POST \
    -H "Authorization: Bearer $API_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"hyperparametersId": "hyperparameters-1", "hyperparameters": {"lol": "rofl"}}' \
    $API_URL/v1/repository/models/$MODEL/hyperparameters

echo

echo "Creating checkpoint-1 for hyperparameters-1..."
curl -X POST \
    -H "Authorization: Bearer $API_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"checkpointId\": \"checkpoint-1\", \"createdAt\": \"1557790163\", \"info\": {\"accuracy\": \"0.934\"}, \"link\": \"$FILE_SERVER_URL/h1c1.tiobundle.zip\"}" \
    $API_URL/v1/repository/models/$MODEL/hyperparameters/hyperparameters-1/checkpoints

echo

echo "Setting checkpoint-1 as canonicalCheckpoint for hyperparameters-1 and hyperparameters-2 as its upgrade..."
curl -X PUT \
    -H "Authorization: Bearer $API_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"canonicalCheckpoint": "checkpoint-1", "upgradeTo": "hyperparameters-2"}' \
    $API_URL/v1/repository/models/$MODEL/hyperparameters/hyperparameters-1

echo

echo "Creating hyperparameters-2"
curl -X POST \
    -H "Authorization: Bearer $API_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"hyperparametersId": "hyperparameters-2", "hyperparameters": {"lol": "wtf"}}' \
    $API_URL/v1/repository/models/$MODEL/hyperparameters

echo

echo "Creating checkpoint-1 for hyperparameters-2..."
curl -X POST \
    -H "Authorization: Bearer $API_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"checkpointId\": \"checkpoint-1\", \"createdAt\": \"1557790252\", \"info\": {\"accuracy\": \"0.921\"}, \"link\": \"$FILE_SERVER_URL/h2c1.tiobundle.zip\"}" \
    $API_URL/v1/repository/models/$MODEL/hyperparameters/hyperparameters-2/checkpoints

echo

echo "Creating checkpoint-2 for hyperparameters-2..."
curl -X POST \
    -H "Authorization: Bearer $API_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"checkpointId\": \"checkpoint-2\", \"createdAt\": \"1557790268\", \"info\": {\"accuracy\": \"0.959\"}, \"link\": \"$FILE_SERVER_URL/h2c2.tiobundle.zip\"}" \
    $API_URL/v1/repository/models/$MODEL/hyperparameters/hyperparameters-2/checkpoints

echo

echo "Setting checkpoint-2 as canonicalCheckpoint for hyperparameters-2..."
curl -X PUT \
    -H "Authorization: Bearer $API_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"canonicalCheckpoint": "checkpoint-2"}' \
    $API_URL/v1/repository/models/$MODEL/hyperparameters/hyperparameters-2

echo
