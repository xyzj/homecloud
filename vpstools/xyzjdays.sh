#!/bin/bash

wget --no-check-certificate -o ca/xyzjdays.xyz.crt https://v4.xyzjdays.xyz/cert/cadir/xyzjdays.xyz.crt >/tmp/certdownload.log
wget --no-check-certificate -o ca/xyzjdays.xyz.key https://v4.xyzjdays.xyz/cert/cadir/xyzjdays.xyz.key >>/tmp/certdownload.log

sslcopy.sh
