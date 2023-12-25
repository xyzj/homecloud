#!/bin/bash

CGO_ENABLED=0 go build -ldflags="-s -w"

CGO_ENABLED=0 GOOW=windows GOARCH=amd64 go build -ldflags="-s -w" -o ncmc.exe
