APP_NAME := mujamalat
OUTPUT_DIR := build

BUILD_TIME := $(shell date +%s)
GIT_COMMIT := $(shell git rev-parse --short HEAD)

# ldflags for embedding variables
LDFLAGS := -ldflags "-X 'main.buildTime=$(BUILD_TIME)' -X 'main.gitCommit=$(GIT_COMMIT)' -s -w"
Tags := -tags 'static netgo'
TagsD := -tags 'netgo'

all: linux

linux:
	@echo "Building for linux amd64"
	@env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(Tags) $(LDFLAGS) -o $(OUTPUT_DIR)/$(APP_NAME) .

linux_dynamic:
	@echo "Building for linux amd64"
	@env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(TagsD) $(LDFLAGS) -o $(OUTPUT_DIR)/$(APP_NAME) .



release: clean format vet release_linux release_mac release_win

release_linux:
	@echo "[+] Building the Linux x86_64 version"
	@env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(Tags) $(LDFLAGS) -o $(OUTPUT_DIR)/$(APP_NAME) .
	@echo "[+] Packaging the Linux x86_64 version"
	@tar -czvf $(OUTPUT_DIR)/mujamalat_linux_x86_64.tar.gz -C $(OUTPUT_DIR) mujamalat > /dev/null
	@rm $(OUTPUT_DIR)/$(APP_NAME)

	@echo "[+] Building the Linux arm64 version"
	@env GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build $(Tags) $(LDFLAGS) -o $(OUTPUT_DIR)/$(APP_NAME) .
	@echo "[+] Packaging the Linux arm64 version"
	@tar -czvf $(OUTPUT_DIR)/mujamalat_linux_arm64.tar.gz -C $(OUTPUT_DIR) mujamalat > /dev/null
	@rm $(OUTPUT_DIR)/$(APP_NAME)

release_mac:
	@echo "[+] Building the macos x86_64 version"
	@env GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build $(Tags) $(LDFLAGS) -o $(OUTPUT_DIR)/$(APP_NAME) .
	@echo "[+] Packaging the macos x86_64 version"
	@tar -czvf $(OUTPUT_DIR)/mujamalat_macos_x86_64.tar.gz -C $(OUTPUT_DIR) mujamalat > /dev/null
	@rm $(OUTPUT_DIR)/$(APP_NAME)

	@echo "[+] Building the macos arm64 version"
	@env GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build $(Tags) $(LDFLAGS) -o $(OUTPUT_DIR)/$(APP_NAME) .
	@echo "[+] Packaging the macos arm64 version"
	@tar -czvf $(OUTPUT_DIR)/mujamalat_macos_arm64.tar.gz -C $(OUTPUT_DIR) mujamalat > /dev/null
	@rm $(OUTPUT_DIR)/$(APP_NAME)

release_win:
	@echo "[+] Building the windows x86_64 version"
	@env GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build $(Tags) $(LDFLAGS) -o $(OUTPUT_DIR)/$(APP_NAME).exe .
	@echo "[+] Packaging the windows x86_64 version"
	@zip -j $(OUTPUT_DIR)/mujamalat_windows_x86_64.zip $(OUTPUT_DIR)/mujamalat.exe > /dev/null
	@rm $(OUTPUT_DIR)/$(APP_NAME).exe

	@echo "[+] Building the windows arm64 version"
	@env GOOS=windows GOARCH=arm64 CGO_ENABLED=0 go build $(Tags) $(LDFLAGS) -o $(OUTPUT_DIR)/$(APP_NAME).exe .
	@echo "[+] Packaging the windows arm64 version"
	@zip -j $(OUTPUT_DIR)/mujamalat_windows_arm64.zip $(OUTPUT_DIR)/mujamalat.exe > /dev/null
	@rm $(OUTPUT_DIR)/$(APP_NAME).exe

clean:
	@rm -rf $(OUTPUT_DIR)/

update:
	@echo "[+] Updating Go dependencies"
	@go get -u
	@echo "[+] Done"

format:
	@echo "[+] Formatting Go files"
	@gofmt -w *.go

vet:
	@echo "[+] Running Go vet"
	@go vet

tidyup:
	@echo "[+] Running go mod tidy"
	@go get -u ./...
	@go mod tidy

