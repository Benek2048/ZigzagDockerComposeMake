all: build build-linux build-rpi

build:
	go build -o dcm.exe -ldflags "-s -w" github.com/Benek2048/ZigzagDockerComposeMake
build-linux:
	env GOOS='linux' go build -o dcm -ldflags "-s -w" github.com/Benek2048/ZigzagDockerComposeMake
build-rpi:
	env GOOS='linux' GOARCH='arm' GOARM='7' go build -o dcm-rpi -ldflags "-s -w" -trimpath github.com/Benek2048/ZigzagDockerComposeMake