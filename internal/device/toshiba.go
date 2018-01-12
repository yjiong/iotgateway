package device

import (
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	//simplejson "github.com/bitly/go-simplejson"
	//"strconv"
	//"strings"
	//"sync"
)

type roomtemp struct {
	RoomTemperature float32 `json:"RoomTemperature,string"`
}

// TOSHIBA ..
type TOSHIBA struct {
	//继承于ModebusRtu
	ModbusRtu
	/**************按不同设备自定义*************************/
	/**************按不同设备自定义*************************/
}

func init() {
	RegDevice["TOSHIBA"] = &TOSHIBA{}
}

// NewDev ..
func (d *TOSHIBA) NewDev(id string, ele map[string]string) (Devicerwer, error) {
	ndev := new(TOSHIBA)
	ndev.Device = d.Device.NewDev(id, ele)
	/***********************初始化设备的特有的参数*****************************/
	ndev.BaudRate = 9600
	ndev.DataBits = 8
	ndev.StopBits = 1
	ndev.Parity = "N"
	//	saint, _ := strconv.Atoi(ele["StartingAddress"])
	//ndev.StartingAddress = 2
	//	qint, _ := strconv.Atoi(ele["Quantity"])
	//ndev.Quantity = 22
	/***********************初始化设备的特有的参数*****************************/
	return ndev, nil
}

// GetElement ..
func (d *TOSHIBA) GetElement() (dict, error) {
	conn := dict{
		"devaddr": d.devaddr,
		/***********************设备的特有的参数*****************************/
		"commif": d.commif,
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
func (d *TOSHIBA) HelpDoc() interface{} {
	conn := dict{
		"devaddr": "设备地址",
		/***********TOSHIBA设备的参数*****************************/
		"commif": "通信接口,比如(rs485-1)",
		/***********TOSHIBA设备的参数*****************************/
	}
	rParameter := dict{
		"_devid": "被读取设备对象的id",
		/***********读取设备的参数*****************************/
		/***********读取设备的参数*****************************/
	}

	wParameter := dict{
		"_devid": "被操作设备对象的id",
		/***********操作设备的参数*****************************/
		"_varname.1": "运行模式设置",
	}
	data := dict{
		"_devid": "添加设备对象的id",
		"_type":  "TOSHIBA", //设备类型
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

// CheckKey ..
/***************************************添加设备参数检验**********************************************/
func (d *TOSHIBA) CheckKey(ele dict) (bool, error) {
	_, sbOk := ele["subAddr"].(json.Number)
	if !sbOk {
		return false, errors.New("TOSHIBA device must have int type element 室内外机编号: subAddr")
	}
	_, mtOk := ele["mtype"].(string)
	if !mtOk {
		return false, errors.New("TOSHIBA device must have  element 室内还是室外机{inM|outM}): mtype")
	}
	return true, nil
}

/***************************************添加设备参数检验**********************************************/

func (d *TOSHIBA) encode(m dict) (json.Number, error) {
	name, _ := m["_varname"]
	var results json.Number = "0"
	switch name {
	case "xxxxx":
		{
			if val, ok := m["_varvalue"]; ok {
				if sval, ok := val.(string); ok {
					results = json.Number(sval)
					log.Debugln("xxxx = ", results)
				}
			}
		}
	default:
		{
			return json.Number("0"), errors.New("错误的_varname")
		}
	}

	return results, nil
}

func (d *TOSHIBA) hex2float32(hex2 []byte) (vf float32, err error) {
	if len(hex2) != 2 {
		return 0, errors.New("wrong len hex data")
	}
	var m int16
	e := (hex2[0] >> 3) & 0xf
	m = int16(hex2[0]<<5)<<3 + int16(hex2[1])
	if e >= 0 {
		if hex2[0]&0x80 == 0x80 {
			m = m - 0x800
		}
		log.Debugf("e=%d,m=%d", e, m)
		vf = float32(int32(m)<<e) * 0.01
		log.Debugf("vf=%f", vf)
	} // else {
	//vf = float32(m)
	//for i := e; i <= 0; i++ {
	//vf = vf / 2
	//}
	//}
	return vf, err
}

/***************************************读写接口实现**************************************************/

// RWDevValue ..
func (d *TOSHIBA) RWDevValue(rw string, m dict) (ret dict, err error) {
	ret = make(dict)
	defer func() {
		if driveErr := recover(); driveErr != nil {
			log.Errorf("drive programer  error : (%s)", driveErr)
			errstr := fmt.Sprintf("drive programer  error : %s", driveErr)
			err = errors.New(errstr)
		}
	}()
	d.mutex.Lock()
	defer d.mutex.Unlock()
	//log.SetLevel(log.DebugLevel)
	ret["_devid"] = d.devid
	if rw == "r" {

		test := []byte{0x0b, 0xb4}
		temperature, _ := d.hex2float32(test)
		log.Debugf("temperature=%f", temperature)
		/****************************************write device**********************************************/
	} else {
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
