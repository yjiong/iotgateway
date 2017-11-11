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
type XP_YSHY struct {
	//继承于ModebusRtu
	ModbusRtu
	/**************按不同设备自定义*************************/

	/**************按不同设备自定义*************************/
}

func init() {
	RegDevice["XP_YSHY"] = &XP_YSHY{}
}

func (d *XP_YSHY) NewDev(id string, ele map[string]string) (DeviceRWer, error) {
	ndev := new(XP_YSHY)
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
	ndev.Quantity = 12
	/***********************初始化设备的特有的参数*****************************/
	return ndev, nil
}

func (d *XP_YSHY) GetElement() (dict, error) {
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
func (d *XP_YSHY) HelpDoc() interface{} {
	conn := dict{
		"devaddr": "设备地址",
		/***********XP_YSHY设备的参数*****************************/
		"commif":   "通信接口,比如(rs485-1)",
		"BaudRate": "波特率,比如(9600)",
		"DataBits": "数据位,比如(8)",
		"Parity":   "校验,(N,E,O)",
		"StopBits": "停止位,比如(1)",
		/***********XP_YSHY设备的参数*****************************/
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
		"_type":  "XP_YSHY", //设备类型
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
func (d *XP_YSHY) CheckKey(ele dict) (bool, error) {

	if _, ok := ele["BaudRate"].(json.Number); !ok {
		return false, errors.New(fmt.Sprintf("XP_YSHY device must have int type element 波特率 :BaudRate"))
	}
	if _, ok := ele["DataBits"].(json.Number); !ok {
		return false, errors.New(fmt.Sprintf("XP_YSHY device must have int type element 数据位 :DataBits"))
	}
	if _, ok := ele["StopBits"].(json.Number); !ok {
		return false, errors.New(fmt.Sprintf("XP_YSHY device must have int type element 停止位 :StopBits"))
	}
	if _, ok := ele["Parity"].(string); !ok {
		return false, errors.New(fmt.Sprintf("XP_YSHY device must have string type element 校验 :Parity"))
	}
	return true, nil
}

/***************************************添加设备参数检验**********************************************/

/***************************************读写接口实现**************************************************/
func (d *XP_YSHY) RWDevValue(rw string, m dict) (ret dict, err error) {
	ret = make(dict)
	ret["_devid"] = d.devid
	status := map[bool]string{
		false: "停止",
		true:  "运行",
	}
	fault := map[bool]string{
		false: "正常",
		true:  "故障",
	}
	warning := map[bool]string{
		false: "正常",
		true:  "报警",
	}
	remote_local := map[bool]string{
		false: "就地",
		true:  "远程",
	}
	all_open := map[bool]string{
		false: "未全开",
		true:  "全开到位",
	}
	all_closs := map[bool]string{
		false: "未全关",
		true:  "全关到位",
	}
	mdict, err := d.ModbusRtu.RWDevValue("r", nil)
	if err == nil {
		tdl := mdict["Modbus-value"]
		dl, ok := tdl.([]int)
		log.Info(dl)
		if ok {
			ret["1#取水泵运行"] = status[dl[22]&0x01 > 0]
			ret["1#取水泵过载"] = fault[dl[22]&0x02 > 0]
			ret["1#取水泵远控"] = remote_local[dl[22]&0x04 > 0]
			ret["2#取水泵运行"] = status[dl[22]&0x08 > 0]
			ret["2#取水泵过载"] = fault[dl[22]&0x10 > 0]
			ret["2#取水泵远控"] = remote_local[dl[22]&0x20 > 0]
			ret["补水泵运行"] = status[dl[22]&0x40 > 0]
			ret["补水泵过载"] = fault[dl[22]&0x80 > 0]
			ret["补水泵远控"] = remote_local[dl[23]&0x01 > 0]
			ret["补水阀全开信号"] = all_open[dl[23]&0x02 > 0]
			ret["补水阀全关信号"] = all_closs[dl[23]&0x04 > 0]
			ret["补水阀自动"] = remote_local[dl[23]&0x08 > 0]
			ret["蓄水池低液位"] = warning[dl[23]&0x10 > 0]
			ret["河道低液位"] = warning[dl[23]&0x20 > 0]
			ret["蓄水池液位"] = fmt.Sprintf("%0.2f%s", float64(dl[0]*0x100+dl[1])/100.0, "米")
			ret["清水池液位"] = fmt.Sprintf("%0.2f%s", float64(dl[2]*0x100+dl[3])/100.0, "米")
			ret["实时流量"] = fmt.Sprintf("%0.2f%s", float64(dl[4]*0x100+dl[5])/100.0, "立方/小时")
			ret["累计流量"] = fmt.Sprintf("%0.1f%s", float64(dl[6]*0x1000000+dl[7]*0x10000+dl[8]*0x100+dl[9])/10.0, "立方/小时")
			ret["水表流量"] = fmt.Sprintf("%0.1f%s", float64(dl[14]*0x1000000+dl[15]*0x10000+dl[16]*0x100+dl[17])/10.0, "立方/小时")
			log.Info(ret)
		} else {
			ret["_status"] = "offline"
		}
	}
	return ret, err
}

/***************************************读写接口实现**************************************************/
