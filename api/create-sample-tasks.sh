#/usr/bin/env sh

# Testing against local backend requires the flea server running in another window:
#   make run-flea
# Then simply type:
#   api/create-sample-tasks.sh
# One can also provide a server:port as argument.

API_URL=${1:-localhost:8083}
FLEA_URL=http://$API_URL/v1/flea

TIMESTAMP=$(date -u +%s)
MODEL="TestModel-$TIMESTAMP"


echo "Setting up API instance at: $FLEA_URL"

echo ""
echo "Health:" 
curl -X GET $FLEA_URL/healthz

echo ""
echo "Config:"
curl -X GET $FLEA_URL/config

echo ""
echo "Create Task: 101"
curl -X POST \
    -H "Content-Type: application/json" \
    -d "{\"modelId\": \"Model17\", \"hyperparametersId\": \"hpset5\", \"checkpointId\": \"ckpt-7\", \"taskId\": \"task101\", \"deadline\": 1558047280, \"active\": false, \"taskSpec\": \"http://goo.gl/T1.zip\"}" \
    $FLEA_URL/create_task

echo ""
echo "Create Task: 102"
curl -X POST \
    -H "Content-Type: application/json" \
    -d "{\"modelId\": \"Model3\", \"hyperparametersId\": \"hpset3\", \"checkpointId\": \"ckpt-3\", \"taskId\": \"task102\", \"deadline\": 1558087280, \"active\": true, \"taskSpec\": \"http://goo.gl/T2.zip\"}" \
    $FLEA_URL/create_task

echo ""
echo "Create Task: aaa"
curl -X POST \
    -H "Content-Type: application/json" \
    -d "{\"modelId\": \"Model5\", \"hyperparametersId\": \"hpset13\", \"checkpointId\": \"ckpt-2\", \"taskId\": \"aaa\", \"deadline\": 1553087280, \"active\": true, \"taskSpec\": \"http://goo.gl/T3.zip\"}" \
    $FLEA_URL/create_task



echo ""
echo "List all tasks:"
curl -X POST \
    -H "Content-Type: application/json" \
    -d "{\"includeInactive\": true}" \
    $FLEA_URL/tasks

echo ""
echo "List all ACTIVE tasks:"
curl -X POST \
     -H "Content-Type: application/json" \
     $FLEA_URL/tasks
