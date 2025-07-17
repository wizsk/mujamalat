#!/bin/sh

set -ex

GOARCH=arm64 GOOS=linux CGO_ENABLED=0 go build -tags static -ldflags "-s -w" -o build/
