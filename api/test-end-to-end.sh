#!/bin/sh

# End-to-end tests for the tensorio-models API.
# Intended to be run against a fresh instance of the API which will be used ONLY for testing.
# For example, using the memory storage backend.
#
# Requirements:
# 1. curl
# 2. jq

print_usage() {
    echo "Usage: $0 [API_URL]"
    echo ""
    echo "API_URL"
    echo "\tURL for API instance to be tested end to end (default: http://localhost:8081)"
}

if [ "$1" = "-h" ] || [ "$1" = "--help" ]; then
    print_usage
    exit 2
fi

API_URL=${1:-"http://localhost:8081"}

set -u

echo "Testing: $API_URL"

# /healthz
echo "- /healthz..."
RESPONSE=$(curl -s -k -X GET "$API_URL/v1/repository/healthz")
RESPONSE_STATUS=$(echo $RESPONSE | jq '.status')
if [ -z "$RESPONSE_STATUS" ] || [ "$RESPONSE_STATUS" != '"SERVING"' ]; then
    echo "\tFailed"
    echo "\tExpected JSON object with \"status\" of \"SERVING\""
    echo "\tActual response body: $RESPONSE"
else
    echo "\tPassed"
fi
