@echo off

set "OUTPUT_DIR=./out/"
mkdir "%OUTPUT_DIR%" 2>nul

go build -o "%OUTPUT_DIR%iwans.exe" ./src/iwans/ && echo Server build completed!
go build -o "%OUTPUT_DIR%iwanc.exe" ./src/iwanc/ && echo Converter build completed!