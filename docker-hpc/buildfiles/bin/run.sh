#!/bin/bash

pkill -9 -f sslrenew
pkill -9 -f deluge-web
pkill -9 -f mlnet
# pkill -9 -f frpc

service nginx stop

start-stop-daemon --start --name sslrenew -d /root/bin --background --exec /root/bin/sslrenew -- -debug

service php7.4-fpm stop

sleep 1

service php7.4-fpm start

start-stop-daemon --start --name mlnet -d /root --background --exec /usr/bin/mlnet
/usr/bin/deluge-web -i 0.0.0.0 -p 60046

service nginx start

# cd /root/bin/frp
# start-stop-daemon --start --name frpc -d /root/bin/frp --background --exec frpc -- -c frpc.ini
