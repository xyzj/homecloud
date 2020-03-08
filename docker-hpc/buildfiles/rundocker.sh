#!/bin/bash

docker run -it -p80:80 -p443:443 -p8112-8115:8112-8115 -p6880-6899:6880-6899 -v/tmp:/tmp/ttt -v/home/xy/mm/cloud/KodExplorer:/root/kodexplorer -v/home/xy/mm/xldown:/root/xldown xyzj/docker-hpc:latest
