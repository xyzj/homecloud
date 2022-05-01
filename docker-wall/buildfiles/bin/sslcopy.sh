#!/bin/ash
cp -f .lego/certificates/_.xyzjdays.xyz.crt /root/bin/ca/xyzjdays.xyz.crt
cp -f .lego/certificates/_.xyzjdays.xyz.key /root/bin/ca/xyzjdays.xyz.key
rc-service nginx restart


