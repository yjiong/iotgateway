package device

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	//	"sync"
	log "github.com/Sirupsen/logrus"
	simplejson "github.com/bitly/go-simplejson"
	"github.com/yjiong/go_tg120/serial"
)

//var mutex sync.Mutex
type RSBAS struct {
	//继承于Device
	Device
	/**************按不同设备自定义*************************/
	BaudRate int
	DataBits int
	StopBits int
	Parity   string

	/**************按不同设备自定义*************************/
}

func init() {
	RegDevice["RSBAS"] = &RSBAS{}
}

func (d *RSBAS) NewDev(id string, ele map[string]string) (DeviceRWer, error) {
	ndev := new(RSBAS)
	ndev.Device = d.Device.NewDev(id, ele)
	/***********************初始化设备的特有的参数*****************************/
	ndev.BaudRate, _ = strconv.Atoi(ele["BaudRate"])
	ndev.DataBits, _ = strconv.Atoi(ele["DataBits "])
	ndev.StopBits, _ = strconv.Atoi(ele["StopBits"])
	ndev.Parity, _ = ele["Parity"]
	/***********************初始化设备的特有的参数*****************************/
	return ndev, nil
}

func (d *RSBAS) GetElement() (dict, error) {
	conn := dict{
		/***********************设备的特有的参数*****************************/
		"devaddr":  d.devaddr,
		"commif":   d.commif,
		"BaudRate": d.BaudRate,
		"DataBits": d.DataBits,
		"StopBits": d.StopBits,
		"Parity":   d.Parity,
		/***********************设备的特有的参数*****************************/
	}
	data := dict{
		"_devid": d.devid,
		"_type":  d.devtype,
		"_conn":  conn,
	}
	return data, nil
}

/***********************设备的参数说明帮助***********************************/
func (d *RSBAS) HelpDoc() interface{} {
	conn := dict{
		"devaddr": "设备地址",
		/***********RSBAS设备的参数*****************************/
		"commif":           "通信接口,比如(rs485-1)",
		"BaudRate":         "波特率,比如(9600)",
		"DataBits":         "数据位,比如(8)",
		"Parity":           "校验,(N,E,O)",
		"StopBits":         "停止位,比如(1)",
		"Function_code":    "modbus功能码 : (1,2,3,4,5,6,15,16)",
		"Starting_address": "操作起始地址,uint类型",
		"Quantity":         "寄存器数量,uint类型",
		/***********RSBAS设备的参数*****************************/
	}
	r_parameter := dict{
		"_devid": "被读取设备对象的id",
		/***********读取设备的参数*****************************/
		"Function_code":    "modbus功能码 : (1,2,3,4)",
		"Starting_address": "操作起始地址,uint类型",
		"Quantity":         "寄存器数量,uint类型",
		"说明":               "如果没有Function_code,Starting_address,Quantity字段,将按添加该设备时的参数读取设备",
		/***********读取设备的参数*****************************/
	}
	w_parameter := dict{
		"_devid": "被操作设备对象的id",
		/***********操作设备的参数*****************************/
		"Starting_address": "操作起始地址,uint类型",
		"Quantity":         "寄存器数量,uint类型",
		"value":            "要写入modbus设备的值,功能码为5和6时,值为uint16,功能码为15,16时,值为 [uint8...]",
		/***********操作设备的参数*****************************/
	}
	data := dict{
		"_devid": "添加设备对象的id",
		"_type":  "RSBAS", //设备类型
		"_conn":  conn,
	}
	dev_update := dict{
		"request": dict{
			"cmd":  "manager/dev/update.do",
			"data": data,
		},
	}
	readdev := dict{
		"request": dict{
			"cmd":  "do/getvar",
			"data": r_parameter,
		},
	}
	writedev := dict{
		"request": dict{
			"cmd":  "do/setvar",
			"data": w_parameter,
		},
	}
	helpdoc := dict{
		"1.添加设备": dev_update,
		"2.读取设备": readdev,
		"3.操作设备": writedev,
	}
	return helpdoc
}

/***********************设备的参数说明帮助***********************************/

/***************************************添加设备参数检验**********************************************/
func (d *RSBAS) CheckKey(ele dict) (bool, error) {

	if _, ok := ele["BaudRate"].(json.Number); !ok {
		return false, errors.New(fmt.Sprintf("RSBAS device must have int type element 波特率 :BaudRate"))
	}
	if _, ok := ele["DataBits"].(json.Number); !ok {
		return false, errors.New(fmt.Sprintf("RSBAS device must have int type element 数据位 :DataBits"))
	}
	if _, ok := ele["StopBits"].(json.Number); !ok {
		return false, errors.New(fmt.Sprintf("RSBAS device must have int type element 停止位 :StopBits"))
	}
	if _, ok := ele["Parity"].(string); !ok {
		return false, errors.New(fmt.Sprintf("RSBAS device must have string type element 校验 :Parity"))
	}
	fc, fc_ok := ele["Function_code"].(json.Number)
	if !fc_ok {
		return false, errors.New(fmt.Sprintf("RSBAS device must have int type element 功能码 :Function_code"))
	}
	if fci64, err := fc.Int64(); err != nil || fci64 < 1 || fci64 > 21 {
		return false, errors.New(fmt.Sprintf("Function_code :0 < value < 22 "))
	}
	if _, ok := ele["Starting_address"].(json.Number); !ok {
		return false, errors.New(fmt.Sprintf("RSBAS device must have int type element 起始地址 :Starting_address"))
	}
	if _, ok := ele["Quantity"].(json.Number); !ok {
		return false, errors.New(fmt.Sprintf("RSBAS device must have int type element 数量 :Quantity"))
	}
	return true, nil
}

/***************************************添加设备参数检验**********************************************/

/***************************************读写接口实现**************************************************/
func (d *RSBAS) read_cmd(taddr byte) []byte {
	sum := (0x81 + int(taddr)) & 0xff
	cmd := []byte{0xa5, 0x81, taddr, 0x00, 0x00, IntToBytes(sum)[3], 0x5a}
	return cmd
}

func (d *RSBAS) r_data_sum(data []byte) byte {
	sum := 0
	for i := 1; i < 9; i++ {
		sum += int(data[i])
	}
	return IntToBytes(sum & 0xff)[3]
}
func (d *RSBAS) RWDevValue(rw string, m dict) (ret dict, err error) {
	//log.SetLevel(log.DebugLevel)
	serconfig := serial.Config{}
	serconfig.Address = "/dev/ttyUSB0" // Commif[d.commif]
	serconfig.BaudRate = 9600          //d.BaudRate
	serconfig.DataBits = 8             //d.DataBits
	serconfig.Parity = "N"             //d.Parity
	serconfig.StopBits = 1             // d.StopBits
	serconfig.Timeout = 2 * time.Second
	taddr := byte(0x2)
	ret = map[string]interface{}{}
	ret["_devid"] = d.devid
	rsport, err := serial.Open(&serconfig)
	if err != nil {
		log.Errorf("open serial error %s", err.Error())
		return nil, err
	}
	defer rsport.Close()

	if rw == "r" {
		results := make([]byte, 11)
		for i := 0; i < 3; i++ {
			log.Debugf("send cmd = %x", d.read_cmd(taddr))
			if _, ok := rsport.Write(d.read_cmd(taddr)); ok != nil {
				log.Errorf("send cmd error  %s", ok.Error())
				return nil, ok
			}
			time.Sleep(100 * time.Millisecond)
			var len int
			len, err = rsport.Read(results)
			//if err != nil {
			//log.Errorf("serial read error  %s", err.Error())
			//}
			if err == nil && len == 11 && results[9] == d.r_data_sum(results) {
				log.Debugf("receive data = %x len = %d sum = %x", results, len, d.r_data_sum(results))
				break
			}
			time.Sleep(10 * time.Second)
		}
		if err != nil {
			log.Errorf("read RSBAS faild %s", err.Error())
			return nil, err
		} else {
			var value = make(map[bool]string)
			D1 := results[3]
			D2 := results[4]
			D3 := results[5]
			D4 := results[6]
			//D5 := results[7]
			D6 := results[8]
			ret["当前轿厢所处层站"] = fmt.Sprintf("%d", int(D1))
			value = map[bool]string{true: "是", false: "否"}
			ret["上行"] = value[D2&0x01 > 0]
			ret["下行"] = value[D2&0x02 > 0]
			ret["运行中"] = value[D2&0x04 > 0]
			ret["检修"] = value[D2&0x08 > 0]
			ret["电梯故障"] = value[D2&0x10 == 0]
			ret["泊梯"] = value[D2&0x20 > 0]
			ret["消防专用"] = value[D2&0x40 > 0]
			ret["消防返回"] = value[D2&0x80 > 0]
			ret["并联正常"] = value[D3&0x01 > 0]
			ret["群管理正常"] = value[D3&0x02 > 0]
			ret["电源正常"] = value[D3&0x04 > 0]
			ret["轿门门锁关闭"] = value[D3&0x08 > 0]
			ret["自发电运行"] = value[D3&0x10 > 0]
			ret["电梯到达"] = value[D3&0x20 > 0]
			ret["电梯开门"] = value[D3&0x40 > 0]
			ret["电梯关门"] = value[D3&0x80 > 0]
			ret["地震运行"] = value[D4&0x01 > 0]
			ret["安全装置正常"] = value[D4&0x02 > 0]
			ret["专用运行"] = value[D4&0x04 > 0]
			ret["火灾管制运行"] = value[D4&0x08 > 0]
			ret["位于门区"] = value[D4&0x10 > 0]
			ret["自救运行"] = value[D4&0x20 > 0]
			ret["发生A2级故障"] = value[D4&0x40 > 0]
			ret["发生A1级故障"] = value[D4&0x80 > 0]
			ret["厅门们锁关闭"] = value[D6&0x01 > 0]
			ret["抱闸打开"] = value[D6&0x02 > 0]
			ret["安全触板动作"] = value[D6&0x04 > 0]
			ret["光电保护动作"] = value[D6&0x08 > 0]
		}
		jsret, _ := json.Marshal(ret)
		inforet, _ := simplejson.NewJson(jsret)
		pinforet, _ := inforet.EncodePretty()
		log.Info(string(pinforet))
	}
	return ret, nil
}

/***************************************读写接口实现**************************************************/
