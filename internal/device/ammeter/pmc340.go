package ammeter

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/yjiong/iotgateway/internal/device"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/yjiong/iotgateway/serial"
)

//PMC340 ..
type PMC340 struct {
	//组合Device
	device.Device `json:"device"`
	/**************按不同设备自定义*************************/
	BaudRate   int    `json:"baud_rate"`
	DataBits   int    `json:"data_bits"`
	StopBits   int    `json:"stop_bits"`
	Parity     string `json:"parity"`
	Buycount   []byte `json:"buycount"`
	CtrCode    byte   `json:"wcmd"`
	Ldate      []byte
	SafetyStat byte
	Phase      bool
	/**************按不同设备自定义*************************/
}

func init() {
	//device.RegDevice["PMC340"] = &PMC340{}
	//device.RegDevice["PMC320"] = &PMC340{}
	//device.RegDevice["PMC220"] = &PMC340{}
	//device.RegDevice["DDZY719"] = &PMC340{}
	//device.RegDevice["DDZY83-Z"] = &PMC340{}
	//device.RegDevice["DTZY83-Z"] = &PMC340{}
	//device.RegDevice["WS-DDSF395"] = &PMC340{}
	device.RegDevice["DLT645-2007-S"] = &PMC340{}
	device.RegDevice["DLT645-2007"] = &PMC340{}
}

var ridmap = map[string][]byte{
	"正向有功总电能":    []byte{0x00, 0x01, 0x00, 0x00},
	"组合有功电能数据块":  []byte{0x00, 0x00, 0xff, 0x00},
	"组合无功1电能数据块": []byte{0x00, 0x03, 0xff, 0x00},
	"组合无功2电能数据块": []byte{0x00, 0x04, 0xff, 0x00},
	"正向有功电能数据块":  []byte{0x00, 0x01, 0xff, 0x00},
	"反向有功电能数据块":  []byte{0x00, 0x02, 0xff, 0x00},
	"电压数据块":      []byte{0x02, 0x01, 0xFF, 0x00},
	"电流数据块":      []byte{0x02, 0x02, 0xFF, 0x00},
	"功率因素数据块":    []byte{0x02, 0x06, 0xFF, 0x00},
	"瞬时有功功率数据块":  []byte{0x02, 0x03, 0xFF, 0x00},
	"瞬时无功功率数据块":  []byte{0x02, 0x04, 0xFF, 0x00},

	"状态字数据块(3)":   []byte{0x04, 0x00, 0x05, 0x3},
	"状态字数据块(1-7)": []byte{0x04, 0x00, 0x05, 0xff},
	"剩余金额":        []byte{0x00, 0x90, 0x02, 0x0},
	"透支金额":        []byte{0x00, 0x90, 0x02, 0x1},
	"剩余免费金额":      []byte{0x00, 0x90, 0x02, 0x20},
	"费率1电价":       []byte{0x04, 0x05, 0x01, 0x01},
	"费率2电价":       []byte{0x04, 0x05, 0x01, 0x02},
	"费率3电价":       []byte{0x04, 0x05, 0x01, 0x03},
	"费率4电价":       []byte{0x04, 0x05, 0x01, 0x04},
	"年月日":         []byte{0x04, 0x00, 0x01, 0x01},
	"时分秒":         []byte{0x04, 0x00, 0x01, 0x02},
	"两套费率单价切换时间":  []byte{0x04, 0x00, 0x01, 0x08},
	"报警金额1限值":     []byte{0x04, 0x00, 0x10, 0x01},
	"报警金额2限值":     []byte{0x04, 0x00, 0x10, 0x02},
	"透支金额限值":      []byte{0x04, 0x00, 0x10, 0x03},
	"囤积金额限值":      []byte{0x04, 0x00, 0x10, 0x04},
	"合闸允许金额限值":    []byte{0x04, 0x00, 0x10, 0x05},

	"购电次数": []byte{0x03, 0x33, 0x02, 0x01},
	"退费记录": []byte{0x03, 0x34, 0x00, 0x01},
	"客户编码": []byte{0x04, 0x00, 0x04, 0x0E},
	"软件配置": []byte{0x04, 0x9F, 0x00, 0x00},

	"退费":        []byte{0x04, 0x00, 0x10, 0x06},
	"购电金额及购电次数": []byte{0x04, 0x90, 0x00, 0x00},
	"本月用电账单":    []byte{0x03, 0x90, 0x00, 0x01},
}

// NewDev ..
func (d *PMC340) NewDev(id string, ele map[string]string) (device.Devicerwer, error) {
	ndev := new(PMC340)
	ndev.Device = d.Device.NewDev(id, ele)
	/***********************初始化设备的特有的参数*****************************/
	if baud, ok := ele["BaudRate"]; ok {
		ndev.BaudRate, _ = strconv.Atoi(baud)
	} else {
		ndev.BaudRate = 2400
	}
	if datab, ok := ele["DataBits"]; ok {
		ndev.DataBits, _ = strconv.Atoi(datab)
	} else {
		ndev.DataBits = 8
	}
	if stopb, ok := ele["StopBits"]; ok {
		ndev.StopBits, _ = strconv.Atoi(stopb)
	} else {
		ndev.StopBits = 1
	}
	if parity, ok := ele["Parity"]; ok {
		ndev.Parity = parity
	} else {
		ndev.Parity = "E"
	}
	/***********************初始化设备的特有的参数*****************************/
	if ndev.Devtype == "DDZY83-Z" ||
		ndev.Devtype == "DLT645-2007-S" ||
		ndev.Devtype == "PMC220" {
		ndev.Phase = true
	}
	return ndev, nil
}

// GetElement ..
func (d *PMC340) GetElement() (device.Dict, error) {
	conn := device.Dict{
		/***********************设备的特有的参数*****************************/
		device.DevAddr: d.Devaddr,
		"commif":       d.Commif,
		"BaudRate":     strconv.Itoa(d.BaudRate),
		"DataBits":     strconv.Itoa(d.DataBits),
		"StopBits":     strconv.Itoa(d.StopBits),
		"Parity":       d.Parity,
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
func (d *PMC340) HelpDoc() interface{} {
	conn := device.Dict{
		device.DevAddr: "设备地址",
		/***********设备的参数*****************************/
		"commif":   "通信接口,比如(rs485-1)",
		"BaudRate": "波特率,比如(9600)",
		"DataBits": "数据位,比如(8)",
		"Parity":   "校验,(N,E,O)",
		"StopBits": "停止位,比如(1)",
		/***********设备的参数*****************************/
	}
	rParameter := device.Dict{
		device.DevID: "被读取设备对象的id",
		/***********读取设备的参数*****************************/
		/***********读取设备的参数*****************************/
	}

	wParameter := device.Dict{
		device.DevID: "被操作设备对象的id",
		/***********操作设备的参数*****************************/
	}
	data := device.Dict{
		device.DevID:   "添加设备对象的id",
		device.DevType: "DLT645-2007|DLT645-2007-S",
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
func (d *PMC340) CheckKey(ele device.Dict) (bool, error) {
	//if _, ok := ele["BaudRate"].(json.Number); !ok {
	//if _, ok := ele["BaudRate"].(string); !ok {
	//return false, fmt.Errorf("device must have  element 波特率 :baudrate")

	//}
	//}
	//if _, ok := ele["DataBits"].(json.Number); !ok {
	//if _, ok := ele["DataBits"].(string); !ok {
	//return false, fmt.Errorf("device must have  element 数据位 :databits")
	//}
	//}
	//if _, ok := ele["StopBits"].(json.Number); !ok {
	//if _, ok := ele["StopBits"].(string); !ok {
	//return false, fmt.Errorf("device must have  element 停止位 :stopbits")
	//}
	//}
	//if _, ok := ele["Parity"].(string); !ok {
	//return false, fmt.Errorf("device must have  element 校验 :parity")
	//}
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
func (d *PMC340) plus33(data []byte) []byte {
	rd := make([]byte, len(data))
	for i := range data {
		if data[i] < 0xCD {
			rd[len(data)-i-1] = (data[i] + 0x33)
		} else {
			rd[len(data)-i-1] = (data[i] - 0xCD)
		}
	}
	return rd
}

func (d *PMC340) sub33(data []byte) []byte {
	rd := make([]byte, len(data))
	for i := range data {
		if data[i] >= 0x33 {
			rd[len(data)-i-1] = (data[i] - 0x33)
		} else {
			rd[len(data)-i-1] = (data[i] + 0xCD)
		}
	}
	return rd
}

/***************************************读写接口实现**************************************************/

func (d *PMC340) readCmd(taddr string, di []byte) []byte {
	saddr := fmt.Sprintf("%012s", taddr)
	cmd := []byte{0x68}
	for i := 12; i >= 2; i -= 2 {
		temp, _ := strconv.Atoi(saddr[i-2 : i])
		cmd = append(cmd, device.Bcd2Hex(device.IntToBytes(temp)[3]))
	}
	cmd = append(cmd, []byte{0x68, 0x11, 0x04}...)
	cmd = append(cmd, d.plus33(di)...)
	cmd = append(cmd, d.rDateSum(cmd))
	cmd = append(cmd, 0x16)
	rcmd := append([]byte{0xFE, 0xFE, 0xFE, 0xFE}, cmd...)
	log.Debugf("send readcmd = % x", rcmd)
	return rcmd
}

func (d *PMC340) writeCmd(taddr string, di, value []byte) []byte {
	saddr := fmt.Sprintf("%012s", taddr)
	cmd := []byte{0x68}
	for i := 12; i >= 2; i -= 2 {
		temp, _ := strconv.Atoi(saddr[i-2 : i])
		cmd = append(cmd, device.Bcd2Hex(device.IntToBytes(temp)[3]))
	}
	cmd = append(cmd, 0x68)
	cmd = append(cmd, d.CtrCode)
	dlen := device.IntToBytes(len(di))[3] + device.IntToBytes(len(value))[3] + 0x8
	cmd = append(cmd, dlen)
	cmd = append(cmd, d.plus33(di)...)
	pwCliCode := []byte{0, 0x45, 0x67, 0, 0, 0x45, 0x67, 0}
	//pwCliCode = []byte{0, 0, 0, 1, 0x88, 0x88, 0x88, 0x02}
	if d.Devtype != "DDZY83-Z" {
		cmd = append(cmd, d.plus33(pwCliCode)...)
	} else {
		cmd[9] -= 8
	}
	cmd = append(cmd, d.plus33(value)...)
	if d.CtrCode == 0x08 {
		cmd = []byte{0x68, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x68, d.CtrCode, 6}
		cmd = append(cmd, d.plus33(value)...)
	}
	cmd = append(cmd, d.rDateSum(cmd))
	cmd = append(cmd, 0x16)
	rcmd := append([]byte{0xFE, 0xFE, 0xFE, 0xFE}, cmd...)
	log.Debugf("send writecmd = % x", rcmd)
	return rcmd
}

func (d *PMC340) verify(taddr string, di []byte) []byte {
	saddr := fmt.Sprintf("%012s", taddr)
	cmd := []byte{0x68}
	for i := 12; i >= 2; i -= 2 {
		temp, _ := strconv.Atoi(saddr[i-2 : i])
		cmd = append(cmd, device.Bcd2Hex(device.IntToBytes(temp)[3]))
	}
	cmd = append(cmd, []byte{0x68, 0x03}...)
	dlen := device.IntToBytes(len(di))[3]
	cmd = append(cmd, dlen)
	cdi := d.plus33(di)
	cmd = append(cmd, cdi...)
	sum := d.rDateSum(cmd)
	cmd = append(cmd, sum)
	cmd = append(cmd, 0x16)
	rcmd := append([]byte{0xFE, 0xFE, 0xFE, 0xFE}, cmd...)
	return rcmd
}

func (d *PMC340) rDateSum(data []byte) byte {
	sum := 0
	for i := 0; i < len(data); i++ {
		sum += int(data[i])
	}
	return device.IntToBytes(sum & 0xff)[3]
}

func (d *PMC340) readhandle(rsport serial.Port, rdi []byte) (results []byte, err error) {
	results = make([]byte, 100)
	for i := 0; i < 1; i++ {
		//log.Debugf("send cmd = %x", d.readCmd(d.Devaddr, rdi))
		log.Debugf("读数据标识 % x", rdi)
		if _, ok := rsport.Write(d.readCmd(d.Devaddr, rdi)); ok != nil {
			log.Errorf("send cmd error  %s", ok.Error())
			return nil, ok
		}
		delayt := 1000000 / d.BaudRate
		time.Sleep(time.Duration(delayt) * time.Millisecond)
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
			if rlen > 9 && d.rDateSum(results[startbyte:rlen-2]) == results[rlen-2] &&
				results[startbyte] == 0x68 && results[rlen-1] == 0x16 {
				results = results[startbyte : rlen-2]
				//log.Debugf("校验正确")
				log.Debugf("receive data = % x rlen = %d sum = %x ", results[:rlen], rlen, d.rDateSum(results[startbyte:rlen-2]))
				break
			} else {
				log.Debugf("校验错误")
				err = errors.New("校验错误")
			}
		}
		if i < 1 {
			time.Sleep(300 * time.Millisecond)
		}
	}
	return
}

func (d *PMC340) writehandle(rsport serial.Port, rdi, value []byte) (results []byte, err error) {
	results = make([]byte, 100)
	for i := 0; i < 1; i++ {
		//log.Debugf("send cmd = %x", d.readCmd(d.Devaddr, rdi))
		log.Debugf("写数据标识 % x", rdi)
		if _, ok := rsport.Write(d.writeCmd(d.Devaddr, rdi, value)); ok != nil {
			log.Errorf("send write cmd error  %s", ok.Error())
			return nil, ok
		}
		//delayt := 2000000 / d.BaudRate
		//time.Sleep(time.Duration(delayt) * time.Millisecond)
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
			if rlen > 9 && d.rDateSum(results[startbyte:rlen-2]) == results[rlen-2] &&
				results[startbyte] == 0x68 && results[rlen-1] == 0x16 {
				results = results[startbyte : rlen-2]
				log.Debugf("receive data = % x rlen = %d sum = %x ", results[:rlen], rlen, d.rDateSum(results[startbyte:rlen-2]))
				//log.Debugf("校验正确")
				break
			} else {
				log.Debugf("校验错误")
				err = errors.New("校验错误")
			}
		}
		if i < 1 {
			time.Sleep(300 * time.Millisecond)
		}
	}
	return
}

func (d *PMC340) decode(ret device.Dict, di, debyte []byte) (err error) {
	if len(debyte) > 10 {
		switch {
		case reflect.DeepEqual(di, ridmap["正向有功总电能"]):
			if debyte[8] == 0x91 {
				valb := d.sub33(debyte[10:])
				log.Debugf("正向有功总电能 decode data = % x", valb)
				ret["正向有功总电能"] = float64(device.Hex2Bcd(valb[0]))*10000 +
					float64(device.Hex2Bcd(valb[1]))*100 +
					float64(device.Hex2Bcd(valb[2])) +
					float64(device.Hex2Bcd(valb[3]))/100
			} else {
				err = errors.New("报文错误")
			}
		case reflect.DeepEqual(di, ridmap["电压数据块"]) ||
			reflect.DeepEqual(di, ridmap["功率因素数据块"]):
			var strb string
			var xs float64
			if di[1] == 0x01 {
				strb = "电压"
				xs = 10
			} else {
				strb = "功率因素"
				xs = 1000
			}
			if debyte[8] == 0x91 {
				valb := d.sub33(debyte[10:])
				log.Debugf(strb+"数据块 decode data = % x", valb)
				tempint, _ := strconv.Atoi(fmt.Sprintf("%x", valb[4:6]))
				if d.Phase {
					ret[strb] = float64(tempint) / xs
				} else {
					ret["A相"+strb] = float64(tempint) / xs
					tempint, _ = strconv.Atoi(fmt.Sprintf("%x", valb[2:4]))
					ret["B相"+strb] = float64(tempint) / xs
					tempint, _ = strconv.Atoi(fmt.Sprintf("%x", valb[0:2]))
					ret["C相"+strb] = float64(tempint) / xs
					if di[1] == 0x06 {
						tempint, _ = strconv.Atoi(fmt.Sprintf("%x", valb[6:8]))
						ret["总"+strb] = float64(tempint) / xs
					}
				}
			} else {
				err = errors.New("报文错误")
			}
		case reflect.DeepEqual(di, ridmap["电流数据块"]):
			if debyte[8] == 0x91 {
				valb := d.sub33(debyte[10:])
				log.Debugf("电流数据块 decode data = % x", valb)
				tempint, _ := strconv.Atoi(fmt.Sprintf("%x", valb[6:9]))
				if d.Phase {
					ret["电流"] = float64(tempint) / 1000
				} else {
					ret["A相电流"] = float64(tempint) / 1000
					tempint, _ = strconv.Atoi(fmt.Sprintf("%x", valb[3:6]))
					ret["B相电流"] = float64(tempint) / 1000
					tempint, _ = strconv.Atoi(fmt.Sprintf("%x", valb[0:3]))
					ret["C相电流"] = float64(tempint) / 1000
				}
			} else {
				err = errors.New("报文错误")
			}
		case reflect.DeepEqual(di, ridmap["瞬时有功功率数据块"]) ||
			reflect.DeepEqual(di, ridmap["瞬时无功功率数据块"]):
			var strb string
			if di[1] == 0x03 {
				strb = "有"
			} else {
				strb = "无"
			}
			if debyte[8] == 0x91 {
				valb := d.sub33(debyte[10:])
				log.Debugf("瞬时"+strb+"功功率数据块 decode data = % x", valb)
				tempint, _ := strconv.Atoi(fmt.Sprintf("%x", valb[9:12]))
				ret["瞬时总"+strb+"功功率"] = float64(tempint) / 1000
				tempint, _ = strconv.Atoi(fmt.Sprintf("%x", valb[6:9]))
				if !d.Phase {
					ret["A相瞬时"+strb+"功功率"] = float64(tempint) / 1000
					tempint, _ = strconv.Atoi(fmt.Sprintf("%x", valb[3:6]))
					ret["B相瞬时"+strb+"功功率"] = float64(tempint) / 1000
					tempint, _ = strconv.Atoi(fmt.Sprintf("%x", valb[0:3]))
					ret["C相瞬时"+strb+"功功率"] = float64(tempint) / 1000
				}
			} else {
				err = errors.New("报文错误")
			}
		case reflect.DeepEqual(di, ridmap["组合有功电能数据块"]) ||
			reflect.DeepEqual(di, ridmap["正向有功电能数据块"]) ||
			reflect.DeepEqual(di, ridmap["反向有功电能数据块"]) ||
			reflect.DeepEqual(di, ridmap["组合无功1电能数据块"]) ||
			reflect.DeepEqual(di, ridmap["组合无功2电能数据块"]):
			var strb string
			if di[1] == 0x00 {
				strb = "组合有功"
			} else if di[1] == 0x01 {
				strb = "正向有功"
			} else if di[1] == 0x02 {
				strb = "反向有功"
			} else if di[1] == 0x03 {
				strb = "组合无功1"
			} else if di[1] == 0x04 {
				strb = "组合无功2"
			}
			if debyte[8] == 0x91 {
				valb := d.sub33(debyte[10:])
				log.Debugf(strb+"有功电能数据块 decode data = % x", valb)
				ret[strb+"费率4电能"] = float64(device.Hex2Bcd(valb[0]))*10000 +
					float64(device.Hex2Bcd(valb[1]))*100 +
					float64(device.Hex2Bcd(valb[2])) +
					float64(device.Hex2Bcd(valb[3]))/100
				ret[strb+"费率3电能"] = float64(device.Hex2Bcd(valb[4]))*10000 +
					float64(device.Hex2Bcd(valb[5]))*100 +
					float64(device.Hex2Bcd(valb[6])) +
					float64(device.Hex2Bcd(valb[7]))/100
				ret[strb+"费率2电能"] = float64(device.Hex2Bcd(valb[8]))*10000 +
					float64(device.Hex2Bcd(valb[9]))*100 +
					float64(device.Hex2Bcd(valb[10])) +
					float64(device.Hex2Bcd(valb[11]))/100
				ret[strb+"费率1电能"] = float64(device.Hex2Bcd(valb[12]))*10000 +
					float64(device.Hex2Bcd(valb[13]))*100 +
					float64(device.Hex2Bcd(valb[14])) +
					float64(device.Hex2Bcd(valb[15]))/100
				ret[strb+"总电能"] = float64(device.Hex2Bcd(valb[16]))*10000 +
					float64(device.Hex2Bcd(valb[17]))*100 +
					float64(device.Hex2Bcd(valb[18])) +
					float64(device.Hex2Bcd(valb[19]))/100
			} else {
				err = errors.New("报文错误")
			}

		case reflect.DeepEqual(di, ridmap["本月用电账单"]):
			if debyte[8] == 0x91 {
				valb := d.sub33(debyte[10:])
				log.Debugf("本月用电账单 decode data = % x", valb)
				ret["当前时间"] = fmt.Sprintf("%x年%x月%x日%x时%x分", valb[44], valb[45], valb[46], valb[47], valb[47])
				ret["本月总电费"] = float64(device.Hex2Bcd(valb[40]))*10000 +
					float64(device.Hex2Bcd(valb[41]))*100 +
					float64(device.Hex2Bcd(valb[42])) +
					float64(device.Hex2Bcd(valb[43]))/100
				ret["本月正向有功总电能"] = float64(device.Hex2Bcd(valb[36]))*10000 +
					float64(device.Hex2Bcd(valb[37]))*100 +
					float64(device.Hex2Bcd(valb[38])) +
					float64(device.Hex2Bcd(valb[39]))/100
				ret["本月费率1正向有功电能"] = float64(device.Hex2Bcd(valb[32]))*10000 +
					float64(device.Hex2Bcd(valb[33]))*100 +
					float64(device.Hex2Bcd(valb[34])) +
					float64(device.Hex2Bcd(valb[35]))/100
				ret["本月费率2正向有功电能"] = float64(device.Hex2Bcd(valb[28]))*10000 +
					float64(device.Hex2Bcd(valb[29]))*100 +
					float64(device.Hex2Bcd(valb[30])) +
					float64(device.Hex2Bcd(valb[31]))/100
				ret["本月费率3正向有功电能"] = float64(device.Hex2Bcd(valb[24]))*10000 +
					float64(device.Hex2Bcd(valb[25]))*100 +
					float64(device.Hex2Bcd(valb[26])) +
					float64(device.Hex2Bcd(valb[27]))/100
				ret["本月费率4正向有功电能"] = float64(device.Hex2Bcd(valb[20]))*10000 +
					float64(device.Hex2Bcd(valb[21]))*100 +
					float64(device.Hex2Bcd(valb[22])) +
					float64(device.Hex2Bcd(valb[23]))/100
				ret["本月反向有功总电能"] = float64(device.Hex2Bcd(valb[16]))*10000 +
					float64(device.Hex2Bcd(valb[17]))*100 +
					float64(device.Hex2Bcd(valb[18])) +
					float64(device.Hex2Bcd(valb[19]))/100
				ret["本月费率1反向有功电能"] = float64(device.Hex2Bcd(valb[12]))*10000 +
					float64(device.Hex2Bcd(valb[13]))*100 +
					float64(device.Hex2Bcd(valb[14])) +
					float64(device.Hex2Bcd(valb[15]))/100
				ret["本月费率2反向有功电能"] = float64(device.Hex2Bcd(valb[8]))*10000 +
					float64(device.Hex2Bcd(valb[9]))*100 +
					float64(device.Hex2Bcd(valb[11])) +
					float64(device.Hex2Bcd(valb[12]))/100
				ret["本月费率3反向有功电能"] = float64(device.Hex2Bcd(valb[4]))*10000 +
					float64(device.Hex2Bcd(valb[5]))*100 +
					float64(device.Hex2Bcd(valb[6])) +
					float64(device.Hex2Bcd(valb[7]))/100
				ret["本月费率4反向有功电能"] = float64(device.Hex2Bcd(valb[0]))*10000 +
					float64(device.Hex2Bcd(valb[1]))*100 +
					float64(device.Hex2Bcd(valb[2])) +
					float64(device.Hex2Bcd(valb[3]))/100
			} else {
				err = errors.New("报文错误")
			}

		case reflect.DeepEqual(di, ridmap["状态字数据块(3)"]):
			if debyte[8] == 0x91 {
				valb := d.sub33(debyte[10:])
				log.Debugf("状态字数据块(3) decode data = % x", valb)
				warning := map[byte]string{
					0: "无",
					1: "有",
				}
				status := map[byte]string{
					0: "通",
					1: "断",
				}
				baodian := map[byte]string{
					0: "非保电",
					1: "保电",
				}
				ret["保电状态"] = baodian[valb[0]&0x10/0x10]
				ret["继电器命令状态"] = status[valb[1]&0x40/0x40]
				ret["继电器状态"] = status[valb[1]&0x10/0x10]
				ret["跳闸报警状态"] = warning[valb[1]&0x80/0x80]
			} else {
				err = errors.New("报文错误")
			}

		case reflect.DeepEqual(di, ridmap["剩余金额"]):
			if debyte[8] == 0x91 {
				valb := d.sub33(debyte[10:])
				log.Debugf("剩余金额 decode data = % x", valb)
				ret["剩余金额"] = float64(device.Hex2Bcd(valb[0]))*10000 +
					float64(device.Hex2Bcd(valb[1]))*100 +
					float64(device.Hex2Bcd(valb[2])) +
					float64(device.Hex2Bcd(valb[3]))/100
			} else {
				err = errors.New("报文错误")
			}

		case reflect.DeepEqual(di, ridmap["透支金额"]):
			if debyte[8] == 0x91 {
				valb := d.sub33(debyte[10:])
				log.Debugf("decode data = % x", valb)
				ret["透支金额"] = float64(device.Hex2Bcd(valb[0]))*10000 +
					float64(device.Hex2Bcd(valb[1]))*100 +
					float64(device.Hex2Bcd(valb[2])) +
					float64(device.Hex2Bcd(valb[3]))/100
			} else {
				err = errors.New("报文错误")
			}

		case reflect.DeepEqual(di, ridmap["剩余免费金额"]):
			if debyte[8] == 0x91 {
				valb := d.sub33(debyte[10:])
				log.Debugf("剩余免费金额 decode data = % x", valb)
				ret["剩余免费金额"] = float64(device.Hex2Bcd(valb[0]))*10000 +
					float64(device.Hex2Bcd(valb[1]))*100 +
					float64(device.Hex2Bcd(valb[2])) +
					float64(device.Hex2Bcd(valb[3]))/100
			} else {
				err = errors.New("报文错误")
			}

		case reflect.DeepEqual(di, ridmap["费率1电价"]) ||
			reflect.DeepEqual(di, ridmap["费率2电价"]) ||
			reflect.DeepEqual(di, ridmap["费率3电价"]) ||
			reflect.DeepEqual(di, ridmap["费率4电价"]):
			strb := fmt.Sprintf("%d", di[3])
			if debyte[8] == 0x91 {
				valb := d.sub33(debyte[10:])
				log.Debugf("费率1电价 decode data = % x", valb)
				ret["费率"+strb+"电价"] = float64(device.Hex2Bcd(valb[0]))*100 +
					float64(device.Hex2Bcd(valb[1])) +
					float64(device.Hex2Bcd(valb[2]))/100 +
					float64(device.Hex2Bcd(valb[3]))/10000
			} else {
				err = errors.New("报文错误")
			}

		case reflect.DeepEqual(di, ridmap["两套费率单价切换时间"]):
			if debyte[8] == 0x91 {
				valb := d.sub33(debyte[10:])
				log.Debugf("两套费率单价切换时间 decode data = % x", valb)
				ret["两套费率单价切换时间"] = fmt.Sprintf("%x年%x月%x日%x时%x分", valb[0], valb[1], valb[2], valb[3], valb[4])
			} else {
				err = errors.New("报文错误")
			}

		case reflect.DeepEqual(di, ridmap["报警金额1限值"]):
			if debyte[8] == 0x91 {
				valb := d.sub33(debyte[10:])
				log.Debugf("报警金额1限值 decode data = % x", valb)
				ret["报警金额1限值"] = float64(device.Hex2Bcd(valb[0]))*10000 +
					float64(device.Hex2Bcd(valb[1]))*100 +
					float64(device.Hex2Bcd(valb[2])) +
					float64(device.Hex2Bcd(valb[3]))/100
			} else {
				err = errors.New("报文错误")
			}

		case reflect.DeepEqual(di, ridmap["报警金额2限值"]):
			if debyte[8] == 0x91 {
				valb := d.sub33(debyte[10:])
				log.Debugf("报警金额2限值 decode data = % x", valb)
				ret["报警金额2限值"] = float64(device.Hex2Bcd(valb[0]))*10000 +
					float64(device.Hex2Bcd(valb[1]))*100 +
					float64(device.Hex2Bcd(valb[2])) +
					float64(device.Hex2Bcd(valb[3]))/100
			} else {
				err = errors.New("报文错误")
			}

		case reflect.DeepEqual(di, ridmap["透支金额限值"]):
			if debyte[8] == 0x91 {
				valb := d.sub33(debyte[10:])
				log.Debugf("透支金额限值 decode data = % x", valb)
				ret["透支金额限值"] = float64(device.Hex2Bcd(valb[0]))*10000 +
					float64(device.Hex2Bcd(valb[1]))*100 +
					float64(device.Hex2Bcd(valb[2])) +
					float64(device.Hex2Bcd(valb[3]))/100
			} else {
				err = errors.New("报文错误")
			}

		case reflect.DeepEqual(di, ridmap["囤积金额限值"]):
			if debyte[8] == 0x91 {
				valb := d.sub33(debyte[10:])
				log.Debugf("囤积金额限值 decode data = % x", valb)
				ret["囤积金额限值"] = float64(device.Hex2Bcd(valb[0]))*10000 +
					float64(device.Hex2Bcd(valb[1]))*100 +
					float64(device.Hex2Bcd(valb[2])) +
					float64(device.Hex2Bcd(valb[3]))/100
			} else {
				err = errors.New("报文错误")
			}

		case reflect.DeepEqual(di, ridmap["合闸允许金额限值"]):
			if debyte[8] == 0x91 {
				valb := d.sub33(debyte[10:])
				log.Debugf("合闸允许金额限值 decode data = % x", valb)
				ret["合闸允许金额限值"] = float64(device.Hex2Bcd(valb[0]))*10000 +
					float64(device.Hex2Bcd(valb[1]))*100 +
					float64(device.Hex2Bcd(valb[2])) +
					float64(device.Hex2Bcd(valb[3]))/100
			} else {
				err = errors.New("报文错误")
			}

		case reflect.DeepEqual(di, ridmap["状态字数据块(1-7)"]):
			if debyte[8] == 0x91 {
				valb := d.sub33(debyte[10:])
				log.Debugf("状态字数据块 decode data = % x", valb)
				warning := map[byte]string{
					0: "无",
					1: "有",
				}
				status := map[byte]string{
					0: "通",
					1: "断",
				}
				ret["继电器命令状态"] = status[valb[9]&0x40/0x40]
				ret["继电器状态"] = status[valb[9]&0x10/0x10]
				ret["跳闸报警状态"] = warning[valb[9]&0x80/0x80]
			} else {
				err = errors.New("报文错误")
			}

		case reflect.DeepEqual(di, ridmap["购电次数"]):
			if debyte[8] == 0x91 {
				valb := d.sub33(debyte[10:])
				d.Buycount = valb[0:2]
				log.Debugf("购电次数 decode data = % x", d.Buycount)

			} else {
				err = errors.New("报文错误")
			}

		case reflect.DeepEqual(di, ridmap["年月日"]) ||
			reflect.DeepEqual(di, ridmap["时分秒"]):
			if debyte[8] == 0x91 {
				valb := d.sub33(debyte[10:])
				if di[3] == 2 {
					log.Debugf("时分秒 decode data = % x", valb)
					ret["表时钟"] = fmt.Sprintf("%x年%x月%x日%x时%x分%x秒", d.Ldate[0], d.Ldate[1], d.Ldate[2], valb[0], valb[1], valb[2])
				} else {
					d.Ldate = valb[0:3]
				}

			} else {
				err = errors.New("报文错误")
			}

		case reflect.DeepEqual(di, ridmap["退费记录"]):
			if debyte[8] == 0x91 {
				valb := d.sub33(debyte[10:])
				log.Debugf("退费记录 decode data = % x", valb)

			} else {
				err = errors.New("报文错误")
			}

		case reflect.DeepEqual(di, ridmap["客户编码"]):
			if debyte[8] == 0x91 {
				valb := d.sub33(debyte[10:])
				log.Debugf("客户编码 decode data = % x", valb)

			} else {
				err = errors.New("报文错误")
			}

		case reflect.DeepEqual(di, ridmap["软件配置"]):
			if debyte[8] == 0x91 {
				valb := d.sub33(debyte[10:])
				d.SafetyStat = valb[1]
				log.Debugf("软件配置 decode data = % x", valb)

			} else {
				err = errors.New("报文错误")
			}

		default:
			valb := d.sub33(debyte[10:])
			log.Debugf("decode data = % x", valb)
			err = errors.New("报文解析错误")
		}
	} else {
		err = errors.New("报文长度错误")
	}
	return
}

func (d *PMC340) readValue(rsport serial.Port, rdi []byte, ret device.Dict) (err error) {
	if results, inerr := d.readhandle(rsport, rdi); inerr == nil {
		err = d.decode(ret, rdi, results)
	} else {
		err = inerr
	}
	return
}

func (d *PMC340) wencode(k string, v interface{}) (rval, di []byte, err error) {
	//d.CtrCode = 0x1c //控制
	//d.CtrCode = 0x14 //写值
	patternPrice := regexp.MustCompile(`费率[1,2,3,4]电价`)
	patternWarning := regexp.MustCompile(`报警金额[1,2]限值|透支金额限值|囤积金额限值|合闸允许金额限值|退费`)
	switch {
	case k == "软件配置":
		d.CtrCode = 0x14
		di = ridmap[k]
		rval = []byte{0, 4} //退出安全认证,投入{0,6}
		log.Debugf("写%s  value = % x", k, rval)
		return

	case k == "广播校时":
		dtpattern := regexp.MustCompile(`^(\d{2}\W){5}\d{2}`)
		dt, _ := v.(string)
		if dtpattern.MatchString(dt) {
			for _, symbol := range []string{"-", " ", ":", "/"} {
				dt = strings.Replace(dt, symbol, ",", -1)
			}
			dts := strings.Split(dt, ",")
			for i := 0; i < 6; i++ {
				di, _ := strconv.Atoi(dts[i])
				rval = append(rval, device.Bcd2Hex(device.IntToBytes(di)[3]))
			}
			log.Debugf("写%s  value = % x", k, rval)
			//dt = dtpattern.ReplaceAllString(dt, "$1$2$3$4$5$6")
			d.CtrCode = 0x8
			return
		}
		err = errors.New("参数格式错误")

	case k == "控制":
		d.CtrCode = 0x1c //控制
		motion := map[string]byte{
			"报警":   0x2A,
			"报警解除": 0x2B,
			"跳闸":   0x1A,
			"合闸":   0x1C,
			"保电":   0x3A,
			"保电解除": 0x3B,
			"合闸允许": 0x1B,
		}
		if m, ok := v.(string); ok {
			for km, vm := range motion {
				if km == m {
					rval = append(d.Ldate, []byte{0x23, 0x59, 0x59, 0x00, vm}...)
					if d.Devtype == "DDZY83-Z" {
						rval = []byte{0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0, vm, 0, 0, 0, 1, 0, 0, 0, 2}
					}
					log.Debugf("写%s %s value = % x", k, m, rval)
					return
				}
			}
		}
		err = errors.New("控制命令错误")

	case patternPrice.MatchString(k):
		d.CtrCode = 0x14 //写值
		if m, ok := v.(json.Number); ok {
			var fvm float64
			if fvm, err = m.Float64(); err == nil {
				im := int(fvm * 10000)
				price := make([]byte, 4)
				price[0] = device.Bcd2Hex(device.IntToBytes(int(im / 1000000))[3])
				price[1] = device.Bcd2Hex(device.IntToBytes(int(im % 1000000 / 10000))[3])
				price[2] = device.Bcd2Hex(device.IntToBytes(int(im % 10000 / 100))[3])
				price[3] = device.Bcd2Hex(device.IntToBytes(int(im % 100))[3])
				di = ridmap[k]
				rval = price
				log.Debugf("写%s  value = % x", k, rval)
			}
			return
		}

	case patternWarning.MatchString(k):
		d.CtrCode = 0x14
		if m, ok := v.(json.Number); ok {
			var fvm float64
			if fvm, err = m.Float64(); err == nil {
				im := int(fvm * 100)
				price := make([]byte, 4)
				price[0] = device.Bcd2Hex(device.IntToBytes(int(im / 1000000))[3])
				price[1] = device.Bcd2Hex(device.IntToBytes(int(im % 1000000 / 10000))[3])
				price[2] = device.Bcd2Hex(device.IntToBytes(int(im % 10000 / 100))[3])
				price[3] = device.Bcd2Hex(device.IntToBytes(int(im % 100))[3])
				di = ridmap[k]
				rval = price
				log.Debugf("写%s  value = % x", k, rval)
			}
			return
		}

	case k == "充值":
		d.CtrCode = 0x14 //写值
		//d.CtrCode = 0x3 //写值
		di = ridmap["购电金额及购电次数"]
		//di = []byte{0x07, 0x01, 0x02, 0xFF} //充值
		//rval = []byte{0x00, 0x10, 0x01, 0x72, 0x81, 0x54, 0, 0, 0, 1, 0, 0, 0, 1}
		if m, ok := v.(json.Number); ok {
			var fvm float64
			if fvm, err = m.Float64(); err == nil {
				im := int(fvm * 100)
				money := make([]byte, 8)
				money[0] = device.Bcd2Hex(device.IntToBytes(int(im / 1000000))[3])
				money[1] = device.Bcd2Hex(device.IntToBytes(int(im % 1000000 / 10000))[3])
				money[2] = device.Bcd2Hex(device.IntToBytes(int(im % 10000 / 100))[3])
				money[3] = device.Bcd2Hex(device.IntToBytes(int(im % 100))[3])
				//log.Debugf("% x", device.Hex2Bcd(d.Buycount[1]))
				buyc := int(device.Hex2Bcd(d.Buycount[0]))*100 + int(device.Hex2Bcd(d.Buycount[1]))
				log.Debugf("buyc=%d", buyc)
				money[6] = device.Bcd2Hex(device.IntToBytes(int(buyc+1) / 100)[3])
				money[7] = device.Bcd2Hex(device.IntToBytes(int(buyc+1) % 100)[3])
				rval = money
				log.Debugf("写%s  value = % x", k, rval)
			}
			return
		}

	default:
		err = errors.New("错误的命令参数")
	}

	return
}

func (d *PMC340) writeValue(rsport serial.Port, k string, v interface{}) (err error) {
	var val []byte
	var di []byte
	if val, di, err = d.wencode(k, v); err == nil {
		if results, inerr := d.writehandle(rsport, di, val); inerr == nil {
			log.Debugf("write results byte = % x", results)
			if results[8]&0xf0 != 0x90 {
				err = errors.New("命令执行失败")
				return
			}
		} else {
			err = inerr
		}
	}
	return
}

// RWDevValue ..
func (d *PMC340) RWDevValue(rw string, m device.Dict) (ret device.Dict, err error) {
	var dilist [][]byte
	if d.Devtype == "PMC220" {
		dilist = [][]byte{
			ridmap["正向有功总电能"],
			//ridmap["正向有功电能数据块"],
		}
	} else if d.Devtype == "DDZY83-Z" ||
		d.Devtype == "DTZY83-Z" ||
		d.Devtype == "DLT645-2007" ||
		d.Devtype == "DLT645-2007-S" {
		dilist = [][]byte{
			ridmap["年月日"],
			ridmap["时分秒"],
			ridmap["组合有功电能数据块"],
			ridmap["正向有功电能数据块"],
			ridmap["反向有功电能数据块"],
			ridmap["组合无功1电能数据块"],
			ridmap["组合无功2电能数据块"],
			ridmap["电压数据块"],
			ridmap["电流数据块"],
			ridmap["瞬时有功功率数据块"],
			ridmap["瞬时无功功率数据块"],
			ridmap["功率因素数据块"],
			ridmap["状态字数据块(3)"],
		}
	} else if d.Devtype == "DDZY719" {
		dilist = [][]byte{
			ridmap["组合有功电能数据块"],
			ridmap["正向有功电能数据块"],
			ridmap["反向有功电能数据块"],
			ridmap["电压数据块"],
			ridmap["电流数据块"],
			ridmap["瞬时有功功率数据块"],
			//ridmap["状态字数据块(3)"],
		}
	} else if d.Devtype == "WS-DDSF395" {
		dilist = [][]byte{
			ridmap["正向有功总电能"],
			ridmap["反向有功电能数据块"],
			ridmap["组合有功电能数据块"],
			//ridmap["状态字数据块(3)"],
		}
	} else {
		dilist = [][]byte{
			ridmap["正向有功电能数据块"],
			ridmap["状态字数据块(3)"],
			ridmap["剩余金额"],
			ridmap["透支金额"],
		}
	}

	extralist1 := [][]byte{
		ridmap["剩余免费金额"],
		ridmap["费率1电价"],
		ridmap["费率2电价"],
		ridmap["费率3电价"],
		ridmap["费率4电价"],
		ridmap["报警金额1限值"],
		ridmap["报警金额2限值"],
		ridmap["透支金额限值"],
		ridmap["囤积金额限值"],
		ridmap["合闸允许金额限值"],
	}
	extralist2 := [][]byte{
		ridmap["本月用电账单"],
		ridmap["反向有功电能数据块"],
		ridmap["组合有功电能数据块"],
	}

	rbwdilist := [][]byte{
		ridmap["购电次数"],
		//ridmap["退费记录"],
		//ridmap["客户编码"],
		ridmap["年月日"],
		ridmap["软件配置"],
	}

	//[]byte{0x07, 0x01, 0x01, 0xFF}, //开户
	//[]byte{0x07, 0x01, 0x02, 0xFF}, //充值
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
	serconfig.BaudRate = d.BaudRate
	serconfig.DataBits = d.DataBits
	serconfig.Parity = d.Parity
	serconfig.StopBits = d.StopBits
	serconfig.Timeout = time.Microsecond * time.Duration(500000000/d.BaudRate)
	ret = map[string]interface{}{}
	ret[device.DevID] = d.Devid
	rsport, err := serial.Open(&serconfig)
	if err != nil {
		log.Errorf("open serial error %s", err.Error())
		return nil, err
	}
	defer rsport.Close()
	if rw == "r" {
		for _, di := range dilist {
			err = d.readValue(rsport, di, ret)
		}
		if _, ok := m["all"]; ok && (d.Devtype == "PMC340" || d.Devtype == "PMC320") {
			for _, di := range extralist1 {
				err = d.readValue(rsport, di, ret)
			}
			for _, di := range extralist2 {
				err = d.readValue(rsport, di, ret)
			}
		}
	} else {
		if d.Devtype == "PMC340" || d.Devtype == "PMC320" || d.Devtype == "DTZY83-Z" {
			for _, di := range rbwdilist {
				err = d.readValue(rsport, di, ret)
			}
		}
		if d.SafetyStat&0x02 == 0x02 {
			log.Debugln("safetystat", d.SafetyStat)
			d.writeValue(rsport, "软件配置", nil)
		}
		for k, v := range m {
			if k != device.DevID {
				if err = d.writeValue(rsport, k, v); err != nil {
					return
				}
			}
		}
		time.Sleep(time.Millisecond * 50)
		for _, di := range dilist {
			err = d.readValue(rsport, di, ret)
		}
		if d.Devtype == "PMC340" || d.Devtype == "PMC320" {
			for _, di := range extralist1 {
				err = d.readValue(rsport, di, ret)
			}
		}
	}
	return
}

/***************************************读写接口实现**************************************************/
