#!/bin/bash
#########################################################################
# File Name: com-down.sh
# Author: yjiong
# mail: 4418229@qq.com
# Created Time: 2017-11-07 09:34:01
#########################################################################
export GOARCH=$1 
#go build  -o ~/GOPATH/src/github.com/yjiong/go_tg120/gateway\
go build  -ldflags "-s -w" -o ~/update-sh/iot/gateway\
    /home/yjiong/GOPATH/src/github.com/yjiong/iotgateway/cmd/gateway/main.go 


