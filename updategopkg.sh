#!/bin/bash

if [ ! "$1" = "localonly" ]; then
    export https_proxy=http://192.168.1.9:6871

    # go get golang.org/x/tools/cmd/godoc
    echo "go get gjson/sjson"
    go get -u github.com/tidwall/gjson
    go get -u github.com/tidwall/sjson
    echo "go get redis"
    go get -u github.com/go-redis/redis
    echo "go get sql-driver"
    go get -u github.com/go-sql-driver/mysql
    go get -u github.com/denisenkom/go-mssqldb
    echo "go get rabbitmq"
    go get -u github.com/streadway/amqp
    echo "go get proto"
    go get -u github.com/golang/protobuf/proto
    go get -u github.com/golang/protobuf/protoc-gen-go
    go get -u github.com/gogo/protobuf/proto
    go get -u github.com/gogo/protobuf/gogoproto
    go get -u github.com/gogo/protobuf/protoc-gen-gogofaster
    echo "go get grpc"
    go get -u google.golang.org/grpc
    go get -u google.golang.org/grpc/keepalive
    go get -u google.golang.org/grpc/credentials
    echo "go get etcdv3"
    go get -u go.etcd.io/etcd/clientv3
    go get -u go.etcd.io/etcd/pkg/transport
    echo "go get gin"
    go get -u github.com/json-iterator/go
    go get -u github.com/gin-gonic/gin
    go get -u github.com/gin-gonic/gin/render
    go get -u github.com/gin-contrib/gzip
    go get -u github.com/gin-contrib/pprof
    go get -u github.com/gin-contrib/multitemplate
    go get -u github.com/gin-contrib/cors
    echo "go get base64Captcha"
    go get -u github.com/mojocn/base64Captcha

    echo "go get others"
    go get -u github.com/robfig/cron
    go get -u github.com/tealeg/xlsx
    go get -u github.com/google/uuid
    go get -u github.com/golang/snappy
    go get -u github.com/pierrec/lz4
    go get -u github.com/pkg/errors
    go get -u github.com/xyzj/gopsu
    
    go install github.com/xyzj/gopsu/db
    go install github.com/xyzj/gopsu/mq
    go install github.com/xyzj/gopsu/microgo
    go install github.com/xyzj/gopsu/gin-middleware

    go download github.com/xyzj/wlstmicro
    go install github.com/xyzj/wlstmicro
    
    export -n https_proxy

    rm -rf ~/go/src/go.etcd.io/etcd/vendor/golang.org/x/net/trace
fi
