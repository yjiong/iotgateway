#!/bin/bash
#########################################################################
# File Name: rmup.sh
# Author: yjiong
# mail: 4418229@qq.com
# Created Time: 2017-11-14 11:34:10
#########################################################################
if [ -z $0 ] || [ -z $2 ]
then echo "Usage:$0 username passwd"
    exit 1
fi
echo -e "更新结果\r" >updatelog.txt
for ip in `cat ipaddress.txt`
do  
    if [[ $ip =~ ^#.* ]];then
        continue
        fi
./autossh.exp $ip $1 $2 stop
ret=$?
if [ $ret -ne 0 ];then 
    echo  "connect $ip field " >>updatelog.txt
else
./autoscp.exp $ip $1 $2 tg150
./autossh.exp $ip $1 $2 start 
    echo  "$ip update successful " >>updatelog.txt
fi
done
