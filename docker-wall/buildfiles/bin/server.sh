#!/bin/ash

start-stop-daemon -p /run/vpstools.pid --stop 
start-stop-daemon -p /run/v2ray.pid --stop

sleep 1

start-stop-daemon -m -p /run/v2ray.pid --start --background --exec /usr/bin/v2ray -- -c=/root/bin/v2server.json
start-stop-daemon -m -p /run/vpstools.pid --start --background --exec /root/bin/vpstools -- -http=2052

#sleep 1

#rc-service nginx restart
