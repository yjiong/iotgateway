package device

import (
	"bytes"
	"encoding/binary"
	log "github.com/Sirupsen/logrus"
	"github.com/yjiong/go_tg120/config"
	"github.com/yjiong/go_tg120/internal/common"
	"sync"
	//	"strconv"
	//	"fmt"
	//	"strings"
)

var RegDevice = make(Devlist)
var Commif = make(map[string]string)
var Mutex = make(map[string]*sync.Mutex)

type dict map[string]interface{}

func init() {
	con, err := config.LoadConfigFile(common.CONFILEPATH)
	if err != nil {
		log.Errorf("load commif file failed : %s", err)
		return
	}
	comm, err := con.GetSection("commif")
	if err != nil {
		log.Errorf("get section commif file failed : %s", err)
		return
	}
	Commif = comm
	for ifname, _ := range comm {
		Mutex[ifname] = new(sync.Mutex)
		//log.Info(ifname)
	}
}

type DeviceRWer interface {
	NewDev(id string, ele map[string]string) (DeviceRWer, error)
	RWDevValue(rw string, m dict) (dict, error)
	CheckKey(e dict) (bool, error)
	GetElement() (dict, error)
	HelpDoc() interface{}
	//	Devid() string
}

type Device struct {
	devid   string
	devtype string
	commif  string
	devaddr string
	mutex   *sync.Mutex
}

type Devlist map[string]DeviceRWer

func NewDevHandler(devlistfile string) (map[string]DeviceRWer, error) {
	con, err := config.LoadConfigFile(devlistfile)
	if err != nil {
		log.Errorf("load config file failed: %s", err)
		return nil, err
	}
	devlist := map[string]DeviceRWer{}
	seclist := con.GetSectionList()
	for _, devid := range seclist {
		ele, err := con.GetSection(devid)
		if err != nil {
			log.Errorf("get %s element error : %s", devid, err)
			continue
		}
		dtype, ok_type := ele["_type"]
		if !ok_type {
			log.Errorf("get %s element type error : %s", devid, err)
			continue
		}
		if _, ok := ele["devaddr"]; !ok {
			log.Errorf("get %s element devaddr error : %s", devid, err)
			continue
		}
		if _, ok := ele["commif"]; !ok {
			log.Errorf("get %s element commif error : %s", devid, err)
			continue
		}
		if _, ok := RegDevice[dtype]; ok {
			devlist[devid], _ = RegDevice[dtype].NewDev(devid, ele)
		}
	}
	return devlist, nil
}

func (d *Device) NewDev(id string, ele map[string]string) Device {
	dmutex := new(sync.Mutex)
	return Device{
		devid:   id,
		devtype: ele["_type"],
		commif:  ele["commif"],
		devaddr: ele["devaddr"],
		mutex:   dmutex,
	}
}

func IntToBytes(n int) []byte {
	tmp := int32(n)
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, tmp)
	return bytesBuffer.Bytes()
}

func BytesToInt(b []byte) int {
	bytesBuffer := bytes.NewBuffer(b)
	var tmp int32
	binary.Read(bytesBuffer, binary.BigEndian, &tmp)
	return int(tmp)
}

func Hex2Bcd(n byte) byte {
	return IntToBytes(int(n)>>4*10 + int(n)&0x0f)[3]
}

func Bcd2Hex(n byte) byte {
	return IntToBytes((int(n)/10)<<4 + int(n)%10)[3]
}
func Bcd2_2f(a, b int) float64 {
	return float64((a>>4*10+a&0x0f)*100 + (b>>4*10 + b&0x0f))
}
