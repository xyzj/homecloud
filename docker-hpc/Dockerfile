FROM ubuntu:latest
LABEL maintainer="X.Minamoto"
ENV DEBIAN_FRONTEND=noninteractive LANG=C.UTF-8

RUN			apt-get -y update; \
			apt-get -y upgrade; \
			apt-get -y install software-properties-common; \
            apt-get update; \
			apt-get -y install kmod curl socat apt-utils net-tools nano tzdata ssh cron nginx deluged deluge-web deluge-console mldonkey-server openssh-client php7.4 php7.4-fpm php7.4-curl php7.4-mbstring php7.4-gd php7.4-sqlite3; \
			apt-get -y autoremove; \
			apt-get -y clean; \
			apt-get -y autoclean; \
			rm -rfv /tmp/*; \
			\
			/bin/echo 'root:administratorishere' |chpasswd; \
            ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime; \
        	echo "Asia/Shanghai" > /etc/timezone; \
            dpkg-reconfigure -f noninteractive tzdata; \
            /bin/echo "net.ipv4.tcp_fastopen = 3">>/etc/sysctl.conf; \
            /bin/echo "net.ipv4.ip_forward=1">>/etc/sysctl.conf; \
            /bin/echo "net.ipv6.ip_forward=1">>/etc/sysctl.conf; \
            /bin/echo 'export PATH=$PATH:/root/bin'>> /root/.bashr

COPY		buildfiles /root

RUN			cp /root/bin/nginx.default /etc/nginx/sites-enabled/default; \
			mkdir /www; \
			chown -R www-data:www-data /www; \
			sed -i "s/='ngettext'/=['ngettext']/g" /usr/lib/python3/dist-packages/deluge/i18n/util.py
# WORKDIR		/root
# ENTRYPOINT	["/root/bin/run.sh"]
