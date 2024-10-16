#!/bin/bash

# Set environment variables for cross-compilation
export GOARCH=amd64
export GOOS=linux

# Project directory (update this path to your project directory, or use current directory as default)
PROJECT_DIR=${1:-"."}

# Navigate to the project directory
echo "Navigating to the project directory: $PROJECT_DIR"
cd "$PROJECT_DIR" || { echo "Failed to navigate to the project directory"; exit 1; }

# Clean the Go cache and temporary files
echo "Cleaning Go cache..."
go clean -cache -modcache -i -r
if [ $? -ne 0 ]; then
    echo "Failed to clean Go cache"
    exit 1
fi

# Build the project
echo "Building the project..."
go build
if [ $? -ne 0 ]; then
    echo "Build failed!"
    exit 1
else
    echo "Build succeeded!"
fi

# Run the tests in verbose mode
echo "Running tests..."
go test -v ./...
if [ $? -ne 0 ]; then
    echo "Tests failed!"
    exit 1
else
    echo "All tests passed successfully!"
fi

# Clean up build artifacts (optional)
echo "Cleaning build artifacts..."
go clean
if [ $? -ne 0 ]; then
    echo "Failed to clean build artifacts"
    exit 1
else
    echo "Cleaned up build artifacts successfully!"
fi

# Done
echo "Script executed successfully!"
