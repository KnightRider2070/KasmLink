#!/bin/bash
# Build for Windows
export GOOS=windows
export GOARCH=amd64
go build -o kasmlink.exe main.go

# Build for Linux
export GOOS=linux
export GOARCH=amd64
go build -o kasmlink-linux main.go

echo "Build completed."