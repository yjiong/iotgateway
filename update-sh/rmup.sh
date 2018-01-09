#!/bin/bash
#########################################################################
# File Name: rmup.sh
# Author: yjiong
# mail: 4418229@qq.com
# Created Time: 2017-11-14 11:34:10
#########################################################################
if [ -z $1 ] || [ -z $2 ] || [ -z $3 ]
then echo "Usage:$0 username passwd filename"
    USER=root
    PW=93010005
    FILE=tg150
    #exit 1
else
    USER=$1
    PW=$2
    FILE=$3
fi

echo -e "更新结果\r" >updatelog.txt
for ip in `cat ipaddress.txt`
do  
    if [[ $ip =~ ^#.* ]];then
        continue
        fi
./autossh.exp $ip ${USER} ${PW} stop
ret=$?
if [ $ret -ne 0 ];then 
    echo  "connect $ip failed " >>updatelog.txt
else
    ./autoscp.exp $ip ${USER} ${PW} ${FILE}
    retscp=$?
    ./autossh.exp $ip ${USER} ${PW} start 
    if [ $retscp -ne 0 ];then 
        echo  "$ip update failed " >>updatelog.txt
    else
        echo  "$ip update successful " >>updatelog.txt
    fi
fi
done
