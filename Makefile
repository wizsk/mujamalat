APP_NAME := mujadalat
OUTPUT_DIR := build

BUILD_TIME := $(shell date +%s)
GIT_COMMIT := $(shell git rev-parse --short HEAD)

# ldflags for embedding variables
LDFLAGS := -ldflags "-X 'main.buildTime=$(BUILD_TIME)' -X 'main.gitCommit=$(GIT_COMMIT)' -s -w"
Tags := -tags 'static netgo'
TagsD := -tags 'netgo'

# Default target
all: linux

# Build the application
linux:
	@echo "Building for linux amd64"
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(Tags) $(LDFLAGS) -o $(OUTPUT_DIR)/$(APP_NAME) .

linux_dynamic:
	@echo "Building for linux amd64"
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(TagsD) $(LDFLAGS) -o $(OUTPUT_DIR)/$(APP_NAME) .
