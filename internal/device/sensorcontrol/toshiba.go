package sensorcontrol

import (
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/yjiong/iotgateway/internal/device"
	"math"
	//simplejson "github.com/bitly/go-simplejson"
	"strconv"
	//"strings"
	//"sync"
)

var onoff = map[int]string{
	0: "off",
	1: "on",
}

var setonoff = map[int]string{
	0:      "off",
	0XFF00: "on",
}

var filtersign = map[int]string{
	0:      "noaction",
	0xff00: "reset",
}

var permit = map[int]string{
	0: "permit",
	1: "prohibit",
}

var setpermit = map[int]string{
	0:      "permit",
	0xff00: "prohibit",
}

var alarm = map[int]string{
	0: "normal",
	1: "abnormal",
}

var operationMode = map[int]string{
	0: "invalid",
	1: "heat",
	2: "cool",
	3: "dry",
	4: "fan",
	5: "auto heat",
	6: "auto cool",
	7: "unfix",
}

var setOperationMode = map[int]string{
	0: "Invalid",
	1: "heat",
	2: "cool",
	3: "dry",
	4: "fan",
	5: "auto",
}

var fanSpeed = map[int]string{
	0: "invalid",
	1: "Fan Stop",
	2: "Auto",
	3: "High",
	4: "Medium",
	5: "Low",
	6: "Ultra Low",
	7: "unfix",
}

var louver = map[int]string{
	0: "invalid",
	1: "swing",
	2: "f1",
	3: "f2",
	4: "f3",
	5: "f4",
	6: "f5",
	7: "stop",
}

var fanrsfm = map[byte]string{
	0: "disabled",
	1: "enabled",
}

var fanrsfm76 = map[byte]string{
	0: "All operation modes enabled",
	1: "Cooling/drying disabled",
	2: "Heating drying disabled",
	3: "Fan only enabled",
}

// ValToshiba ...
type ValToshiba struct {
	RoomTemperature float32 `json:"RoomTemperature,string"`
}

// TOSHIBA ..
type TOSHIBA struct {
	//组合ModebusRtu
	device.ModbusRtu
	/**************按不同设备自定义*************************/
	IndoorNum string
	/**************按不同设备自定义*************************/
}

func init() {
	device.RegDevice["TOSHIBA-VRV"] = &TOSHIBA{}
}

// NewDev ..
func (d *TOSHIBA) NewDev(id string, ele map[string]string) (device.Devicerwer, error) {
	ndev := new(TOSHIBA)
	ndev.Device = d.Device.NewDev(id, ele)
	/***********************初始化设备的特有的参数*****************************/
	ndev.BaudRate = 9600
	ndev.DataBits = 8
	ndev.StopBits = 1
	ndev.Parity = "E"
	ndev.IndoorNum = ele["IndoorNum"]
	/***********************初始化设备的特有的参数*****************************/
	return ndev, nil
}

// GetElement ..
func (d *TOSHIBA) GetElement() (device.Dict, error) {
	conn := device.Dict{
		device.DevAddr: d.Devaddr,
		/***********************设备的特有的参数*****************************/
		"commif": d.Commif,
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
func (d *TOSHIBA) HelpDoc() interface{} {
	conn := device.Dict{
		device.DevAddr: "设备地址",
		/***********TOSHIBA设备的参数*****************************/
		"commif":    "通信接口,比如(rs485-1)",
		"IndoorNum": "内机编号",
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
		"_varname.1":   "ON/OFF setting",
		"_varvalue.1":  "(on|off)",
		"_varname.2":   "Filter sign reset setting",
		"_varvalue.2":  "(noaction|reset)",
		"_varname.3":   "Relay 1ch output for TCB-IFCG1TLE",
		"_varvalue.3":  "(on|off)",
		"_varname.4":   "Relay 2ch output for TCB-IFCG1TLE",
		"_varvalue.4":  "(on|off)",
		"_varname.5":   "Relay 3ch output for TCB-IFCG3TLE",
		"_varvalue.5":  "(on|off)",
		"_varname.6":   "Relay 4ch output for TCB-IFCG1TLE",
		"_varvalue.6":  "(on|off)",
		"_varname.7":   "Local operation prohibit for TCB-IFCG1TLE",
		"_varvalue.7":  "(permit|prohibit)",
		"_varname.8":   "Setting Temperature",
		"_varvalue.8":  "(float)",
		"_varname.9":   "Accumulated operation time",
		"_varvalue.9":  "(uint16)",
		"_varname.10":  "Operation mode",
		"_varvalue.10": "(Invalid|heat|cool|dry|fan|auto)",
		"_varname.11":  "Fan speed",
		"_varvalue.11": "(Invalid|Auto|High|Medium|Low|unfix)",
		"_varname.12":  "Louver",
		"_varvalue.12": "(invalid|swing|f1|f2|f3|f4|f5|stop)",
		"_varname.13":  "Remote controller permit/Prohibit",
		"_varvalue.13": `(Remote controller on/off prohibit setting(bit0)
Remote controller mode prohibit setting(bit1)
Remote controller setpoint prohibit setting(bit2)
Remote controller louver prohibit setting(bit3)
Remote controller fan speed prohibit setting(bit4)
1=prohibit 0=permit)`,
		"解释": "one cmd set one value",
	}
	data := device.Dict{
		device.DevID:   "添加设备对象的id",
		device.DevType: "TOSHIBA", //设备类型
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

// CheckKey ..
/***************************************添加设备参数检验**********************************************/
func (d *TOSHIBA) CheckKey(ele device.Dict) (bool, error) {
	im, imok := ele["commif"].(string)
	if !imok {
		return false, fmt.Errorf("ModbusRtu device must have string type element 通讯接口 :commif")
	}
	if _, ok := device.Commif[im]; !ok {
		return false, fmt.Errorf("通讯接口:%s不存在", im)
	}
	_, sbOk := ele["IndoorNum"].(json.Number)
	if !sbOk {
		return false, errors.New("TOSHIBA device must have int type element : IndoorNum")
	}
	return true, nil
}

/***************************************添加设备参数检验**********************************************/

func (d *TOSHIBA) caseinternal(f int, setrange map[int]string, m device.Dict) (res json.Number, err error) {
	if val, ok := m["_varvalue"]; ok {
		if sval, ok := val.(string); ok {
			for k, v := range setrange {
				if v == sval {
					sk := strconv.Itoa(k)
					res = json.Number(sk)
					break
				}
			}
		} else {
			return res, errors.New("varvalue is not string")
		}
	} else {
		return res, errors.New("varvalue invalid")
	}
	d.FunctionCode = f
	return res, err
}

func (d *TOSHIBA) encode(m device.Dict) (json.Number, error) {
	name, _ := m["_varname"]
	var err error
	IndoorNum, _ := strconv.Atoi(d.IndoorNum)
	switch name {
	case "ON/OFF setting":
		{
			d.StartingAddress = uint16(152*IndoorNum - 152)
			return d.caseinternal(5, setonoff, m)
		}
	case "Filter sign reset setting":
		{
			d.StartingAddress = uint16(152*IndoorNum - 151)
			return d.caseinternal(5, filtersign, m)
		}
	case "Relay 1ch output for TCB-IFCG1TLE":
		{
			d.StartingAddress = uint16(152*IndoorNum - 112)
			return d.caseinternal(5, setonoff, m)
		}
	case "Relay 2ch output for TCB-IFCG1TLE":
		{
			d.StartingAddress = uint16(152*IndoorNum - 111)
			return d.caseinternal(5, setonoff, m)
		}
	case "Relay 3ch output for TCB-IFCG1TLE":
		{
			d.StartingAddress = uint16(152*IndoorNum - 110)
			return d.caseinternal(5, setonoff, m)
		}
	case "Relay 4ch output for TCB-IFCG1TLE":
		{
			d.StartingAddress = uint16(152*IndoorNum - 109)
			return d.caseinternal(5, setonoff, m)
		}
	case "Local operation prohibit for TCB-IFCG1TLE":
		{
			d.StartingAddress = uint16(152*IndoorNum - 108)
			return d.caseinternal(5, setpermit, m)
		}
	case "Setting Temperature":
		{
			d.StartingAddress = uint16(156*IndoorNum - 156)
			var res json.Number
			if val, ok := m["_varvalue"]; ok {
				if jnval, ok := val.(json.Number); ok {
					fval, _ := jnval.Float64()
					rbyte, _ := d.float2hex2(fval)
					itemp := int(rbyte[0])*0x100 + int(rbyte[1])
					res = json.Number(strconv.Itoa(itemp))
				} else {
					err = errors.New("varvalue is not string or float")
				}
				if sval, ok := val.(string); ok {
					jnval := json.Number(sval)
					fval, _ := jnval.Float64()
					rbyte, _ := d.float2hex2(fval)
					itemp := int(rbyte[0])*0x100 + int(rbyte[1])
					res = json.Number(strconv.Itoa(itemp))
					err = nil
				} else {
					err = errors.New("varvalue is not string or float")
				}
			} else {
				err = errors.New("varvalue invalid")
			}
			d.FunctionCode = 6
			return res, err
		}
	case "Accumulated operation time":
		{
			d.StartingAddress = uint16(156*IndoorNum - 155)
			var res json.Number
			if val, ok := m["_varvalue"]; ok {
				if jnval, ok := val.(json.Number); ok {
					ival, _ := jnval.Int64()
					res = json.Number(strconv.Itoa(int(ival)))
				} else {
					err = errors.New("varvalue is not string or int")
				}
				if sval, ok := val.(string); ok {
					jnval := json.Number(sval)
					ival, _ := jnval.Int64()
					res = json.Number(strconv.Itoa(int(ival)))
					err = nil
				} else {
					err = errors.New("varvalue is not string or int")
				}
			} else {
				err = errors.New("varvalue invalid")
			}
			d.FunctionCode = 6
			return res, err
		}
	case "Operation mode":
		{
			d.StartingAddress = uint16(156*IndoorNum - 150)
			return d.caseinternal(6, setOperationMode, m)
		}
	case "Fan speed":
		{
			d.StartingAddress = uint16(156*IndoorNum - 149)
			return d.caseinternal(6, fanSpeed, m)
		}
	case "Louver":
		{
			d.StartingAddress = uint16(156*IndoorNum - 148)
			return d.caseinternal(6, louver, m)
		}
	case "Remote controller permit/Prohibit":
		{
			d.StartingAddress = uint16(156*IndoorNum - 147)
			var res json.Number
			if val, ok := m["_varvalue"]; ok {
				if jnval, ok := val.(json.Number); ok {
					ival, _ := jnval.Int64()
					res = json.Number(strconv.Itoa(int(ival)))
				} else {
					err = errors.New("varvalue is not string or int")
				}
				if sval, ok := val.(string); ok {
					jnval := json.Number(sval)
					ival, _ := jnval.Int64()
					res = json.Number(strconv.Itoa(int(ival)))
					err = nil
				} else {
					err = errors.New("varvalue is not string or int")
				}
			} else {
				err = errors.New("varvalue invalid")
			}
			d.FunctionCode = 6
			return res, err
		}
	default:
		{
			return json.Number("0"), errors.New("错误的_varname")
		}
	}
}

func (d *TOSHIBA) hex2float(hex2 []byte) (vf float64, err error) {
	if len(hex2) != 2 {
		return 0, errors.New("wrong len hex data")
	}
	var m int16
	e := (hex2[0] >> 3) & 0xf
	m = ((int16(hex2[0]))<<8)&0x700 + int16(hex2[1])
	if e >= 0 {
		if hex2[0]&0x80 == 0x80 {
			m = m - 0x800
		}
		log.Debugf("e=%d,m=%d", e, m)
		//vf = float64(int32(m)<<e) * 0.01
		vf = float64(int32(m)) * 0.1
		log.Debugf("vf=%f", vf)
	} // else {
	//vf = float32(m)
	//for i := e; i <= 0; i++ {
	//vf = vf / 2
	//}
	//}
	return vf, err
}

func (d *TOSHIBA) float2hex2(vf float64) (hex2 []byte, err error) {
	hex2 = make([]byte, 2)
	if vf < -671088.63 || vf > 670760.95 {
		return hex2, errors.New("wrong date")
	}
	var powerE uint16
	var m uint16
	var h2 uint16
	if vf >= 0 {
		//powerE = uint16(math.Ceil(vf*100) / 0x800)
		//m = uint16(int(math.Ceil(vf*100)) % 0x800)
		//h2 = uint16(powerE<<11 + m)
		powerE = uint16(math.Ceil(vf*10) / 0x800)
		m = uint16(int(math.Ceil(vf*10)) % 0x800)
		h2 = uint16(powerE<<11 + m)
	} else {
		//powerE = uint16(math.Ceil(0-vf*100) / 0x800)
		//m = uint16(int(math.Ceil(0-vf*100)) % 0x800)
		//h2 = uint16(0x8000 + powerE<<11 + m)
		powerE = uint16(math.Ceil(vf*10) / 0x800)
		m = uint16(int(math.Ceil(vf*10)) % 0x800)
		h2 = uint16(powerE<<11 + m)
	}
	log.Debugln(powerE, m, h2)
	log.Debugf("h2=%x", h2)
	hex2[0] = byte(h2 >> 8)
	hex2[1] = byte(h2 & 0xff)
	return hex2, nil
}

/***************************************读写接口实现**************************************************/

// RWDevValue ..
func (d *TOSHIBA) RWDevValue(rw string, m device.Dict) (ret device.Dict, err error) {
	ret = make(device.Dict)
	defer func() {
		if driveErr := recover(); driveErr != nil {
			log.Errorf("drive programer  error : (%s)", driveErr)
			errstr := fmt.Sprintf("drive programer  error : %s", driveErr)
			err = errors.New(errstr)
		}
	}()
	d.Mutex.Lock()
	defer d.Mutex.Unlock()
	//log.SetLevel(log.DebugLevel)
	ret[device.DevID] = d.Devid
	if rw == "r" {
		d.Quantity = 8
		d.FunctionCode = 2
		IndoorNum, _ := strconv.Atoi(d.IndoorNum)
		d.StartingAddress = uint16(152*IndoorNum - 152)
		log.Debugf("IndoorNum=%d", IndoorNum)
		var inputStatus device.Dict
		inputStatus, err = d.ModbusRtu.RWDevValue("r", nil)
		inputStatusInt, _ := inputStatus["Modbus-value"].([]int)
		ret["ON/OFF setting status"] = onoff[0x01&inputStatusInt[0]]
		ret["Filter sign status"] = alarm[0x01&(inputStatusInt[0]>>1)]
		ret["Alarm status"] = alarm[0x01&(inputStatusInt[0]>>2)]
		d.StartingAddress = uint16(152*IndoorNum - 96)
		inputStatus, err = d.ModbusRtu.RWDevValue("r", nil)
		inputStatusInt, _ = inputStatus["Modbus-value"].([]int)
		ret["ON/OFF input for TCB-IFCG1TLE"] = 1 & inputStatusInt[0]
		ret["Alarm input for TCB-IFCG1TLE"] = 1 & (inputStatusInt[0] >> 1)
		ret["Din2 input for TCB-IFCG1TLE"] = 1 & (inputStatusInt[0] >> 2)
		ret["Din3 input for TCB-IFCG1TLE"] = 1 & (inputStatusInt[0] >> 3)
		ret["Din4 input for TCB-IFCG1TLE"] = 1 & (inputStatusInt[0] >> 4)
		ret["Din1 input for TCB-IFCG1TLE"] = 1 & (inputStatusInt[0] >> 5)
		d.Quantity = 39
		d.FunctionCode = 4
		d.StartingAddress = uint16(156*IndoorNum - 156)
		var inputRegister device.Dict
		inputRegister, err = d.ModbusRtu.RWDevValue("r", nil)
		inputRegisterInt, _ := inputRegister["Modbus-value"].([]int)
		var inputRegisterByte []byte
		for _, vi := range inputRegisterInt {
			inputRegisterByte = append(inputRegisterByte, byte(vi))
		}
		log.Debugln(inputRegisterByte[0:4])
		ret["Room Temperature"], err = d.hex2float(inputRegisterByte[0:2])
		ret["Setting Temperature status"], err = d.hex2float(inputRegisterByte[2:4])
		ret["Alarm code"] = fmt.Sprintf("%d", inputRegisterByte[4:12])
		ret["Model name"] = string(inputRegisterByte[12:28])
		ret["Serial number"] = string(inputRegisterByte[28:44])
		ret["Indoor unit capacity"], err = d.hex2float(inputRegisterByte[44:46])
		ret["Indoo unit type"] = fmt.Sprintf("%x", inputRegisterByte[46:48])
		ret["Analog input for TCB-IFCG1TLE"] = fmt.Sprintf("%x,%x,%x,%x",
			inputRegisterByte[48:50],
			inputRegisterByte[50:52],
			inputRegisterByte[52:54],
			inputRegisterByte[54:56])
		ret["Operation mode/Fan range"] = map[string]string{
			"Cooling mode":        fanrsfm[inputRegisterByte[60]&0x02/0x02],
			"Drying mode":         fanrsfm[inputRegisterByte[60]&0x04/0x04],
			"Heating mode":        fanrsfm[inputRegisterByte[60]&0x08/0x08],
			"Ventilation":         fanrsfm[inputRegisterByte[60]&0x10/0x10],
			"Auto mode":           fanrsfm[inputRegisterByte[60]&0x20/0x20],
			"All Operatin modes":  fanrsfm76[inputRegisterByte[60]>>6],
			"Ultra-low fan speed": fanrsfm[inputRegisterByte[61]&0x01/0x01],
			"low fan speed":       fanrsfm[inputRegisterByte[61]&0x02/0x02],
			"Medium fan speed":    fanrsfm[inputRegisterByte[61]&0x04/0x04],
			"High fan speed":      fanrsfm[inputRegisterByte[61]&0x08/0x08],
		} //fmt.Sprintf("%x", inputRegisterByte[60:62])
		ret["Cooling temperature range"] = fmt.Sprintf("Upper-limit:%d°C,Lower-limit:%d°C",
			inputRegisterByte[62]/2-35,
			inputRegisterByte[63]/2-35)
		ret["Heating temperature range"] = fmt.Sprintf("Upper-limit:%d°C,Lower-limit:%d°C",
			inputRegisterByte[64]/2-35,
			inputRegisterByte[65]/2-35)
		ret["Dry temperature range"] = fmt.Sprintf("Upper-limit:%d°C,Lower-limit:%d°C",
			inputRegisterByte[66]/2-35,
			inputRegisterByte[67]/2-35)
		ret["Auto temperature range"] = fmt.Sprintf("Upper-limit:%d°C,Lower-limit:%d°C",
			inputRegisterByte[68]/2-35,
			inputRegisterByte[69]/2-35)
		ret["Operatin mode"] = operationMode[int(inputRegisterByte[71])]
		ret["Fan speed"] = fanSpeed[int(inputRegisterByte[73])]
		ret["Louver"] = louver[int(inputRegisterByte[75])]
		ppbit := int(inputRegisterByte[75])
		ret["Remote controller on/off prohibit setting"] = permit[(ppbit)&0x01]
		ret["Remote controller mode porhibit setting"] = permit[(ppbit>>1)&0x01]
		ret["Remote controller setpoint prohibit setting"] = permit[(ppbit>>2)&0x01]
		ret["Remote controller louver prohibit setting"] = permit[(ppbit>>3)&0x01]
		ret["Remote controller fan speed prohibit setting"] = permit[(ppbit>>4)&0x01]
		/****************************************write device**********************************************/
	} else {
		wval, werr := d.encode(m)
		if werr != nil {
			ret["error"] = werr.Error()
			return ret, nil
		}
		log.Debugln("wval", wval)
		log.Debugln("functioncode=", d.FunctionCode, "startAddress=", d.StartingAddress)
		bmDict, berr := d.ModbusRtu.RWDevValue("w", device.Dict{"value": wval})
		if berr == nil {
			log.Infof("设置-%s receive data = %v", m["_varname"], bmDict)
			ret["cmdStatus"] = "successful"
		} else {
			ret["error"] = berr.Error()
			log.Debugln(ret)
			return ret, nil
		}
	}
	return
}

//func getnm(inf interface{}) uint16 {
//var ruint uint16
//ruint = 0
//if nm, ok := inf.(json.Number); ok {
//nmi64, _ := nm.Int64()
//ruint = uint16(nmi64)
//} else {
//if nm, ok := inf.(string); ok {
//snm, _ := strconv.Atoi(nm)
//ruint = uint16(snm)
//}
//}
//return ruint
//}
