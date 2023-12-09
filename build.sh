#bin/bash
set -m
echo "Building..."

echo build tools
cd ./tools/sound/ 
$GOROOT/bin/go build -o ./../../build/tools/
echo sound


echo "Build success"