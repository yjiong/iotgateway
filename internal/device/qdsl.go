package device

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	//	"sync"
	log "github.com/Sirupsen/logrus"
)

//var mutex sync.Mutex
type QDSL struct {
	//继承于ModebusRtu
	ModbusRtu
	/**************按不同设备自定义*************************/

	/**************按不同设备自定义*************************/
}

func init() {
	RegDevice["QDSL"] = &QDSL{}
}

func (d *QDSL) NewDev(id string, ele map[string]string) (DeviceRWer, error) {
	ndev := new(QDSL)
	ndev.Device = d.Device.NewDev(id, ele)
	/***********************初始化设备的特有的参数*****************************/
	ndev.BaudRate, _ = strconv.Atoi(ele["BaudRate"])
	ndev.DataBits, _ = strconv.Atoi(ele["DataBits "])
	ndev.StopBits, _ = strconv.Atoi(ele["StopBits"])
	ndev.Parity, _ = ele["Parity"]
	//ndev.Function_code = 3
	//	saint, _ := strconv.Atoi(ele["Starting_address"])
	//ndev.Starting_address = 0
	//	qint, _ := strconv.Atoi(ele["Quantity"])
	//ndev.Quantity = 12
	/***********************初始化设备的特有的参数*****************************/
	return ndev, nil
}

func (d *QDSL) GetElement() (dict, error) {
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
func (d *QDSL) HelpDoc() interface{} {
	conn := dict{
		"devaddr": "设备地址",
		/***********QDSL设备的参数*****************************/
		"commif":   "通信接口,比如(rs485-1)",
		"BaudRate": "波特率,比如(9600)",
		"DataBits": "数据位,比如(8)",
		"Parity":   "校验,(N,E,O)",
		"StopBits": "停止位,比如(1)",
		/***********QDSL设备的参数*****************************/
	}
	r_parameter := dict{
		"_devid": "被读取设备对象的id",
		/***********读取设备的参数*****************************/
		/***********读取设备的参数*****************************/
	}
	w_parameter := dict{
		"_devid": "被操作设备对象的id",
		/***********操作设备的参数*****************************/
		/***********操作设备的参数*****************************/
	}
	data := dict{
		"_devid": "添加设备对象的id",
		"_type":  "QDSL", //设备类型
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
func (d *QDSL) CheckKey(ele dict) (bool, error) {

	if _, ok := ele["BaudRate"].(json.Number); !ok {
		return false, errors.New(fmt.Sprintf("QDSL device must have int type element 波特率 :BaudRate"))
	}
	if _, ok := ele["DataBits"].(json.Number); !ok {
		return false, errors.New(fmt.Sprintf("QDSL device must have int type element 数据位 :DataBits"))
	}
	if _, ok := ele["StopBits"].(json.Number); !ok {
		return false, errors.New(fmt.Sprintf("QDSL device must have int type element 停止位 :StopBits"))
	}
	if _, ok := ele["Parity"].(string); !ok {
		return false, errors.New(fmt.Sprintf("QDSL device must have string type element 校验 :Parity"))
	}
	return true, nil
}

/***************************************添加设备参数检验**********************************************/

/***************************************读写接口实现**************************************************/
func (d *QDSL) RWDevValue(rw string, m dict) (ret dict, err error) {
	ret = make(dict)
	ret["_devid"] = d.devid
	if rw == "r" {
		d.Starting_address = 330
		d.Quantity = 32
		mdict, err := d.ModbusRtu.RWDevValue("r", nil)
		if err == nil {
			tdl := mdict["Modbus-value"]
			dl, ok := tdl.([]int)
			log.Info(dl)
			if ok {
				//				ret["1#取水泵运行"] = status[dl[22]&0x01 > 0]
				log.Info(ret)
			}
		} else {
			ret["status"] = "offline"
		}
	}
	return ret, err
}

/***************************************读写接口实现**************************************************/
