#!/bin/bash
#########################################################################
# File Name: migrate.sh
# Author: yjiong
# mail: 4418229@qq.com
# Created Time: 2019-08-14 15:09:07
#########################################################################
CWD="$( cd "$( dirname "$0"  )" && pwd -P )" 
PKGNAME=$(basename $CWD)
IMPORT=\"github.com\/yjiong\/iotgateway\/internal\/device\"
SSTR=(RegDevice Dict Device Commif Mutex Devicerwer IntToBytes BytesToInt Hex2Bcd Bcd2Hex Bcd_2f Float32ToByte \
    ByteToFloat32 Float64ToByte ByteToFloat64 SerialRead StringReverse HexStringReverse ModbusRtu ModbusTCP)
SEDCMD="$(which sed) -r -i "

$(which find) ${CWD} -name "*.go" -type f -print0 |xargs -0 ${SEDCMD} "s/(package ).*/\1$PKGNAME/;/internal\/device/d;/import.*/a \ \ \ \ ${IMPORT}"
$(which find) ${CWD} -name "*.go" -type f -print0 |xargs -0 ${SEDCMD} "s/(d\.)(dev)/\1\u\2/g;s/(d\.)(comm)/\1\u\2/g;s/(d\.)(mutex)/\1\u\2/g;s/dict/Dict/g"

for sstr in ${SSTR[*]}
do
    $(which find) ${CWD} -name "*.go" -type f -print0 |xargs -0 ${SEDCMD} "s/(\\s|\()(${sstr})/\1device.\2/g"
done
