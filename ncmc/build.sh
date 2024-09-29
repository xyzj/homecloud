#!/bin/bash

CGO_ENABLED=0 go build -ldflags="-s -w"

CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o ncmc.exe
