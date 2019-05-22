#!/usr/bin/env sh

# Requires helm binary to be on PATH

# Change to script directory before running the tests
cd $(dirname $0)

TARGETS=$(ls -1 helm/values.*.yaml | xargs -L1 basename | sed 's/^values\.\(.*\)\.yaml/\1/')

ERRORS=0

for target in $TARGETS; do
    echo "Testing: helm/values.${target}.yaml against helm/tests/fixtures/manifest.${target}.yaml"
    helm template -f helm/values.${target}.yaml helm/ | diff -u - helm/tests/fixtures/manifest.${target}.yaml
    if [ $? -ne 0 ]; then
        echo "    FAILED"
        ERRORS=$((ERRORS+1))
    else
        echo "    PASSED"
    fi
done

echo "Number of errors: $ERRORS"

exit $ERRORS
