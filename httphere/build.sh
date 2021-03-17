#!/bin/bash

go build -ldflags="-s -w" -o httphere

GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o httphere.exe
