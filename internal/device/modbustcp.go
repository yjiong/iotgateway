package device

import (
	//"encoding/hex"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"
	//	"sync"
	log "github.com/sirupsen/logrus"
	"github.com/yjiong/iotgateway/modbus"
)

// ModbusTcp ..
type ModbusTcp struct {
	//组合Device
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
func (d *ModbusTcp) GetElement() (Dict, error) {
	conn := Dict{
		/***********************设备的特有的参数*****************************/
		DevAddr:           d.Devaddr,
		"commif":          d.Commif,
		"FunctionCode":    d.FunctionCode,
		"StartingAddress": d.StartingAddress,
		"Quantity":        d.Quantity,
		/***********************设备的特有的参数*****************************/
	}
	data := Dict{
		DevID:   d.Devid,
		DevType: d.Devtype,
		DevConn: conn,
	}
	return data, nil
}

/***********************设备的参数说明帮助***********************************/

// HelpDoc ..
func (d *ModbusTcp) HelpDoc() interface{} {
	conn := Dict{
		DevAddr: "设备地址",
		/***********ModbusTcp设备的参数*****************************/
		"commif":          "通信接口,比如 : 192.168.1.20:502",
		"FunctionCode":    "modbus功能码 : (1,2,3,4,5,6,15,16)",
		"StartingAddress": "操作起始地址,uint类型",
		"Quantity":        "寄存器数量,uint类型",
		/***********ModbusTcp设备的参数*****************************/
	}
	rParameter := Dict{
		DevID: "被读取设备对象的id",
		/***********读取设备的参数*****************************/
		"FunctionCode":    "modbus功能码 : (1,2,3,4)",
		"StartingAddress": "操作起始地址,uint类型",
		"Quantity":        "寄存器数量,uint类型",
		"说明":              "如果没有FunctionCode,StartingAddress,Quantity字段,将按添加该设备时的参数读取设备",
		/***********读取设备的参数*****************************/
	}
	wParameter := Dict{
		DevID: "被操作设备对象的id",
		/***********操作设备的参数*****************************/
		"FunctionCode":    "modbus功能码 : (5,6,15,16)",
		"StartingAddress": "操作起始地址,uint类型",
		"Quantity":        "寄存器数量,uint类型",
		"value":           "要写入modbus设备的值,功能码为5和6时,值为uint16,功能码为15,16时,值为 [uint8...]",
		/***********操作设备的参数*****************************/
	}
	data := Dict{
		DevID:   "添加设备对象的id",
		DevType: "MudbusTcp", //设备类型
		DevConn: conn,
	}
	devUpdate := Dict{
		"request": Dict{
			"cmd":  UpdateDevItem,
			"data": data,
		},
	}
	readdev := Dict{
		"request": Dict{
			"cmd":  GetDevVar,
			"data": rParameter,
		},
	}
	writedev := Dict{
		"request": Dict{
			"cmd":  SetDevVar,
			"data": wParameter,
		},
	}
	helpdoc := Dict{
		"1.添加设备": devUpdate,
		"2.读取设备": readdev,
		"3.操作设备": writedev,
	}
	return helpdoc
}

/***********************设备的参数说明帮助***********************************/

/***************************************添加设备参数检验**********************************************/

// CheckKey ..
func (d *ModbusTcp) CheckKey(ele Dict) (bool, error) {

	fc, fcOk := ele["FunctionCode"].(json.Number)
	if !fcOk {
		if fcs, ok := ele["FunctionCode"].(string); ok {
			fc = json.Number(fcs)
		} else {
			return false, fmt.Errorf("ModbusTcp device must have int type element 功能码 :FunctionCode")
		}
	}
	if fci64, err := fc.Int64(); err != nil || fci64 < 1 || fci64 > 21 {
		return false, fmt.Errorf("FunctionCode :0 < value < 22 ")
	}
	if _, ok := ele["StartingAddress"].(json.Number); !ok {
		if _, ok := ele["StartingAddress"].(string); !ok {
			return false, fmt.Errorf("ModbusTcp device must have int type element 起始地址 :StartingAddress")
		}
	}
	if _, ok := ele["Quantity"].(json.Number); !ok {
		if _, ok := ele["Quantity"].(string); !ok {
			return false, fmt.Errorf("ModbusTcp device must have int type element 数量 :Quantity")
		}
	}
	return true, nil
}

// GetMultiType .....
func GetMultiType(results []byte) (ret Dict) {
	log.Debugf("% x", results)
	ret = make(Dict)
	var retlist []int
	for _, b := range results {
		retlist = append(retlist, int(b))
	}
	var retlist16 []int16
	var retlist16u []uint16
	var retlist32 []int
	var retlist32inverse []int
	var retlistFloat32 []float32
	var retlistFloat32inverse []float32
	var retlistFloat64 []float64
	var retlistFloat64inverse []float64
	var i int
	for k := 0; k < len(results); k += 2 {
		i = k
		retlist16 = append(retlist16,
			int16(results[i])<<8+
				int16(results[i+1]))
		retlist16u = append(retlist16u,
			uint16(results[i])<<8+
				uint16(results[i+1]))
		if i%4 == 0 && i != 0 {
			i -= 4
			retlist32 = append(retlist32,
				BytesToInt([]byte{results[i+2],
					results[i+3],
					results[i],
					results[i+1],
				}))
			retlist32inverse = append(retlist32inverse,
				BytesToInt([]byte{results[i],
					results[i+1],
					results[i+2],
					results[i+3],
				}))
			retlistFloat32 = append(retlistFloat32,
				ByteToFloat32([]byte{results[i+1],
					results[i],
					results[i+3],
					results[i+2],
				}))
			retlistFloat32inverse = append(retlistFloat32inverse,
				ByteToFloat32([]byte{results[i+3],
					results[i+2],
					results[i+1],
					results[i],
				}))
		}
		i = k
		if i%8 == 0 && i != 0 {
			i -= 8
			retlistFloat64 = append(retlistFloat64,
				ByteToFloat64([]byte{results[i+1],
					results[i],
					results[i+3],
					results[i+2],
					results[i+5],
					results[i+4],
					results[i+7],
					results[i+6],
				}))
			retlistFloat64inverse = append(retlistFloat64inverse,
				ByteToFloat64([]byte{results[i+7],
					results[i+6],
					results[i+5],
					results[i+4],
					results[i+3],
					results[i+2],
					results[i+1],
					results[i],
				}))

		}
	}

	ret["Modbus-value"] = retlist
	ret["int16-value"] = retlist16
	ret["uint16-value"] = retlist16u
	ret["long-value"] = retlist32
	ret["long-inverse-value"] = retlist32inverse
	ret["hexstr-value"] = fmt.Sprintf("% x", results)
	ret["Float32-value"] = retlistFloat32
	ret["Float32-inverse-value"] = retlistFloat32inverse
	ret["Float64-value"] = retlistFloat64
	ret["Float64-inverse-value"] = retlistFloat64inverse
	ret["binstr-value"] = fmt.Sprintf("% b", results)
	return
}

// Ifa2uint16 ...
func Ifa2uint16(ifa interface{}) (uint16, error) {
	var retu16 uint16
	switch ifa.(type) {
	case json.Number:
		if v64, err := ifa.(json.Number).Int64(); err == nil {
			retu16 = uint16(v64)
		} else {
			return 0, err
		}
	case string:
		if v, err := strconv.Atoi(ifa.(string)); err == nil {
			retu16 = uint16(IntToBytes(v)[3])
		} else {
			return 0, err
		}
	case int:
		retu16 = uint16(IntToBytes(ifa.(int))[3])
	case int16:
		retu16 = uint16(ifa.(int16))
	case uint16:
		retu16 = ifa.(uint16)
	case float64:
		retu16 = uint16(ifa.(float64))
	default:
		return 0, errors.New("assert interface failed")
	}
	return retu16, nil
}

// Ifal2bytel ...
func Ifal2bytel(ifal interface{}) ([]byte, error) {
	var bl []byte
	if abl, ok := ifal.([]byte); ok {
		log.Debugf("ifal type is %s", "[]byte")
		return abl, nil
	}
	if ail, ok := ifal.([]int); ok {
		log.Debugf("ifal type is %s", "[]int")
		for _, ai := range ail {
			bl = append(bl, IntToBytes(ai)[3])
		}
		return bl, nil
	}
	if abl, ok := ifal.(string); ok {
		log.Debugf("in string assert ifal type is %s",
			reflect.TypeOf(abl).Name())
		if hv, err := base64.StdEncoding.DecodeString(abl); err == nil {
			bl = hv
		} else {
			return nil, err
		}
	}
	if assertv, ok := ifal.([]interface{}); ok {
		log.Debugf("ifal type is %s", reflect.TypeOf(assertv[0]).Name())
		for _, fal := range assertv {
			switch fal.(type) {
			case json.Number:
				if v64, err := fal.(json.Number).Int64(); err == nil {
					bl = append(bl, IntToBytes(int(v64))[3])
				} else {
					return nil, err
				}
			case string:
				if v, err := strconv.Atoi(fal.(string)); err == nil {
					bl = append(bl, IntToBytes(v)[3])
				} else {
					return nil, err
				}
			case int:
				bl = append(bl, IntToBytes(fal.(int))[3])
			case int64:
				bl = append(bl, IntToBytes(int(fal.(int64)))[3])
			case float64:
				bl = append(bl, IntToBytes(int(fal.(float64)))[3])
			default:
				return nil, errors.New("assert interface failed")
			}
		}
	} else {
		return nil, errors.New("assert interface failed")
	}
	log.Debugf("return byte list = % x", bl)
	return bl, nil
}

/***************************************添加设备参数检验**********************************************/

/***************************************读写接口实现**************************************************/

// RWDevValue ..
func (d *ModbusTcp) RWDevValue(rw string, m Dict) (ret Dict, err error) {
	handler := modbus.NewTCPClientHandler(d.Commif)
	slaveid, _ := strconv.Atoi(d.Devaddr)
	handler.SlaveId = byte(slaveid)
	handler.Timeout = 1 * time.Second
	ret = map[string]interface{}{}
	err = handler.Connect()
	if err != nil {
		return nil, err
	}
	defer handler.Close()
	functionCode := d.FunctionCode
	startAddr := d.StartingAddress
	quantity := d.Quantity
	if fc, err := Ifa2uint16(m["FunctionCode"]); err == nil {
		functionCode = int(fc)
	}
	if sd, err := Ifa2uint16(m["StartingAddress"]); err == nil {
		startAddr = sd
	}
	if qt, err := Ifa2uint16(m["Quantity"]); err == nil {
		quantity = qt
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
			ret = GetMultiType(results)
		}
	} else if rw == "w" {
		var results []byte
		var value uint16
		var valuelist []byte
		if functionCode == 5 || functionCode == 6 {
			if value, err = Ifa2uint16(m["value"]); err == nil {
				return nil, errors.New("write modbus singlecoil or registers need value : uint16")
			}
		}
		if functionCode == 15 || functionCode == 16 {
			if valuelist, err = Ifal2bytel(m["value"]); err != nil {
				log.Debugf("valueslist:%t", m["value"])
				return nil, err
			}
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
