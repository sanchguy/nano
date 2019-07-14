@echo off
cd ginrummy
set CGO_ENABLED=0
set GOOS=linux
set GOARCH=amd64
go build -o "../ginrummyServer/rummy_bin"
echo current path : %cd%
cd ../ginrummyServer
rd /q /s conf
md conf
cd ../ginrummy
copy "./game_init.ini" "../ginrummyServer/conf/"
echo "webserver build finish..."
pause
