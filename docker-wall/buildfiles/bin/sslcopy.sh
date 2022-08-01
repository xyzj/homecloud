#!/bin/ash

cp -f /root/bin/.lego/certificates/xyzjx.xyz.crt /root/bin/ca/xyzjx.xyz.crt
cp -f /root/bin/.lego/certificates/xyzjx.xyz.key /root/bin/ca/xyzjx.xyz.key

start-stop-daemon --stop -p /run/caddy.pid

sleep 1

start-stop-daemon --start --background -m -p /run/caddy.pid --exec /usr/sbin/caddy -- run --config /etc/caddy/Caddyfile

#rc-service nginx restart
