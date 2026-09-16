#!/bin/bash
# Go Backend iOS Build Script via Gomobile
set -e

echo "Installing gomobile..."
go install golang.org/x/mobile/cmd/gomobile@latest
gomobile init

echo "Building Tmaxbackend.xcframework for iOS..."
gomobile bind -target=ios -o Tmaxbackend.xcframework ./tmaxbackend

echo "Successfully built Tmaxbackend.xcframework!"
