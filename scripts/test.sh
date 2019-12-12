#!/bin/sh

set -e

echo "Running golint:"
golint  -set_exit_status ./...
echo "OK"

echo "Running go vet:"
go vet ./...
echo "OK"

echo "Running unit tests:"
go test ./... -coverprofile cover.out && go tool cover -func cover.out
sh ./scripts/coverage-check.sh cover.out
