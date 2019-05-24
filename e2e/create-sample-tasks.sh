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
    -d "{\"modelId\": \"Model17\", \"hyperparametersId\": \"hpset5\", \"checkpointId\": \"ckpt-7\", \"taskId\": \"task101\", \"deadline\": \"2019-12-31T23:59:59Z\", \"active\": false, \"link\": \"http://goo.gl/T1.zip\"}" \
    $FLEA_URL/create_task

echo ""
echo "Create Task: 102"
curl -X POST \
    -H "Content-Type: application/json" \
    -d "{\"modelId\": \"Model3\", \"hyperparametersId\": \"hpset3\", \"checkpointId\": \"ckpt-3\", \"taskId\": \"task102\", \"deadline\": \"2019-07-31T23:59:59Z\", \"active\": true, \"link\": \"http://goo.gl/T2.zip\"}" \
    $FLEA_URL/create_task

echo ""
echo "Create Task: aaa"
curl -X POST \
    -H "Content-Type: application/json" \
    -d "{\"modelId\": \"Model5\", \"hyperparametersId\": \"hpset13\", \"checkpointId\": \"ckpt-2\", \"taskId\": \"aaa\", \"deadline\": \"2019-12-14T23:59:59Z\", \"active\": true, \"link\": \"http://goo.gl/T3.zip\"}" \
    $FLEA_URL/create_task


echo ""
echo "List all tasks:"
curl "$FLEA_URL/tasks?includeInactive=true"

echo ""
echo "List all ACTIVE tasks:"
curl $FLEA_URL/tasks

echo ""
echo "Create Task: x3"
curl -X POST \
    -H "Content-Type: application/json" \
    -d "{\"modelId\": \"Model5\", \"hyperparametersId\": \"hpset3\", \"checkpointId\": \"ckpt-5\", \"taskId\": \"x3\", \"deadline\": \"2023-12-31T23:59:00Z\", \"active\": true, \"link\": \"http://goo.gl/Tx3.zip\"}" \
    $FLEA_URL/create_task

echo ""
echo "Create Task: y5"
curl -X POST \
    -H "Content-Type: application/json" \
    -d "{\"modelId\": \"Model15\", \"hyperparametersId\": \"hpset1\", \"checkpointId\": \"ckpt-42\", \"taskId\": \"y5\", \"deadline\": \"2019-12-21T23:59:59Z\", \"active\": true, \"link\": \"http://goo.gl/Tx5.zip\"}" \
    $FLEA_URL/create_task

echo ""
echo "Create Task: b7"
curl -X POST \
    -H "Content-Type: application/json" \
    -d "{\"modelId\": \"Model-bozo5\", \"hyperparametersId\": \"hpset1\", \"checkpointId\": \"ckpt-52\", \"taskId\": \"b7\", \"deadline\": \"2029-12-31T23:59:59Z\", \"active\": true, \"link\": \"http://goo.gl/Tzxe3.zip\"}" \
    $FLEA_URL/create_task


echo ""
echo "List first 3 tasks:"
curl "$FLEA_URL/tasks?inludeInactive=true&maxItems=3"

echo ""
echo "List tasks starting from b7:"
curl "$FLEA_URL/tasks?includeInactive=true&startTaskId=b7"

echo ""
echo "List 2 tasks starting from task101:"
curl "$FLEA_URL/tasks?includeInactive=true&startTaskId=task101&maxItems=2"

echo ""
echo "Get details for task x3:"
curl $FLEA_URL/tasks/x3

echo ""
echo "Modify task x3:"
curl -X POST \
    -H "Content-Type: application/json" \
    -d "{\"deadline\": \"2020-02-12T23:00:00Z\", \"active\": false}" \
    $FLEA_URL/modify_task/x3

echo ""
echo "Get details for task x3: (NOTE: Missing active => inactive)"
curl $FLEA_URL/tasks/x3

echo ""
echo "Start task b7:"
curl $FLEA_URL/start_task/b7

echo ""
echo "Start task b7 (emulate another client):"
curl $FLEA_URL/start_task/b7

