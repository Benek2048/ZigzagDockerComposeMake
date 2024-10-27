# Makefile for building and testing the ZigzagDockerComposeMake project

# Default target that runs all the main tasks
all: test build build-linux build-rpi install

# Refresh Go module dependencies
refresh:
	go mod tidy -v

# Run tests with verbose output and without caching
test:
	go test -v ./... -count=1

# Build the project for Windows and output the binary to the bin directory
build:
	go build -o bin/dcm.exe -ldflags "-s -w" github.com/Benek2048/ZigzagDockerComposeMake

# Build the project for Linux and output the binary to the bin directory
build-linux:
	env GOOS='linux' go build -o bin/dcm -ldflags "-s -w" github.com/Benek2048/ZigzagDockerComposeMake

# Build the project for Raspberry Pi (ARM architecture) and output the binary to the bin directory
build-rpi:
	env GOOS='linux' GOARCH='arm' GOARM='7' go build -o bin/dcm-rpi -ldflags "-s -w" -trimpath github.com/Benek2048/ZigzagDockerComposeMake

# Install the project binary to the Go binary directory
install:
	#go install github.com/Benek2048/ZigzagDockerComposeMake
    # Using `go build` instead of `go install` to control the output name
	go build -o "$(shell go env GOBIN)/dcm$(shell go env GOEXE)" -ldflags "-s -w" github.com/Benek2048/ZigzagDockerComposeMake