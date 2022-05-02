#!/bin/bash

docker run -it -p80:80 -p443:443 -p2052:2052 -v/root/short.conf:/root/bin/short.conf xyzj/wall:alpine
