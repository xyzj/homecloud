FROM ubuntu:latest
MAINTAINER      X.Minamoto "xuyuan8720@189.cn"

ENV DEBIAN_FRONTEND noninteractive

RUN			apt-get -y update; \
			apt-get -y upgrade; \
			apt-get -y install software-properties-common; \
			add-apt-repository ppa:deluge-team/stable; \
                        apt-get update; \
                        apt-get install deluge; \
			apt-get -y install kmod curl socat apt-utils net-tools nano tzdata ssh cron nginx deluged deluge-web deluge-console mldonkey-server openssh-client php7.2 php7.2-fpm php7.2-curl php7.2-mbstring php7.2-gd; \
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

COPY			buildfiles /root

RUN			cp /root/bin/nginx.default /etc/nginx/sites-enabled/default; \
			mkdir /www; \
			chown -R www-data:www-data /www
# WORKDIR		/root
# ENTRYPOINT	["/root/bin/run.sh"]
