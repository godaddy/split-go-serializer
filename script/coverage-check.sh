#!/bin/sh

set -e

THRESHOLD='85.0'

COVERAGE=$(go tool cover -func $1 | grep 'total:' | awk '{ print(substr($3, 1, length($3)-1)) }')
PASSED=$(echo "${COVERAGE}>=${THRESHOLD}" | bc -l)

if [ $PASSED = 0 ]; then
    echo "Failed code coverage threshold check: ${COVERAGE} < ${THRESHOLD}"
    exit 1
fi

echo "Passed code coverage threshold check: ${COVERAGE} >= ${THRESHOLD}"
