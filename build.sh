#!/bin/bash
#rsrc -manifest hancom.exe.manifest -ico hancom.ico -o hancom.syso
rsrc -ico hancom.ico -o hancom.syso
GOOS=windows GOARCH=386 CGO_ENABLED=1 CXX=i686-w64-mingw32-g++ CC=i686-w64-mingw32-gcc go build -ldflags "-X main.version=1.2 -X main.build=20210727 -X main.platform=ubuntu" 
cp hancom.exe ../receipt/uploads
