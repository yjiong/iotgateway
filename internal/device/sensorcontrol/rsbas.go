package sensorcontrol

import (
	"encoding/json"
	"github.com/yjiong/iotgateway/internal/device"
	//"errors"
	"fmt"
	"strconv"
	"time"

	//	"sync"
	simplejson "github.com/bitly/go-simplejson"
	log "github.com/sirupsen/logrus"
	"github.com/yjiong/iotgateway/serial"
)

//RSBAS ..
type RSBAS struct {
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
	device.RegDevice["RSBAS"] = &RSBAS{}
}

//NewDev ..
func (d *RSBAS) NewDev(id string, ele map[string]string) (device.Devicerwer, error) {
	ndev := new(RSBAS)
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
func (d *RSBAS) GetElement() (device.Dict, error) {
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

//HelpDoc ..
func (d *RSBAS) HelpDoc() interface{} {
	conn := device.Dict{
		device.DevAddr: "设备地址",
		/***********RSBAS设备的参数*****************************/
		/***********读取设备的参数*****************************/
	}
	rParameter := device.Dict{
		device.DevID: "被读取设备对象的id",
		/***********读取设备的参数*****************************/
		/***********读取设备的参数*****************************/
	}
	data := device.Dict{
		device.DevID:   "添加设备对象的id",
		device.DevType: "RSBAS", //设备类型
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
	helpdoc := device.Dict{
		"1.添加设备": devUpdate,
		"2.读取设备": readdev,
	}
	return helpdoc
}

/***********************设备的参数说明帮助***********************************/

/***************************************添加设备参数检验**********************************************/

//CheckKey ..
func (d *RSBAS) CheckKey(ele device.Dict) (bool, error) {
	return true, nil
}

/***************************************添加设备参数检验**********************************************/

/***************************************读写接口实现**************************************************/
func (d *RSBAS) readCmd(taddr byte) []byte {
	sum := (0x81 + int(taddr)) & 0xff
	cmd := []byte{0xa5, 0x81, taddr, 0x00, 0x00, device.IntToBytes(sum)[3], 0x5a}
	return cmd
}

func (d *RSBAS) rDataSum(data []byte) byte {
	sum := 0
	for i := 1; i < 9; i++ {
		sum += int(data[i])
	}
	return device.IntToBytes(sum & 0xff)[3]
}

//RWDevValue ..
func (d *RSBAS) RWDevValue(rw string, m device.Dict) (ret device.Dict, err error) {
	//log.SetLevel(log.DebugLevel)
	serMutex := device.Mutex[d.Commif]
	serMutex.Lock()
	defer serMutex.Unlock()
	serconfig := serial.Config{}
	serconfig.Address = device.Commif[d.Commif]
	serconfig.BaudRate = 9600 //d.BaudRate
	serconfig.DataBits = 8    //d.DataBits
	serconfig.Parity = "N"    //d.Parity
	serconfig.StopBits = 1    // d.StopBits
	slaveid, _ := strconv.Atoi(d.Devaddr)
	taddr := byte(slaveid)
	serconfig.Timeout = 2 * time.Second
	ret = map[string]interface{}{}
	ret[device.DevID] = d.Devid
	rsport, err := serial.Open(&serconfig)
	if err != nil {
		log.Errorf("open serial error %s", err.Error())
		return nil, err
	}
	defer rsport.Close()

	if rw == "r" {
		results := make([]byte, 11)
		for i := 0; i < 2; i++ {
			log.Debugf("send cmd = %x", d.readCmd(taddr))
			if _, ok := rsport.Write(d.readCmd(taddr)); ok != nil {
				log.Errorf("send cmd error  %s", ok.Error())
				return nil, ok
			}
			time.Sleep(100 * time.Millisecond)
			var len int
			len, err = rsport.Read(results)
			//if err != nil {
			//log.Errorf("serial read error  %s", err.Error())
			//}
			if err == nil && len == 11 && results[9] == d.rDataSum(results) {
				log.Debugf("receive data = %x len = %d sum = %x", results, len, d.rDataSum(results))
				break
			}
			if i < 1 {
				time.Sleep(10 * time.Second)
			}
		}
		if err == nil {
			var value = make(map[bool]string)
			D1 := results[3]
			D2 := results[4]
			D3 := results[5]
			D4 := results[6]
			//D5 := results[7]
			D6 := results[8]
			ret["原始报文"] = fmt.Sprintf("% x", results)
			ret["当前轿厢所处层站"] = fmt.Sprintf("%d", int(D1))
			value = map[bool]string{true: "是", false: "否"}
			ret["上行"] = value[D2&0x01 > 0]
			ret["下行"] = value[D2&0x02 > 0]
			ret["运行中"] = value[D2&0x04 > 0]
			ret["检修"] = value[D2&0x08 > 0]
			if d.Devtype == "RSBAS" {
				ret["电梯故障"] = value[D2&0x10 == 0]
				ret["自发电运行"] = value[D3&0x10 > 0]
				ret["发生A2级故障"] = value[D4&0x40 > 0]
				ret["发生A1级故障"] = value[D4&0x80 > 0]

				ret["安全装置正常"] = value[D4&0x02 > 0]
				ret["自救运行"] = value[D4&0x20 > 0]
				ret["厅门门锁关闭"] = value[D6&0x01 > 0]
				ret["抱闸打开"] = value[D6&0x02 > 0]
				ret["光电保护动作"] = value[D6&0x08 > 0]
			} else if d.Devtype == "AKS_LIFT" {
				ret["电梯故障"] = value[D2&0x10 > 0]
				ret["自发电运行"] = value[D3&0x10 > 0]
				//ret["故障"] = value[D4&0x40 > 0 || D4&0x80 > 0]
			}
			ret["泊梯"] = value[D2&0x20 > 0]
			ret["消防专用"] = value[D2&0x40 > 0]
			ret["消防返回"] = value[D2&0x80 > 0]
			ret["并联正常"] = value[D3&0x01 > 0]
			ret["群管理正常"] = value[D3&0x02 > 0]
			ret["电源正常"] = value[D3&0x04 > 0]
			ret["轿门门锁关闭"] = value[D3&0x08 > 0]
			ret["电梯到达"] = value[D3&0x20 > 0]
			ret["电梯开门"] = value[D3&0x40 > 0]
			ret["电梯关门"] = value[D3&0x80 > 0]
			ret["地震运行"] = value[D4&0x01 > 0]
			ret["专用运行"] = value[D4&0x04 > 0]
			ret["火灾管制运行"] = value[D4&0x08 > 0]
			ret["位于门区"] = value[D4&0x10 > 0]
			ret["安全触板动作"] = value[D6&0x04 > 0]
		} else {
			log.Errorf("read RSBAS faild %s", err.Error())
			return nil, err
		}
		jsret, _ := json.Marshal(ret)
		inforet, _ := simplejson.NewJson(jsret)
		pinforet, _ := inforet.EncodePretty()
		log.Debugln(string(pinforet))
	}
	return ret, nil
}

/***************************************读写接口实现**************************************************/
