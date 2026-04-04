#!/bin/bash

OUT_DIR="./out/"
mkdir -p $OUT_DIR

go build -o $OUT_DIR/iwans ./src/iwans/ && echo Server build completed!
go build -o $OUT_DIR/iwanc ./src/iwanc/ && echo Converter build completed!
