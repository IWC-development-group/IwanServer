@echo off

set "OUTPUT_DIR=./out/"
mkdir "%OUTPUT_DIR%" 2>nul
go build -o "%OUTPUT_DIR%iwans.exe" ./src/ && echo Build completed!