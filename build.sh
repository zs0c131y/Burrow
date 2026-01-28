#!/bin/bash

# Burrow Build Script for Unix systems (cross-compile to Windows)

set -e

echo "===================================="
echo "Building Burrow for Windows..."
echo "===================================="
echo

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "ERROR: Go is not installed"
    echo "Please install Go from https://golang.org/dl/"
    exit 1
fi

echo "Go version:"
go version
echo

# Clean previous builds
if [ -f wm.exe ]; then
    echo "Removing previous build..."
    rm wm.exe
fi

# Download dependencies
echo "Downloading dependencies..."
go mod download

# Build for Windows
echo "Building Burrow for Windows (amd64)..."
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o wm.exe

echo
echo "===================================="
echo "Build successful!"
echo "===================================="
echo
echo "Binary: wm.exe"
echo "Size: $(ls -lh wm.exe | awk '{print $5}')"
echo
echo "Transfer wm.exe to your Windows system and run as Administrator"
echo
