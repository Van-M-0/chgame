REM go build -i -v -o F:\chcore\bin\broker.exe apps/app-broker
REM go build -i -v -o F:\chcore\bin\proxy.exe apps/app-proxy
REM go build -i -v -o F:\chcore\bin\master.exe apps/app-master
REM go build -i -v -o F:\chcore\bin\gate.exe apps/app-gate
REM go build -i -v -o F:\chcore\bin\lobby.exe apps/app-lobby
REM go build -i -v -o F:\chcore\bin\world.exe apps/app-world

set GOARCH=amd64
set GOOS=linux
set GOPATH=F:/chcore;F:/goser/chexportor;F:/goser/chgamelib;F:/op/mygo

go build -i -v -o F:\chcore\bin\broker apps/app-broker
go build -i -v -o F:\chcore\bin\proxy apps/app-proxy
go build -i -v -o F:\chcore\bin\master apps/app-master
go build -i -v -o F:\chcore\bin\gate apps/app-gate
go build -i -v -o F:\chcore\bin\lobby apps/app-lobby
go build -i -v -o F:\chcore\bin\world apps/app-world




