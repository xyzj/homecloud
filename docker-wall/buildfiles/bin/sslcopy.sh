#!/bin/ash

cp -f /root/bin/.lego/certificates/xyzjdays.xyz.crt /root/bin/ca/xyzjdays.xyz.crt
cp -f /root/bin/.lego/certificates/xyzjdays.xyz.key /root/bin/ca/xyzjdays.xyz.key

#start-stop-daemon --stop -p /run/caddy.pid

sleep 1
caddy reload
#start-stop-daemon --start --background -m -p /run/caddy.pid --exec /usr/sbin/caddy -- run --config /etc/caddy/Caddyfile

#rc-service nginx restart
