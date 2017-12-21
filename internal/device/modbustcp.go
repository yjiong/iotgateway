package device

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"
	//	"reflect"
	//	"sync"
	//	log "github.com/Sirupsen/logrus"
	"github.com/yjiong/go_tg120/modbus"
)

// ModbusTcp ..
type ModbusTcp struct {
	//继承于Device
	Device
	/**************按不同设备自定义*************************/
	FunctionCode    int
	StartingAddress uint16
	Quantity        uint16
	/**************按不同设备自定义*************************/
}

func init() {
	RegDevice["ModbusTcp"] = &ModbusTcp{}
}

// NewDev ..
func (d *ModbusTcp) NewDev(id string, ele map[string]string) (Devicerwer, error) {
	ndev := new(ModbusTcp)
	ndev.Device = d.Device.NewDev(id, ele)
	/***********************初始化设备的特有的参数*****************************/
	ndev.FunctionCode, _ = strconv.Atoi(ele["FunctionCode"])
	saint, _ := strconv.Atoi(ele["StartingAddress"])
	ndev.StartingAddress = uint16(saint)
	qint, _ := strconv.Atoi(ele["Quantity"])
	ndev.Quantity = uint16(qint)
	/***********************初始化设备的特有的参数*****************************/
	return ndev, nil
}

// GetElement ..
func (d *ModbusTcp) GetElement() (dict, error) {
	conn := dict{
		/***********************设备的特有的参数*****************************/
		"devaddr":         d.devaddr,
		"commif":          d.commif,
		"FunctionCode":    d.FunctionCode,
		"StartingAddress": d.StartingAddress,
		"Quantity":        d.Quantity,
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

// HelpDoc ..
func (d *ModbusTcp) HelpDoc() interface{} {
	conn := dict{
		"devaddr": "设备地址",
		/***********ModbusTcp设备的参数*****************************/
		"commif":          "通信接口,比如 : 192.168.1.20:502",
		"FunctionCode":    "modbus功能码 : (1,2,3,4,5,6,15,16)",
		"StartingAddress": "操作起始地址,uint类型",
		"Quantity":        "寄存器数量,uint类型",
		/***********ModbusTcp设备的参数*****************************/
	}
	rParameter := dict{
		"_devid": "被读取设备对象的id",
		/***********读取设备的参数*****************************/
		"FunctionCode":    "modbus功能码 : (1,2,3,4)",
		"StartingAddress": "操作起始地址,uint类型",
		"Quantity":        "寄存器数量,uint类型",
		"说明":              "如果没有FunctionCode,StartingAddress,Quantity字段,将按添加该设备时的参数读取设备",
		/***********读取设备的参数*****************************/
	}
	wParameter := dict{
		"_devid": "被操作设备对象的id",
		/***********操作设备的参数*****************************/
		"FunctionCode":    "modbus功能码 : (5,6,15,16)",
		"StartingAddress": "操作起始地址,uint类型",
		"Quantity":        "寄存器数量,uint类型",
		"value":           "要写入modbus设备的值,功能码为5和6时,值为uint16,功能码为15,16时,值为 [uint8...]",
		/***********操作设备的参数*****************************/
	}
	data := dict{
		"_devid": "添加设备对象的id",
		"_type":  "MudbusTcp", //设备类型
		"_conn":  conn,
	}
	devUpdate := dict{
		"request": dict{
			"cmd":  "manager/dev/update.do",
			"data": data,
		},
	}
	readdev := dict{
		"request": dict{
			"cmd":  "do/getvar",
			"data": rParameter,
		},
	}
	writedev := dict{
		"request": dict{
			"cmd":  "do/setvar",
			"data": wParameter,
		},
	}
	helpdoc := dict{
		"1.添加设备": devUpdate,
		"2.读取设备": readdev,
		"3.操作设备": writedev,
	}
	return helpdoc
}

/***********************设备的参数说明帮助***********************************/

/***************************************添加设备参数检验**********************************************/

// CheckKey ..
func (d *ModbusTcp) CheckKey(ele dict) (bool, error) {

	fc, fcOk := ele["FunctionCode"].(json.Number)
	if !fcOk {
		return false, fmt.Errorf("ModbusTcp device must have int type element 功能码 :FunctionCode")
	}
	if fci64, err := fc.Int64(); err != nil || fci64 < 1 || fci64 > 21 {
		return false, fmt.Errorf("FunctionCode :0 < value < 22 ")
	}
	if _, ok := ele["StartingAddress"].(json.Number); !ok {
		return false, fmt.Errorf("ModbusTcp device must have int type element 起始地址 :StartingAddress")
	}
	if _, ok := ele["Quantity"].(json.Number); !ok {
		return false, fmt.Errorf("ModbusTcp device must have int type element 数量 :Quantity")
	}
	return true, nil
}

/***************************************添加设备参数检验**********************************************/

/***************************************读写接口实现**************************************************/

// RWDevValue ..
func (d *ModbusTcp) RWDevValue(rw string, m dict) (ret dict, err error) {
	handler := modbus.NewTCPClientHandler(d.commif)
	slaveid, _ := strconv.Atoi(d.devaddr)
	handler.SlaveId = byte(slaveid)
	handler.Timeout = 1 * time.Second
	ret = map[string]interface{}{}
	ret["_devid"] = d.devid
	err = handler.Connect()
	if err != nil {
		return nil, err
	}
	defer handler.Close()
	functionCode := d.FunctionCode
	startAddr := d.StartingAddress
	quantity := d.Quantity
	fc, fcOk := m["FunctionCode"].(json.Number)
	sd, sdOk := m["StartingAddress"].(json.Number)
	qt, qtOk := m["Quantity"].(json.Number)
	if fcOk && sdOk && qtOk {
		fc64, _ := fc.Int64()
		functionCode = int(fc64)
		sd64, _ := sd.Int64()
		startAddr = uint16(sd64)
		qt64, _ := qt.Int64()
		quantity = uint16(qt64)
	}
	client := modbus.NewClient(handler)
	var myRfunc func(address, quantity uint16) (results []byte, err error)
	if rw == "r" {
		switch functionCode {
		case 1:
			myRfunc = client.ReadCoils
		case 2:
			myRfunc = client.ReadDiscreteInputs
		case 3:
			myRfunc = client.ReadHoldingRegisters
		case 4:
			myRfunc = client.ReadInputRegisters
			//		client.ReadWriteMultipleRegisters
			//		client.ReadFIFOQueue
		default:
			return nil, fmt.Errorf("尚未支持的读操作  FunctionCode : %d", functionCode)
		}
		var results []byte
		results, err = myRfunc(startAddr, quantity)
		if err == nil {
			var retlist []int
			for _, b := range results {
				retlist = append(retlist, int(b))
			}
			ret["Modbus-value"] = retlist
		}
	} else if rw == "w" {
		var results []byte
		var value uint16
		var valuelist []byte
		if v, ok := m["value"].(json.Number); ok && (functionCode == 5 || functionCode == 6) {
			v64, _ := v.Int64()
			value = uint16(v64)
		} else {
			return nil, errors.New("write modbus singlecoil or registers need value : uint16")
		}
		if vif, ok := m["value"].([]interface{}); ok && (functionCode == 15 || functionCode == 16) {
			for _, v := range vif {
				if vi, ok := v.(json.Number); ok {
					vi64, _ := vi.Int64()
					valuelist = append(valuelist, IntToBytes(int(vi64))[3])
				} else {
					return nil, errors.New("write modbus singlecoil or registers need value : [uint8...]")
				}
			}
		} else {
			return nil, errors.New("write modbus singlecoil or registers need value : [uint8...]")
		}
		switch functionCode {
		case 5:
			{
				results, err = client.WriteSingleCoil(startAddr, value)
				if err == nil {
					var retlist []int
					for _, b := range results {
						retlist = append(retlist, int(b))
					}
					ret["Modbus-write"] = retlist
				}
			}
		case 15:
			{
				results, err = client.WriteMultipleCoils(startAddr, quantity, valuelist)
				if err == nil {
					var retlist []int
					for _, b := range results {
						retlist = append(retlist, int(b))
					}
					ret["Modbus-write"] = retlist
				}
			}
		case 6:
			{
				results, err = client.WriteSingleRegister(startAddr, value)
				if err == nil {
					var retlist []int
					for _, b := range results {
						retlist = append(retlist, int(b))
					}
					ret["Modbus-write"] = retlist
				}
			}
		case 16:
			{
				results, err = client.WriteMultipleRegisters(startAddr, quantity, valuelist)
				if err == nil {
					var retlist []int
					for _, b := range results {
						retlist = append(retlist, int(b))
					}
					ret["Modbus-write"] = retlist
				}
			}
		default:
			return nil, fmt.Errorf("尚未支持的写操作  FunctionCode : %d", functionCode)

		}
	}
	return ret, err
}

/***************************************读写接口实现**************************************************/
