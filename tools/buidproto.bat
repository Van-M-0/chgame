
echo " "
echo "build proto start"

cd %cd%

echo %cd%

set root=%cd%
set exedir=%root%/protoc-3.3.0-win32/bin

cd %cd%/..

echo %cd%

set protodir=%cd%/src/gameproto/cliproto
set outdir=%cd%/src/gameproto/clipb


cd %root%

%exedir%\protoc.exe --proto_path=%protodir%/ --plugin=protoc-gen-go=%GOPATH%/bin/protoc-gen-go.exe --go_out=%outdir%/ %protodir%\*.proto

echo "proto build finish"

pause