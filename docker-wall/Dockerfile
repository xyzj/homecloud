FROM alpine
LABEL maintainer="X.Minamoto"
ENV DEBIAN_FRONTEND=noninteractive LANG=C.UTF-8

COPY	buildfiles /root

RUN	/sbin/apk update && \
	/sbin/apk upgrade && \
	/sbin/apk add nginx v2ray openrc && \
	/bin/cp /root/bin/localtime /etc/localtime && \
	/bin/cp /root/bin/default.conf /etc/nginx/http.d && \
	/sbin/rc-update add nginx default --update && \
	/bin/touch /run/openrc/softlevel && \
        /bin/echo "net.ipv4.tcp_fastopen = 3">>/etc/sysctl.conf; \
        /bin/echo "net.ipv4.ip_forward=1">>/etc/sysctl.conf; \
        /bin/echo "net.ipv6.ip_forward=1">>/etc/sysctl.conf; \
