#!/usr/bin/env sh

# Assumptions:
# This script should be run in a suitable Python environment
# If the environment variable TEST=true, then package is pushed to Test PyPI. Otherwise, it is
# pushed to PyPI.

set -ex

SCRIPT_DIR=$(dirname $0)

REPOSITORY_STRING=""
if [ "$TEST" = true ]; then
    REPOSITORY_STRING="--repository-url https://test.pypi.org/legacy/"
fi

python setup.py sdist bdist_wheel

twine upload $REPOSITORY_STRING dist/*
