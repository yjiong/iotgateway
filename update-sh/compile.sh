#!/bin/bash
#########################################################################
# File Name: com-down.sh
# Author: yjiong
# mail: 4418229@qq.com
# Created Time: 2017-11-07 09:34:01
#########################################################################
export GOARCH=arm 
go build  -o ~/GOPATH/src/github.com/yjiong/go_tg120/gateway\
    /home/yjiong/GOPATH/src/github.com/yjiong/go_tg120/cmd/tg120/main.go 


