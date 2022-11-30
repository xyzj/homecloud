#!/bin/ash

start-stop-daemon -p /run/vpstools.pid --stop
start-stop-daemon -p /run/v2ray.pid --stop

sleep 1

start-stop-daemon -m -p /run/v2ray.pid -d /root/bin --start --background --exec /root/bin/v2ray -- run -c /root/bin/v2server.json
start-stop-daemon -m -p /run/vpstools.pid --start --background --exec /root/bin/vpstools -- -http=2052

sleep 1

start-stop-daemon --start --background -m -p /run/caddy.pid --exec /usr/sbin/caddy -- run --config /etc/caddy/Caddyfile

#rc-service nginx restart
