@echo off
echo "Building..."

@cd ./tools/sound/ 
echo "sound..."
@%GOROOT%/bin/go build -o ./../../build/tools/

@cd ../..
echo "Build success"