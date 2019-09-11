package watermeter

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/yjiong/iotgateway/internal/device"
	"github.com/yjiong/iotgateway/serial"
	"strconv"
	"time"
)

//ZENNER ..
type ZENNER struct {
	//组合Device
	device.Device
	/**************按不同设备自定义*************************/
	//BaudRate int
	//DataBits int
	//StopBits int
	//Parity   string
	/**************按不同设备自定义*************************/
}

func init() {
	device.RegDevice["ZENNER"] = &ZENNER{}
}

// NewDev ..
func (d *ZENNER) NewDev(id string, ele map[string]string) (device.Devicerwer, error) {
	ndev := new(ZENNER)
	ndev.Device = d.Device.NewDev(id, ele)
	/***********************初始化设备的特有的参数*****************************/
	//ndev.BaudRate, _ = strconv.Atoi(ele["BaudRate"])
	//ndev.DataBits, _ = strconv.Atoi(ele["DataBits "])
	//ndev.StopBits, _ = strconv.Atoi(ele["StopBits"])
	//ndev.Parity, _ = ele["Parity"]
	/***********************初始化设备的特有的参数*****************************/
	return ndev, nil
}

// GetElement ..
func (d *ZENNER) GetElement() (device.Dict, error) {
	conn := device.Dict{
		/***********************设备的特有的参数*****************************/
		device.DevAddr: d.Devaddr,
		"commif":       d.Commif,
		//"BaudRate": d.BaudRate,
		//"DataBits": d.DataBits,
		//"StopBits": d.StopBits,
		//"Parity":   d.Parity,
		/***********************设备的特有的参数*****************************/
	}
	data := device.Dict{
		device.DevID:   d.Devid,
		device.DevType: d.Devtype,
		device.DevConn: conn,
	}
	return data, nil
}

/***********************设备的参数说明帮助***********************************/

// HelpDoc ..
func (d *ZENNER) HelpDoc() interface{} {
	conn := device.Dict{
		device.DevAddr: "设备地址",
		/***********TOSHIBA设备的参数*****************************/
		"commif": "通信接口,比如(rs485-1)",
		/***********TOSHIBA设备的参数*****************************/
	}
	rParameter := device.Dict{
		device.DevID: "被读取设备对象的id",
		/***********读取设备的参数*****************************/
		/***********读取设备的参数*****************************/
	}
	wParameter := device.Dict{
		device.DevID: "被操作设备对象的id",
		/***********操作设备的参数*****************************/
		//"初始化":  initval,
	}
	data := device.Dict{
		device.DevID:   "添加设备对象的id",
		device.DevType: "ZENNER", //设备类型
		device.DevConn: conn,
	}
	devUpdate := device.Dict{
		"request": device.Dict{
			"cmd":  device.UpdateDevItem,
			"data": data,
		},
	}
	readdev := device.Dict{
		"request": device.Dict{
			"cmd":  device.GetDevVar,
			"data": rParameter,
		},
	}
	writedev := device.Dict{
		"request": device.Dict{
			"cmd":  device.SetDevVar,
			"data": wParameter,
		},
	}
	helpdoc := device.Dict{
		"1.添加设备": devUpdate,
		"2.读取设备": readdev,
		"3.操作设备": writedev,
	}
	return helpdoc
}

/***********************设备的参数说明帮助***********************************/

/***************************************添加设备参数检验**********************************************/

// CheckKey ..
func (d *ZENNER) CheckKey(ele device.Dict) (bool, error) {
	im, imok := ele["commif"].(string)
	if !imok {
		return false, fmt.Errorf("device must have string type element 通讯接口 :commif")
	}
	if _, ok := device.Commif[im]; !ok {
		return false, fmt.Errorf("通讯接口:%s不存在", im)
	}
	return true, nil
}

/***************************************添加设备参数检验**********************************************/

func (d *ZENNER) readCmd(taddr string, di []byte) []byte {
	var saddr string
	cmd := []byte{0x68, 0x19}
	if len(taddr) <= 10 {
		saddr = fmt.Sprintf("%010s", taddr)
	} else {
		saddr = fmt.Sprintf("%014s", taddr)
	}
	if d.Devtype == "ZENNER-S" || d.Devtype == "WM-LXZD" || d.Devtype == "WHSF-L" {
		cmd[1] = 0x10
	}
	for i := len(saddr); i >= 2; i -= 2 {
		temp, _ := strconv.Atoi(saddr[i-2 : i])
		cmd = append(cmd, device.Bcd2Hex(device.IntToBytes(temp)[3]))
	}
	if len(taddr) <= 10 {
		cmd = append(cmd, []byte{0x33, 0x78, 0x01, 0x03, 0x1F, 0x90, 0x01}...)
	} else {
		cmd = append(cmd, []byte{0x01, 0x03, 0x90, 0x1f, 0x00}...)
	}
	//读地址
	//cmd = []byte{0x68, 0x19, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0x03, 0x03, 0x0A, 0x81, 0x01}
	//cmd = []byte{0x68, 0x10, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0x03, 0x03, 0x0A, 0x81, 0x01}
	//cdi := d.plus33(di)
	cmd = append(cmd, di...)
	sum := d.rDateSum(cmd)
	cmd = append(cmd, sum)
	cmd = append(cmd, 0x16)
	rcmd := append([]byte{0xFE, 0xFE, 0xFE, 0xFE}, cmd...)
	return rcmd
}

func (d *ZENNER) rDateSum(data []byte) byte {
	sum := 0
	for i := 0; i < len(data); i++ {
		sum += int(data[i])
	}
	return device.IntToBytes(sum & 0xff)[3]
}

func (d *ZENNER) rdvalue(rsport serial.Port, cmd int, ret map[string]interface{}) (err error) {
	var di []byte
	results := make([]byte, 80)
	//di = []byte{0x27, 0, 0}
	di = []byte{}
	for i := 0; i < 2; i++ {
		log.Debugf("send cmd = % x", d.readCmd(d.Devaddr, di))
		if _, err = rsport.Write(d.readCmd(d.Devaddr, di)); err != nil {
			log.Errorf("send cmd error  %s", err.Error())
			return err
		}
		time.Sleep(800 * time.Millisecond)
		var rlen int
		rlen, err = device.SerialRead(rsport, time.Second, results)
		if err == nil {
			var startbyte int
			for i, v := range results {
				if v == 0x68 {
					startbyte = i
					break
				}
			}
			log.Debugf("receive data = % x rlen = %d ", results[:rlen], rlen)
			//log.Debugf("receive data = % x rlen = %d sum = %x ", results[startbyte:rlen], rlen, d.rDateSum(results[startbyte:rlen-2]))
			if rlen > 9 && d.rDateSum(results[startbyte:rlen-2]) == results[rlen-2] &&
				results[startbyte] == 0x68 && results[rlen-1] == 0x16 {
				log.Debugf("校验正确")
				valb := results[startbyte+14 : startbyte+18]
				ret["当前读数"] = (float64(device.Hex2Bcd(valb[0])) +
					float64(device.Hex2Bcd(valb[1]))*100 +
					float64(device.Hex2Bcd(valb[2]))*10000 +
					float64(device.Hex2Bcd(valb[3]))*1000000) / 100
				err = nil
				break
			} else {
				err = errors.New("校验错误")
			}
		}
		if i < 1 {
			time.Sleep(300 * time.Millisecond)
		}
	}
	return
}

// RWDevValue ..
func (d *ZENNER) RWDevValue(rw string, m device.Dict) (ret device.Dict, err error) {
	serMutex := device.Mutex[d.Commif]
	serMutex.Lock()
	defer serMutex.Unlock()
	defer func() {
		if driveErr := recover(); driveErr != nil {
			log.Errorf("drive programer  error : (%s)", driveErr)
			errstr := fmt.Sprintf("drive programer  error : %s", driveErr)
			err = errors.New(errstr)
		}
	}()
	serconfig := serial.Config{}
	serconfig.Address = device.Commif[d.Commif]
	serconfig.BaudRate = 2400 //d.BaudRate
	serconfig.DataBits = 8    //d.DataBits
	serconfig.Parity = "E"    //d.Parity
	serconfig.StopBits = 1    // d.StopBits
	serconfig.Timeout = time.Microsecond * time.Duration(500000000/serconfig.BaudRate)
	if d.Devtype == "HZJD-B" {
		serconfig.BaudRate = 1200 //d.BaudRate
	}
	ret = map[string]interface{}{}
	ret[device.DevID] = d.Devid
	rsport, err := serial.Open(&serconfig)
	if err != nil {
		log.Errorf("open serial error %s", err.Error())
		return nil, err
	}
	defer rsport.Close()
	if rw == "r" {
		err = d.rdvalue(rsport, 1, ret)
		if err != nil {
			log.Errorf("read device %s-%s faild %s", d.Devtype, d.Devid, err.Error())
			return nil, err
		}
	}
	log.Debugln(ret)
	return ret, nil
}

/***************************************读写接口实现**************************************************/
