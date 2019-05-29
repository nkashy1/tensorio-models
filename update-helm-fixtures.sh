#!/usr/bin/env sh

# Requires helm binary to be on PATH

# Change to script directory before running the tests
cd $(dirname $0)

TARGETS=$(ls -1 helm/values.*.yaml | xargs -L1 basename | sed 's/^values\.\(.*\)\.yaml/\1/')

for target in $TARGETS; do
    echo "Updating: helm/values.${target}.yaml -> helm/tests/fixtures/manifest.${target}.yaml"
    helm template -f helm/values.${target}.yaml helm/ >helm/tests/fixtures/manifest.${target}.yaml
done

