#/usr/bin/env sh

# We have added basic token authentication. The client should provide a header like shown below.
# Note that tokens differ by type of user.
# Then, testing against local backend requires the flea server running in another window:
#   make run-flea
# Then simply type:
#   e2e/create-sample-tasks.sh
# One can also provide a server:port as argument.

API_URL=${1:-localhost:8083}
FLEA_URL=$API_URL/v1/flea
FLEA_ADMIN_TOKEN=${FLEA_ADMIN_TOKEN:-AdminToken}
FLEA_TASKGEN_TOKEN=${FLEA_TASKGEN_TOKEN:-TaskGenToken}
FLEA_CLIENT_TOKEN=${FLEA_CLIENT_TOKEN:-ClientToken}

TIMESTAMP=$(date -u +%s)
MODEL="TestModel-$TIMESTAMP"

echo "Setting up API instance at: $FLEA_URL"

echo ""
echo "Health:"
curl -s -X GET $FLEA_URL/healthz

echo ""
echo "Config:"
curl -s -X GET $FLEA_URL/config

echo ""
echo "Create Task: 101"
curl -s -X POST -H "Authorization: Bearer $FLEA_TASKGEN_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"modelId\": \"Model17\", \"hyperparametersId\": \"hpset5\", \"checkpointId\": \"ckpt-7\", \"taskId\": \"task101\", \"deadline\": \"2019-12-31T23:59:59.001Z\", \"active\": false, \"link\": \"http://goo.gl/T1.zip\"}" \
    $FLEA_URL/create_task

echo ""
echo "Create Task: 102"
curl -s -X POST -H "Authorization: Bearer $FLEA_TASKGEN_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"modelId\": \"Model3\", \"hyperparametersId\": \"hpset3\", \"checkpointId\": \"ckpt-3\", \"taskId\": \"task102\", \"deadline\": \"2019-07-31T23:59:59.002Z\", \"active\": true, \"link\": \"http://goo.gl/T2.zip\"}" \
    $FLEA_URL/create_task

echo ""
echo "Create Task: aaa"
curl -s -X POST -H "Authorization: Bearer $FLEA_TASKGEN_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"modelId\": \"Model5\", \"hyperparametersId\": \"hpset13\", \"checkpointId\": \"ckpt-2\", \"taskId\": \"aaa\", \"deadline\": \"2019-12-14T23:59:59.003Z\", \"active\": true, \"link\": \"http://goo.gl/T3.zip\"}" \
    $FLEA_URL/create_task


echo ""
echo "List all tasks:"
curl -H "Authorization: Bearer $FLEA_CLIENT_TOKEN" "$FLEA_URL/tasks?includeInactive=true"

echo ""
echo "List all ACTIVE tasks:"
curl -H "Authorization: Bearer $FLEA_CLIENT_TOKEN" $FLEA_URL/tasks

echo ""
echo "Create Task: x3"
curl -s -X POST -H "Authorization: Bearer $FLEA_TASKGEN_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"modelId\": \"Model5\", \"hyperparametersId\": \"hpset3\", \"checkpointId\": \"ckpt-5\", \"taskId\": \"x3\", \"deadline\": \"2023-12-31T23:59:00.004Z\", \"active\": true, \"link\": \"http://goo.gl/Tx3.zip\"}" \
    $FLEA_URL/create_task

echo ""
echo "Create Task: y5"
curl -s -X POST -H "Authorization: Bearer $FLEA_TASKGEN_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"modelId\": \"Model15\", \"hyperparametersId\": \"hpset1\", \"checkpointId\": \"ckpt-42\", \"taskId\": \"y5\", \"deadline\": \"2019-12-21T23:59:59.005Z\", \"active\": true, \"link\": \"http://goo.gl/Tx5.zip\"}" \
    $FLEA_URL/create_task

echo ""
echo "Create Task: b7"
curl -s -X POST -H "Authorization: Bearer $FLEA_TASKGEN_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"modelId\": \"Model-bozo5\", \"hyperparametersId\": \"hpset1\", \"checkpointId\": \"ckpt-52\", \"taskId\": \"b7\", \"deadline\": \"2029-12-31T23:59:59.006Z\", \"active\": true, \"link\": \"http://goo.gl/Tzxe3.zip\"}" \
    $FLEA_URL/create_task

echo ""
echo "List first 3 tasks:"
curl -H "Authorization: Bearer $FLEA_CLIENT_TOKEN" "$FLEA_URL/tasks?inludeInactive=true&maxItems=3"

echo ""
echo "List tasks starting from b7:"
curl -H "Authorization: Bearer $FLEA_CLIENT_TOKEN" "$FLEA_URL/tasks?includeInactive=true&startTaskId=b7"

echo ""
echo "List 2 tasks starting from task101:"
curl -H "Authorization: Bearer $FLEA_CLIENT_TOKEN" "$FLEA_URL/tasks?includeInactive=true&startTaskId=task101&maxItems=2"

echo ""
echo "Get details for task x3:"
curl -H "Authorization: Bearer $FLEA_CLIENT_TOKEN" $FLEA_URL/tasks/x3

echo ""
echo "Modify task x3:"
curl -s -X POST -H "Authorization: Bearer $FLEA_TASKGEN_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"deadline\": \"2020-02-12T23:00:00.888888Z\", \"active\": false}" \
    $FLEA_URL/modify_task/x3

echo ""
echo "Get details for task x3:"
curl -s  -H "Authorization: Bearer $FLEA_CLIENT_TOKEN" $FLEA_URL/tasks/x3

echo ""
echo "Reload credentials:"
curl -s  -X POST -H "Authorization: Bearer $FLEA_ADMIN_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"type\": 1}" \
    $FLEA_URL/admin

echo 
echo "Get details for task x3:"
curl -H "Authorization: Bearer $FLEA_CLIENT_TOKEN" $FLEA_URL/tasks/x3

echo ""
echo "Start task b7:"
curl -s  -H "Authorization: Bearer $FLEA_CLIENT_TOKEN" $FLEA_URL/start_task/b7

echo ""
echo "Start task b7 (emulate another client):"
curl -s  -H "Authorization: Bearer $FLEA_CLIENT_TOKEN" $FLEA_URL/start_task/b7

echo ""
echo "Unauthorized 1:"
curl -s  -H "Authorization: $FLEA_CLIENT_TOKEN" $FLEA_URL/start_task/b7

echo ""
echo "Unauthorized 2:"
curl -s  $FLEA_URL/start_task/b7

JOBID=$(curl -s  -H "Authorization: Bearer $FLEA_CLIENT_TOKEN" $FLEA_URL/start_task/x3 | jq .jobId | tr -d \")

echo
echo "Add Task Error Report"
curl -s -X POST -H "Authorization: Bearer $FLEA_CLIENT_TOKEN" \
     -H "Content-Type: application/json" \
     -d "{\"errorMessage\": \"Save this error!\"}" \
     $FLEA_URL/job_error/x3/$JOBID

echo
echo "Unauthenticated log request"
curl -s -X POST \
     -H "Content-Type: application/json" \
     -d "{\"message\": \"Ignore me \"}" \
     $FLEA_URL/log/client1

echo
echo "Valid log request"
curl -s -X POST -H "Authorization: Bearer $FLEA_CLIENT_TOKEN" \
     -H "Content-Type: application/json" \
     -d "{\"message\": \"Only besties make it into the log.\"}" \
     $FLEA_URL/log/BestClient
