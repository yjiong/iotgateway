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
type Ximengzi struct {
	//继承于Device
	ModbusRtu
	/**************按不同设备自定义*************************/

	/**************按不同设备自定义*************************/
}

func init() {
	RegDevice["Ximengzi"] = &Ximengzi{}
}

func (d *Ximengzi) NewDev(id string, ele map[string]string) (DeviceRWer, error) {
	ndev := new(Ximengzi)
	ndev.Device = d.Device.NewDev(id, ele)
	/***********************初始化设备的特有的参数*****************************/
	ndev.BaudRate, _ = strconv.Atoi(ele["BaudRate"])
	ndev.DataBits, _ = strconv.Atoi(ele["DataBits "])
	ndev.StopBits, _ = strconv.Atoi(ele["StopBits"])
	ndev.Parity, _ = ele["Parity"]
	ndev.Function_code = 3
	//	saint, _ := strconv.Atoi(ele["Starting_address"])
	ndev.Starting_address = 0
	//	qint, _ := strconv.Atoi(ele["Quantity"])
	ndev.Quantity = 10
	/***********************初始化设备的特有的参数*****************************/
	return ndev, nil
}

func (d *Ximengzi) GetElement() (dict, error) {
	conn := dict{
		/***********************设备的特有的参数*****************************/
		"devaddr":          d.devaddr,
		"commif":           d.commif,
		"BaudRate":         d.BaudRate,
		"DataBits":         d.DataBits,
		"StopBits":         d.StopBits,
		"Parity":           d.Parity,
		"Function_code":    d.Function_code,
		"Starting_address": d.Starting_address,
		"Quantity":         d.Quantity,
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
func (d *Ximengzi) HelpDoc() interface{} {
	conn := dict{
		"devaddr": "设备地址",
		/***********Ximengzi设备的参数*****************************/
		"commif":           "通信接口,比如(rs485-1)",
		"BaudRate":         "波特率,比如(9600)",
		"DataBits":         "数据位,比如(8)",
		"Parity":           "校验,(N,E,O)",
		"StopBits":         "停止位,比如(1)",
		"Function_code":    "modbus功能码 : (1,2,3,4,5,6,15,16)",
		"Starting_address": "操作起始地址,uint类型",
		"Quantity":         "寄存器数量,uint类型",
		/***********Ximengzi设备的参数*****************************/
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
		"Function_code":    "modbus功能码 : (5,6,15,16)",
		"Starting_address": "操作起始地址,uint类型",
		"Quantity":         "寄存器数量,uint类型",
		"value":            "要写入modbus设备的值,功能码为5和6时,值为uint16,功能码为15,16时,值为 [uint8...]",
		/***********操作设备的参数*****************************/
	}
	data := dict{
		"_devid": "添加设备对象的id",
		"_type":  "Ximengzi", //设备类型
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
func (d *Ximengzi) CheckKey(ele dict) (bool, error) {

	if _, ok := ele["BaudRate"].(json.Number); !ok {
		return false, errors.New(fmt.Sprintf("Ximengzi device must have int type element 波特率 :BaudRate"))
	}
	if _, ok := ele["DataBits"].(json.Number); !ok {
		return false, errors.New(fmt.Sprintf("Ximengzi device must have int type element 数据位 :DataBits"))
	}
	if _, ok := ele["StopBits"].(json.Number); !ok {
		return false, errors.New(fmt.Sprintf("Ximengzi device must have int type element 停止位 :StopBits"))
	}
	if _, ok := ele["Parity"].(string); !ok {
		return false, errors.New(fmt.Sprintf("Ximengzi device must have string type element 校验 :Parity"))
	}
	return true, nil
}

/***************************************添加设备参数检验**********************************************/

/***************************************读写接口实现**************************************************/
func (d *Ximengzi) RWDevValue(rw string, m dict) (ret dict, err error) {
	ret = make(dict)
	status := map[int]string{
		0: "启动",
		1: "停止",
	}
	mdict, err := d.ModbusRtu.RWDevValue("r", nil)
	if err == nil {
		tdl := mdict["Modbus-value"]
		dl, ok := tdl.([]int)
		log.Info(dl)
		if ok {
			ret["电机启动"] = status[dl[0]]
			ret["流量"] = float64(dl[2]*0x100 + dl[3])
			log.Info(ret)

		}
	}
	return ret, err
}

/***************************************读写接口实现**************************************************/
