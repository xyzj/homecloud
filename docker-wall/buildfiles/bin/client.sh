#!/bin/ash

start-stop-daemon --stop -p /run/v2client.pid

sleep 1

start-stop-daemon --start --background -m -p /run/v2client.pid /root/bin/v2ray -- run -c /root/bin/v2less.json

sleep 1

#start-stop-daemon --start --background -m -p /run/caddy.pid --exec /usr/sbin/caddy -- run --config /etc/caddy/Caddyfile
