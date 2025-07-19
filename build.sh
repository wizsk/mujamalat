#!/bin/sh

set -ex

# -ldflags="-X 'main.compiledDate=$(date +%s)' -X 'main.gitCommit=$(git rev-parse --short HEAD)'" -tags 'static'
GOARCH=arm64 GOOS=linux CGO_ENABLED=0 go build -tags static -ldflags "-s -w" -o build/
