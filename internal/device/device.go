package device

import (
	"bytes"
	"encoding/binary"
	"math"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/yjiong/iotgateway/config"
	"github.com/yjiong/iotgateway/internal/common"
	"github.com/yjiong/iotgateway/serial"
)

// RegDevice ..
var RegDevice = make(Devlist)

// Commif ..
var Commif = make(map[string]string)

// Mutex ..
var Mutex = make(map[string]*sync.Mutex)

// Dict ...
type Dict map[string]interface{}

func init() {
	con, err := config.LoadConfigFile(common.CONFILEPATH)
	if err != nil {
		log.Errorf("load commif file failed : %s", err)
		return
	}
	comm, err := con.GetSection("commif")
	if err != nil {
		log.Errorf("get section Commif file failed : %s", err)
		return
	}
	Commif = comm
	for ifname := range comm {
		Mutex[ifname] = new(sync.Mutex)
		//log.Info(ifname)
	}
	Mutex["usb0"] = new(sync.Mutex)
}

// Devicerwer ..
type Devicerwer interface {
	NewDev(id string, ele map[string]string) (Devicerwer, error)
	RWDevValue(rw string, m Dict) (Dict, error)
	CheckKey(e Dict) (bool, error)
	GetElement() (Dict, error)
	HelpDoc() interface{}
	GetCommif() string
	//	Devid() string
}

// Device ..
type Device struct {
	Dname      string
	DepService bool
	Devid      string
	Devtype    string
	Commif     string
	Devaddr    string
	Mutex      *sync.Mutex
}

// Devlist ..
type Devlist map[string]Devicerwer

// NewDevHandler ..
func NewDevHandler(devlistfile string) (Devlist, error) {
	con, err := config.LoadConfigFile(devlistfile)
	if err != nil {
		log.Errorf("load config file failed: %s", err)
		return nil, err
	}
	devlist := Devlist{}
	seclist := con.GetSectionList()
	for _, Devid := range seclist {
		ele, err := con.GetSection(Devid)
		if err != nil {
			log.Errorf("get %s element error : %s", Devid, err)
			continue
		}
		dtype, okType := ele[DevType]
		if !okType {
			log.Errorf("get %s element type error : %s", Devid, err)
			continue
		}
		if _, ok := ele[DevAddr]; !ok {
			log.Errorf("get %s element Devaddr error : %s", Devid, err)
			continue
		}
		if _, ok := ele["commif"]; !ok {
			log.Errorf("get %s element Commif error : %s", Devid, err)
			continue
		}
		if _, ok := RegDevice[dtype]; ok {
			devlist[Devid], _ = RegDevice[dtype].NewDev(Devid, ele)
		}
	}
	return devlist, nil
}

// NewDev ..
func (d *Device) NewDev(id string, ele map[string]string) Device {
	dMutex := new(sync.Mutex)
	dnm := ""
	if nm, ok := ele[DevName]; ok {
		dnm = nm
	}
	return Device{
		Dname:   dnm,
		Devid:   id,
		Devtype: ele[DevType],
		Commif:  ele["commif"],
		Devaddr: ele[DevAddr],
		Mutex:   dMutex,
	}
}

// GetCommif return Commif
func (d *Device) GetCommif() string {
	return d.Commif
}

// IntToBytes ..
func IntToBytes(n int) []byte {
	tmp := int32(n)
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, tmp)
	return bytesBuffer.Bytes()
}

// BytesToInt ..
func BytesToInt(b []byte) int {
	bytesBuffer := bytes.NewBuffer(b)
	var tmp int32
	binary.Read(bytesBuffer, binary.BigEndian, &tmp)
	return int(tmp)
}

// Hex2Bcd ..
func Hex2Bcd(n byte) byte {
	return IntToBytes(int(n)>>4*10 + int(n)&0x0f)[3]
}

// Bcd2Hex ..
func Bcd2Hex(n byte) byte {
	return IntToBytes((int(n)/10)<<4 + int(n)%10)[3]
}

// Bcd2_2f ..
func Bcd2_2f(a, b int) float64 {
	return float64((a>>4*10+a&0x0f)*100 + (b>>4*10 + b&0x0f))
}

// Float32ToByte ..
func Float32ToByte(float float32) []byte {
	bits := math.Float32bits(float)
	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, bits)

	return bytes
}

//ByteToFloat32 ..
func ByteToFloat32(bytes []byte) float32 {
	bits := binary.LittleEndian.Uint32(bytes)

	return math.Float32frombits(bits)
}

// Float64ToByte ..
func Float64ToByte(float float64) []byte {
	bits := math.Float64bits(float)
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, bits)

	return bytes
}

// ByteToFloat64 ..
func ByteToFloat64(bytes []byte) float64 {
	bits := binary.LittleEndian.Uint64(bytes)

	return math.Float64frombits(bits)
}

//SerialRead ..
func SerialRead(serport serial.Port, timeout time.Duration, results []byte) (rblen int, err error) {
	go func() {
		for {
			time.Sleep(time.Millisecond * 100)
			timeout -= (time.Millisecond * 100)
			if timeout <= 0 {
				break
			}
		}
	}()
	for {
		var resArry []byte
		for {
			results := make([]byte, 256)
			if rblen, err = serport.Read(results); rblen != 0 && err == nil {
				resArry = append(resArry, results[:rblen]...)
			} else {
				//time.Sleep(10 * time.Millisecond)
				break
			}
		}
		if rblen = len(resArry); rblen > 0 {
			//fmt.Printf("% x  len=%d\n", resArry, rblen)
			for i, b := range resArry {
				results[i] = b
			}
			err = nil
			timeout = 0
			return
		}
		if timeout == 0 && err != nil {
			break
		}
	}
	return
}

//StringReverse ...
func StringReverse(str string) string {
	var rstr string
	slicstr := []byte(str)
	for i := len(slicstr); i > 0; i-- {
		rstr += string(slicstr[i-1])
	}
	return rstr
}

//HexStringReverse ...
func HexStringReverse(str string) string {
	var rstr string
	slicstr := []byte(str)
	for i := len(slicstr); i > 0; i -= 2 {
		rstr += string(slicstr[i-2 : i])
	}
	return rstr
	//return strings.TrimLeft(rstr, "0")
}
