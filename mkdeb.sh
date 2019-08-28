#!/bin/bash
#########################################################################
# File Name: mkdev.sh
# Author: yjiong
# mail: 4418229@qq.com
# Created Time: 2019-08-28 10:09:53
#########################################################################

CWD=$(cd `dirname $0`;pwd)
PKGDIR=debpackage
CONTROLFILE=${CWD}/${PKGDIR}/DEBIAN/control
BINFILE=${CWD}/${PKGDIR}/opt/iot/gateway
DEBS=(gateway-arm64-v1.3.deb gateway-armhf-v1.3.deb gateway-amd64-v1.3.deb )
BINS=(gateway-arm64 gateway-armhf gateway-amd64)
ARCHS=(arm64 arm amd64)
ARCHSCONT=(arm64 armhf amd64)

for i in `seq 0 2`
do
    env GOARCH=${ARCHS[$i]} go build --tags iotd -ldflags '-s -w -X main.VERSION=v1.3.18' -o $CWD/${BINS[$i]}\
        $CWD/cmd/gateway/
    cp -f ${BINS[$i]} $BINFILE
    sed -i  "s/\(Architecture:\).\+/\1 ${ARCHSCONT[$i]}/" "$CONTROLFILE"
    dpkg -b $PKGDIR ${DEBS[$i]}
    rm ${BINS[$i]}
done
