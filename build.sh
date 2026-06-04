#!/bin/bash

OUT_DIR="./out/"
mkdir -p "$OUT_DIR"

VERSION=$(git describe --tags --abbrev=0 2>/dev/null | sed 's/^v//')
echo "Building version ${VERSION:-dev}..."

go build -ldflags="-X main.version=${VERSION:-dev}" -o "$OUT_DIR/iwans" ./src/iwans/ && echo "Server build completed!"
go build -ldflags="-X main.version=${VERSION:-dev}" -o "$OUT_DIR/iwanc" ./src/iwanc/ && echo "Converter build completed!"