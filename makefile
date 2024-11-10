# Makefile for building and testing the ZigzagDockerComposeMake project

# Variables
BINARY_NAME=dcm
VERSION=$(shell git describe --tags --always --abbrev=0)
BUILD_TIME=$(shell date -u '+%Y-%m-%d %H:%M:%S')
GIT_COMMIT=$(shell git rev-parse HEAD)
LDFLAGS=-s -w \
  -X 'github.com/Benek2048/ZigzagDockerComposeMake/internal/logic.VersionConst=$(shell echo "$(VERSION)" | sed 's/^v//')' \
  -X 'github.com/Benek2048/ZigzagDockerComposeMake/internal/logic.BuildTime=$(BUILD_TIME)' \
  -X 'github.com/Benek2048/ZigzagDockerComposeMake/internal/logic.GitCommit=$(GIT_COMMIT)'

# Show variables
show_vars:
	@echo "VERSION    = $(VERSION)"
	@echo "BUILD_TIME = $(BUILD_TIME)"
	@echo "GIT_COMMIT = $(GIT_COMMIT)"
	@echo "LDFLAGS    = $(LDFLAGS)"
	@echo "SHELL      = $(SHELL)"

# Default target that runs all the main tasks
all: show_vars refresh test build build-linux build-rpi build-darwin install

# Update dependencies
update:
	go get -u ./...

# Refresh Go module dependencies
refresh:
	go mod tidy -v

# Run tests with verbose output and without caching
test:
	go test -v ./... -count=1

# Build the project for Windows
build:
	cp internal/assets/windows/rsrc_windows_amd64.syso ./
	go build -o bin/$(BINARY_NAME).exe -ldflags "$(LDFLAGS)" github.com/Benek2048/ZigzagDockerComposeMake
	rm rsrc_windows_amd64.syso

# Build the project for Linux
build-linux:
	GOOS=linux GOARCH=amd64 go build -o bin/$(BINARY_NAME) -ldflags "$(LDFLAGS)" github.com/Benek2048/ZigzagDockerComposeMake

# Build the project for Raspberry Pi (ARM architecture)
build-rpi:
	GOOS=linux GOARCH=arm GOARM=7 go build -o bin/$(BINARY_NAME)-rpi -ldflags "$(LDFLAGS)" -trimpath github.com/Benek2048/ZigzagDockerComposeMake

# Build for macOS (both Intel and Apple Silicon)
build-darwin:
	GOOS=darwin GOARCH=amd64 go build -o bin/$(BINARY_NAME)-darwin-amd64 -ldflags "$(LDFLAGS)" github.com/Benek2048/ZigzagDockerComposeMake
	GOOS=darwin GOARCH=arm64 go build -o bin/$(BINARY_NAME)-darwin-arm64 -ldflags "$(LDFLAGS)" github.com/Benek2048/ZigzagDockerComposeMake

# Install the project binary
install:
	go build -o "$(shell go env GOBIN)/$(BINARY_NAME)$(shell go env GOEXE)" -ldflags "$(LDFLAGS)" github.com/Benek2048/ZigzagDockerComposeMake

# Clean build artifacts
clean:
	rm -rf bin/
	go clean

# Create release archives
release:
	cd bin && \
	zip $(BINARY_NAME)-windows-amd64.zip $(BINARY_NAME).exe && \
	tar czf $(BINARY_NAME)-linux-amd64.tar.gz $(BINARY_NAME) && \
	tar czf $(BINARY_NAME)-linux-arm.tar.gz $(BINARY_NAME)-rpi && \
	tar czf $(BINARY_NAME)-darwin-amd64.tar.gz $(BINARY_NAME)-darwin-amd64 && \
	tar czf $(BINARY_NAME)-darwin-arm64.tar.gz $(BINARY_NAME)-darwin-arm64 && \
	sha256sum *.zip *.tar.gz > checksums.txt

.PHONY: all refresh test build build-linux build-rpi build-darwin install clean release