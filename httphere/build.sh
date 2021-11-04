#!/bin/bash

go build -ldflags="-s -w" -o httphere
upx httphere

GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o httphere.exe

GOOS=linux GOARCH=mipsle GOMIPS=softfloat CGO_ENABLED=0 go build -ldflags="-s -w" -o httphere-mt7621
upx httphere-mt7621
