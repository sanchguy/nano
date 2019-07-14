@echo off
cd truco
set CGO_ENABLED=0
set GOOS=windows
set GOARCH=amd64
go build -o "../truco_bin.exe"
pause
