#!/bin/ash

start-stop-daemon --stop -p /run/v2client.pid

sleep 1

start-stop-daemon --start --background -m -p /run/v2client.pid /usr/bin/v2ray -- -c /root/bin/v2client.json
