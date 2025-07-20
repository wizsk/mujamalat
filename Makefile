APP_NAME := mujadalat
OUTPUT_DIR := build

BUILD_TIME := $(shell date +%s)
GIT_COMMIT := $(shell git rev-parse --short HEAD)

# ldflags for embedding variables
LDFLAGS := -ldflags "-X 'main.buildTime=$(BUILD_TIME)' -X 'main.gitCommit=$(GIT_COMMIT)'"
Tags := -tags 'static netgo'

# Default target
all: linux

# Build the application
linux:
	@echo "Building for linux amd64"
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(Tags) $(LDFLAGS) -o $(OUTPUT_DIR)/$(APP_NAME) .
