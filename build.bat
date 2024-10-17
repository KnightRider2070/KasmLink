@echo off
REM Build for Windows
set GOOS=windows
set GOARCH=amd64
go build -o kasmlink.exe main.go

REM Build for Linux
set GOOS=linux
set GOARCH=amd64
go build -o kasmlink-linux main.go

echo Build completed.