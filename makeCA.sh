#!/bin/bash
#########################################################################
# File Name: makeca.sh
# Author: yjiong
# mail: 4418229@qq.com
# Created Time: 2017-12-14 17:29:37
##########################################################################
COUNTRYNAME=CN
STATEORPROVINCENAME=JIANGSU
LOCALITYNAME=WUXI
ORGANIZATIONNAME=YJ-technology
ORGANIZATIONALUNITNAME=yj-unit
CA_COMMONNAME=YJ-CA
SERVER_COMMONNAME=server_commonname
CLIENT_COMMONNAME=client_commonname
CA_EMAIL=yjiong@msn.com
SERVER_EMAIL=server@xxx.xxx
CLIENT_EMAIL=client@xxx.xxx

#server ip address
IPLIST="192.168.1.1 192.168.1.134 192.168.1.135 192.168.1.160  192.168.1.122"

#server name
HOSTLIST="www.yjiong.org www.xindong.org www.xindong.net www.xindong.com" 
#CA_ORG='/O=xxxxx.org/OU=org-tech/emailAddress=yourname@xxx.xxx'
#CA_DN="/CN=YJ-CA${CA_ORG}"
DAYS=3650
KEY=(ca.key server.key client.key)
CRT=(ca.crt server.crt client.crt)
REQ=(ca.req server.req client.req)
CONF=(ca.conf server.conf client.conf)
DER=(ca.der server.der client.der)
ALTIP=${IPLIST}
ALTDNS=${HOSTLIST}

function alt_names_list() {

	ALIST=""
	ALIST="${ALIST}IP:127.0.0.1,IP:::1,"
	for ip in $(echo ${ALTIP}); do
		ALIST="${ALIST}IP:${ip},"
	done
	for h in $(echo ${ALTDNS}); do
		ALIST="${ALIST}DNS:$h,"
	done
	ALIST="${ALIST}DNS:localhost"
	echo $ALIST

}
ALTNAME="$(alt_names_list)"
echo $ALTNAME
export ALTNAME
mkdir myca
cd myca
# Generate the openssl configuration files.
if [ ! -f ./ca.conf ]
then
cat > ca.conf << EOF  
[ req ]
default_bits		= 2048
default_keyfile 	= privkey.pem
distinguished_name	= req_distinguished_name
#attributes		= req_attributes
x509_extensions	= v3_ca	# The extensions to add to the self signed cert
#string_mask = utf8only
# Passwords for private keys if not present they will be prompted for
# input_password = secret
# output_password = secret

# This sets a mask for permitted string types. There are several options. 
# default: PrintableString, T61String, BMPString.
# pkix	 : PrintableString, BMPString (PKIX recommendation before 2004)
# utf8only: only UTF8Strings (PKIX recommendation after 2004).
# nombstr : PrintableString, T61String (no BMPStrings or UTF8Strings).
# MASK:XXXX a literal mask value.
# WARNING: ancient versions of Netscape crash on BMPStrings or UTF8Strings.
# req_extensions = v3_req # The extensions to add to a certificate request
prompt                 = no

[ req_distinguished_name ]
countryName			= $COUNTRYNAME
#countryName_default		= AU
#countryName_min			= 2
#countryName_max			= 2
stateOrProvinceName		= $STATEORPROVINCENAME
#stateOrProvinceName_default	= Some-State
localityName			= $LOCALITYNAME
0.organizationName		= $ORGANIZATIONNAME
#0.organizationName_default	= Internet Widgits Pty Ltd
## we can do this but it is not needed normally :-)
#1.organizationName		= second organizationName
##1.organizationName_default	= World Wide Web Pty Ltd
organizationalUnitName		= $ORGANIZATIONALUNITNAME
##organizationalUnitName_default	=
commonName			= $CA_COMMONNAME
#commonName_max			= 64
emailAddress			= $CA_EMAIL
#emailAddress_max		= 64
## SET-ex3			= SET extension number 3
#[ req_attributes ]
#challengePassword		= A challenge password
#challengePassword_min		= 4
#challengePassword_max		= 20
#unstructuredName		= An optional company name

[ v3_ca ]
# Extensions for a typical CA
# PKIX recommendation.
subjectKeyIdentifier=hash
authorityKeyIdentifier=keyid:always,issuer
basicConstraints = critical,CA:true
EOF
fi
cat > server.conf <<EOF  
[ req ]
distinguished_name     = req_distinguished_name
prompt                 = no

[ req_distinguished_name ]
countryName			= $COUNTRYNAME
stateOrProvinceName		= $STATEORPROVINCENAME
localityName			= $LOCALITYNAME
0.organizationName		= $ORGANIZATIONNAME
#1.organizationName		= second org name
organizationalUnitName		= $ORGANIZATIONALUNITNAME
commonName			= $SERVER_COMMONNAME
#commonName_max			= 64
emailAddress			= $SERVER_EMAIL
#emailAddress_max		= 64

[v3_req]
basicConstraints = CA:false
subjectAltName = \$ENV::ALTNAME
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
nsCertType = server
#nsComment = "Server Certificate"
subjectKeyIdentifier    = hash
authorityKeyIdentifier  = keyid,issuer:always
##certificatePolicies     = ia5org,@polsection
##[polsection]
##policyIdentifier	    = 1.3.5.8
##CPS.1		    = "http://localhost"
##userNotice.1	    = @notice
##[notice]
##explicitText            = "This CA is for  ......"
##organization            = "OwnTracks"
##noticeNumbers           = 1

[alt_names]
IP.1 = 192.168.1.1
DNS.1 = abc.example.com
EOF

cat > client.conf << EOF  
[ req ]
distinguished_name     = req_distinguished_name
prompt                 = no

[ req_distinguished_name ]
countryName			= $COUNTRYNAME
stateOrProvinceName		= $STATEORPROVINCENAME
localityName			= $LOCALITYNAME
commonName			= $CLIENT_COMMONNAME
emailAddress			= $CLIENT_EMAIL

[usr_cert]
basicConstraints        = critical,CA:false
subjectAltName          = email:copy
nsCertType              = client,email
extendedKeyUsage        = clientAuth,emailProtection
keyUsage                = digitalSignature, keyEncipherment, keyAgreement
#nsComment               = "Client Certificate"
subjectKeyIdentifier    = hash
authorityKeyIdentifier  = keyid,issuer:always
EOF
#openssl req -newkey rsa:2048 -x509 -nodes -sha512 -days $DAYS -extensions v3_ca -keyout ca.key -out ca.crt -subj "${CA_DN}"
#echo "Created CA certificate in $CACERT.crt"
#$openssl x509 -in $CACERT.crt -nameopt multiline -subject -noout

# private key generation
for key in ${KEY[@]}
do openssl genrsa -out $key 2048
done
# cert requests
for i in ${!KEY[@]}
do openssl req -new -out ${REQ[$i]} -key ${KEY[$i]} -config ./${CONF[$i]}
done
# generate the actual certs.
openssl x509 -req -in ca.req -out ca.crt \
            -sha512 -days $DAYS  -signkey ca.key \
            -extfile ./ca.conf \
            -extensions v3_ca 
openssl x509 -req -in server.req -out server.crt \
            -sha512 -CAcreateserial -days $DAYS \
            -CA ca.crt -CAkey ca.key \
            -extfile ./server.conf \
            -extensions v3_req
openssl x509 -req -in client.req -out client.crt \
            -sha512 -CAcreateserial -days $DAYS \
            -CA ca.crt -CAkey ca.key \
            -extfile ./client.conf \
            -extensions usr_cert

for i in ${!CRT[@]}
do openssl x509 -in ${CRT[$i]} -outform DER -out ${DER[$i]}
done
#mv ca.crt ca.key server.crt server.key client.crt client.key ca.der server.der client.der myca/
rm *.conf
rm *.req
rm *.srl
