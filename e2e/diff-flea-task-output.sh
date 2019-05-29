#/usr/bin/env sh

# Please, run this form repo root, with a fresh instance of 'make run-flea' running locally.
e2e/create-sample-tasks.sh | grep '{' > /tmp/task.output
# The only differences should be in 2 lines due to different random job IDs.
diff_count=$(diff /tmp/task.output common/fixtures/flea-create-tasks-expected-output.txt | wc)
if [ "$diff_count" == "       6      10     640" ]; then
    echo "FLEA E2E Test Passed"
    exit
else
    diff /tmp/task.output common/fixtures/flea-create-tasks-expected-output.txt
    echo
    echo "### Possible reason for diff is: 'make run-flea' is not fresh"
    echo 
    echo "FLEA E2E Test Failed"
    exit 1
fi
