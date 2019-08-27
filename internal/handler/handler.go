package handler

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"reflect"
	"syscall"
	"time"
)

// Handler defines the interface of a handler backend.
type Handler interface {
	Close() error                       // 断开mqtt连接
	SendDataUp(interface{}) error       // 发送上行数据
	SendSerDataUp([]byte) error         // 发送上行数据
	DataDownChan() chan DataDownPayload // 返回订阅到的消息数据channel
	IsConnected() bool
	GetLostConnectTime() time.Time
	SubFirmware(topic string) chan []byte
}

//openssl genrsa -out private.pem 2048
//openssl rsa -in private.pem -pubout -out public.pem

const (
	privatekey = `
-----BEGIN RSA PRIVATE KEY-----
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
-----END RSA PRIVATE KEY-----
	`
	publickey = `
-----BEGIN PUBLIC KEY-----
MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAsqc1/kZCrlWQ52rDFp2w
OG83f92FHGkQIDri/ZlenaPehVhJcGvbBMWzUh/w+e9XyO6IN79HKItcul3VZwJa
8Us8xIYv9qP0EgpzjVW/jhb4NQDwiDOT9VDEOvZnSUTTV6hkZj8TByhX+Fj3Q7Qd
4Ri3sjHPNvciyfEJcXSniuDXQVX3C6utdPuoB5GmuUzGsMnjZQ5w/FcW7uiiTg/4
SuXRKUPdf57t0N0L6JynC092H34+pL1iuI6JNXbs+PmxY2m4446nyMI7aALILySW
eON35Yabv2khA91zfUyyQZY+QQ8YsyJtOPrKZ/vAOJlLY6bMEvEUCdDZCna19/HS
UU3gzLIMOVf8pytu92iv/5pgCpPUs7hyIleskK93F+z9SIY3T0SrhxMmBuAg1/DD
TKU2kDDZaqBdvDNyKTFbTZV1HX9BUeZYNhXKKZVPRtqBbxanl96l0KN5b9BmXcei
WiHWItRuxNRmTBiHdUtoHyfeKUT3N64j67dRyZI5ZWmyAApODK9nqC21HtgOUMvB
INIe2mAwbbS8leCLoQiAlRU20Um8STF8WuqeDS1kFftomBWUaVbuYCUgFchEbxLV
mzbflpMraMsmZC0annqql+rh13L+n5x+roWBmS4D+6N3qZqWl7T8yINvuSIdZrcY
Z226VbSOYndIpNJ+u6DGtAMCAwEAAQ==
-----END PUBLIC KEY-----
	`
)

//RsaEncrypt encryption the origin usr public key
func RsaEncrypt(origin []byte) ([]byte, error) {
	block, _ := pem.Decode([]byte(publickey))
	if block == nil {
		return nil, errors.New("pubkey error")
	}
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	pubif := pubInterface.(*rsa.PublicKey)
	ret, err := rsa.EncryptPKCS1v15(rand.Reader, pubif, origin)
	return ret, err
}

//RsaDecrypt decryption the ciphertext usr private ckey
func RsaDecrypt(ciphertext []byte) ([]byte, error) {
	block, _ := pem.Decode([]byte(privatekey))
	if block == nil {
		return nil, errors.New("pubkey error")
	}
	private, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	cleartext, err := rsa.DecryptPKCS1v15(rand.Reader, private, ciphertext)
	return cleartext, err
}

//Serify ...
func Serify(cipherfile, clearfile string) bool {
	if macAddr, err := ioutil.ReadFile(clearfile); err == nil {
		if enc, err := ioutil.ReadFile(cipherfile); err == nil {
			if dec, err := RsaDecrypt(enc); err == nil {
				return reflect.DeepEqual(dec, macAddr)
			}
		}
	}
	return false
}

//GenEncryptoFile ..
func GenEncryptoFile(dpath, spath string) error {
	if dbyte, err := ioutil.ReadFile(spath); err == nil {
		if enc, err := RsaEncrypt(dbyte); err == nil {
			if ioutil.WriteFile(dpath, enc, 0666) == nil {
				syscall.Sync()
				return nil
			}
		}
	}
	return errors.New("encrypto failed")
}
