#!/bin/bash

docker run -it --restart unless-stopped -p10046-10058:10046-10058 -p 80:80 -p 443:443 -v/tmp:/tmp/ttt -v/home/xy/mm/cloud/kod:/www/kod xyzj/homepc:latest
