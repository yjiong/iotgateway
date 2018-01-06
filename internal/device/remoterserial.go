package device

import (
	"errors"
	log "github.com/Sirupsen/logrus"
	simplejson "github.com/bitly/go-simplejson"
	"github.com/yjiong/go_tg120/serial"
	"strconv"
	"time"
)

//RemotePort ...
var RemotePort serial.Port
var sp string
var readInterval int

// Openser ...
/*
{
  "cmd": "remoteSerial",
  "parse": "openser",
  "data": {
    "port":"/dev/ttyUSB0",
    "baudrate": "9600",
    "databits": "8",
    "stopbits": "1",
    "parity": "E"
  }
}
*/
func Openser(param *simplejson.Json) (err error) {
	par, _ := param.Map()
	resConfig := serial.Config{}
	if cif, ok := par["port"].(string); ok && cif != "" {
		sp = cif
		resConfig.Address = Commif[cif]
	} else {
		resConfig.Address = "/dev/ttyUSB0"
		sp = "usb0"
	}
	Mutex[sp].Lock()
	baudrate, _ := par["baudrate"].(string)
	resConfig.BaudRate, _ = strconv.Atoi(baudrate) // d.BaudRate
	databits, _ := par["databits"].(string)
	resConfig.DataBits, _ = strconv.Atoi(databits) //d.DataBits
	resConfig.Parity, _ = par["parity"].(string)   //d.Parity
	stopbits, _ := par["stopbits"].(string)
	resConfig.StopBits, _ = strconv.Atoi(stopbits) // d.StopBits
	if readstrval, ok := par["readInterval"]; ok {
		if readstr, ok := readstrval.(string); ok {
			readInterval, _ = strconv.Atoi(readstr) // d.StopBits
		}
	} else {
		readInterval = 2560000 / resConfig.BaudRate
	}
	resConfig.Timeout = time.Second * 2
	for i := 0; i < 5; i++ {
		RemotePort, err = serial.Open(&resConfig)
		if err == nil {
			log.Infof("open %s successful", resConfig.Address)
			return err
		}
		time.Sleep(time.Second)
	}
	log.Infof("open %s faild", resConfig.Address)
	return errors.New("open serial faild")
}

// Closeser ...
/*
{
  "cmd": "remoteSerial",
  "parse": "closeser"
}
*/
func Closeser() error {
	Mutex[sp].Unlock()
	return RemotePort.Close()
}

// Wser ...
/*
{
  "cmd": "remoteSerial",
  "parse": "wser",
  "data":88888888
}
*/
func Wser(data []byte) error {
	_, err := RemotePort.Write(data)
	return err
}

// Rser ...
func Rser() (results []byte, err error) {
	var len int
	results = make([]byte, 256)
	time.Sleep(time.Duration(readInterval) * time.Millisecond)
	len, err = RemotePort.Read(results)
	if len != 0 {
		return results[:len], err
	}
	return nil, err
}
