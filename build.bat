@echo off

set "OUTPUT_DIR=./out/"
mkdir "%OUTPUT_DIR%" 2>nul

FOR /F "tokens=*" %%i IN ('git describe --tags --abbrev^=0') DO SET VERSION=%%i
echo:Building version %VERSION%...

go build -ldflags="-X main.version=%VERSION%" -o "%OUTPUT_DIR%iwans.exe" ./src/iwans/ && echo Server build completed!
go build -ldflags="-X main.version=%VERSION%" -o "%OUTPUT_DIR%iwanc.exe" ./src/iwanc/ && echo Converter build completed!