package device

import (
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	simplejson "github.com/bitly/go-simplejson"
	"strconv"
	//"strings"
	//"sync"
)

var MASTER_SLAVE = map[int]string{
	0: "主机",
	1: "副机",
}
var RUN_STATUS = map[int]string{
	1: "自动",
	2: "制冷",
	3: "干燥",
	4: "加热",
	5: "风扇",
}
var ON_OFF = map[int]string{
	1: "停止",
	2: "运行",
}
var ARIFLOW_STATUS = map[int]string{
	1: "自动",
	2: "安静",
	3: "低",
	4: "中",
	5: "高",
	6: "中低",
	7: "中高",
}
var MALFUNCTION = map[int]string{
	0: "无故障",
	1: "故障",
}
var VERTICAL_HORIZONTAL = map[int]string{
	1: "摆动",
	2: "位置1",
	3: "位置2",
	4: "位置3",
	5: "位置4",
	6: "位置5",
}

type FUJITSU struct {
	//继承于ModebusRtu
	ModbusRtu
	/**************按不同设备自定义*************************/
	sub_addr string
	mtype    string
	/**************按不同设备自定义*************************/
}

func init() {
	RegDevice["FUJITSU"] = &FUJITSU{}
}

func (d *FUJITSU) NewDev(id string, ele map[string]string) (DeviceRWer, error) {
	ndev := new(FUJITSU)
	ndev.Device = d.Device.NewDev(id, ele)
	/***********************初始化设备的特有的参数*****************************/
	ndev.sub_addr = ele["sub_addr"]
	ndev.mtype = ele["mtype"]
	ndev.BaudRate = 19200
	ndev.DataBits = 8
	ndev.StopBits = 1
	ndev.Parity = "E"
	//	saint, _ := strconv.Atoi(ele["Starting_address"])
	//ndev.Starting_address = 2
	//	qint, _ := strconv.Atoi(ele["Quantity"])
	//ndev.Quantity = 22
	/***********************初始化设备的特有的参数*****************************/
	return ndev, nil
}

func (d *FUJITSU) GetElement() (dict, error) {
	conn := dict{
		"devaddr": d.devaddr,
		/***********************设备的特有的参数*****************************/
		"mtype":    d.mtype,
		"sub_addr": d.sub_addr,
		"commif":   d.commif,
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
func (d *FUJITSU) HelpDoc() interface{} {
	conn := dict{
		"devaddr": "设备地址",
		/***********FUJITSU设备的参数*****************************/
		"commif":   "通信接口,比如(rs485-1)",
		"sub_addr": "子地址",
		/***********FUJITSU设备的参数*****************************/
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
		"_type":  "FUJITSU", //设备类型
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
func (d *FUJITSU) CheckKey(ele dict) (bool, error) {
	_, sb_ok := ele["sub_addr"].(json.Number)
	if !sb_ok {
		return false, errors.New(fmt.Sprintf("FUJITSU device must have int type element 室内外机编号:sub_addr"))
	}
	_, mt_ok := ele["mtype"].(string)
	if !mt_ok {
		return false, errors.New(fmt.Sprintf("FUJITSU device must have  element 室内还是室外机{in_m|out_m}):mtype"))
	}
	return true, nil
}

/***************************************添加设备参数检验**********************************************/

func (d *FUJITSU) inside_status(ret dict, iarray []int) {
	ret["VRF地址"] = iarray[1]
	ret["主副机信息"] = MASTER_SLAVE[iarray[5]]
	ret["运行模式状态"] = RUN_STATUS[iarray[7]]
	ret["运行开关状态"] = ON_OFF[iarray[9]]
	ret["设置温度状态"] = (iarray[10]*0x100 + iarray[11]) / 4
	ret["气流状态"] = ARIFLOW_STATUS[iarray[13]]
	ret["室内温度状态"] = (iarray[14]*0x100 + iarray[15]) / 4
	ret["故障监控"] = MALFUNCTION[iarray[17]]
	ret["垂直空气方向位置状态"] = VERTICAL_HORIZONTAL[iarray[19]]
	ret["水平空气方向位置状态"] = VERTICAL_HORIZONTAL[iarray[21]]
	ret["遥控器运行禁止是遏制状态"] = iarray[1]
}

/***************************************读写接口实现**************************************************/
func (d *FUJITSU) RWDevValue(rw string, m dict) (ret dict, err error) {
	ret = make(dict)
	defer func() {
		if drive_err := recover(); drive_err != nil {
			log.Errorf("drive programer  error : %s", drive_err)
			errstr := fmt.Sprintf("drive programer  error : %s", drive_err)
			err = errors.New(errstr)
		}
	}()
	d.mutex.Lock()
	defer d.mutex.Unlock()
	//log.SetLevel(log.DebugLevel)
	ret["_devid"] = d.devid
	if rw == "r" {
		if d.mtype == "in_m" {
			d.Quantity = 25
			d.Function_code = 4
			var addr uint16
			if dno, err := strconv.Atoi(d.sub_addr); err == nil {
				addr = uint16(dno)
				if 1 > addr || addr > 128 {
					return nil, errors.New("室内机地址参数错误")
				}
				d.Starting_address = 60*(addr-1) + 50
				log.Debugln("start_address=", d.Starting_address)
				bmdict, berr := d.ModbusRtu.RWDevValue("r", nil)
				if berr == nil {
					btdl := bmdict["Modbus-value"]
					bdl, _ := btdl.([]int)
					log.Debugf("室内机-%d receive data = %d", addr, bdl)
					d.inside_status(ret, bdl)

				} else {
					ret["error"] = err.Error()
					log.Debugln(ret)
					return ret, nil
				}
			}
		} else if d.mtype == "out_m" {

		}
	} else {
		var method func(dict) (dict, error)
		if k, ok := m["_varname"]; ok {
			switch k {
			case "写单一内机":
				{
					d.Quantity = 20
					d.Function_code = 16
					var addr uint16
					if dno, ok := m["_varvalue"]; ok {
						addr = getnm(dno)
						d.Starting_address = 60*(addr-1) + 2
						log.Debugln("start_address=", d.Starting_address)

						bmdict, berr := d.ModbusRtu.RWDevValue("w", nil)
						if berr == nil {
							btdl := bmdict["Modbus-value"]
							bdl, _ := btdl.([]int)
							log.Debugf("写单一内机-%d receive data = %d", addr, bdl)
						} else {
							ret["error"] = err.Error()
							log.Debugln(ret)
							return ret, nil
						}
					} else {
						return nil, errors.New("地址参数错误")
					}
				}
			default:
				return nil, errors.New("错误的_varname")
			}
		}
		var wval dict
		wval, err = method(m)
		if err == nil {
			log.Debugln("send modbus data =", wval)
			ret, err = d.ModbusRtu.RWDevValue("w", wval)
		}
	}
	jsret, _ := json.Marshal(ret)
	inforet, _ := simplejson.NewJson(jsret)
	pinforet, _ := inforet.EncodePretty()
	log.Info(string(pinforet))
	if err != nil {
		log.Debugln(err)
	}
	return ret, err
}

func getnm(inf interface{}) uint16 {
	var ruint uint16 = 0
	if nm, ok := inf.(json.Number); ok {
		nmi64, _ := nm.Int64()
		ruint = uint16(nmi64)
	} else {
		if nm, ok := inf.(string); ok {
			snm, _ := strconv.Atoi(nm)
			ruint = uint16(snm)
		}
	}
	return ruint
}
