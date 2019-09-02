@echo off
cd truco
set CGO_ENABLED=0
set GOOS=linux
set GOARCH=amd64
go build -o "../truco_bin"
echo "webserver build finish..."
pause
