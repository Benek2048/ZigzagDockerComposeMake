all: test build build-linux build-rpi

test:
	go test -v ./... -count=1
build:
	go build -o bin/dcm.exe -ldflags "-s -w" github.com/Benek2048/ZigzagDockerComposeMake
build-linux:
	env GOOS='linux' go build -o bin/dcm -ldflags "-s -w" github.com/Benek2048/ZigzagDockerComposeMake
build-rpi:
	env GOOS='linux' GOARCH='arm' GOARM='7' go build -o bin/dcm-rpi -ldflags "-s -w" -trimpath github.com/Benek2048/ZigzagDockerComposeMake
install:
	go install github.com/Benek2048/ZigzagDockerComposeMake