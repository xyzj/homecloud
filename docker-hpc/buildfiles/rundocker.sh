#!/bin/bash

docker run -it --restart unless-stopped -p8112-8117:8112-8117 -p 80:80 -p 443:443 -p6875-6990:6875-6990 -v/tmp:/tmp/ttt -v/home/xy/mm/cloud/kod:/www/kod xyzj/homepc:latest
