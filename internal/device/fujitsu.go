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

//var mutex sync.Mutex
type FUJITSU struct {
	//继承于ModebusRtu
	ModbusRtu
	/**************按不同设备自定义*************************/

	/**************按不同设备自定义*************************/
}

func init() {
	RegDevice["FUJITSU"] = &FUJITSU{}
}

func (d *FUJITSU) NewDev(id string, ele map[string]string) (DeviceRWer, error) {
	ndev := new(FUJITSU)
	ndev.Device = d.Device.NewDev(id, ele)
	/***********************初始化设备的特有的参数*****************************/
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
		/***********************设备的特有的参数*****************************/
		"devaddr": d.devaddr,
		"commif":  d.commif,
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
		"commif": "通信接口,比如(rs485-1)",
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

	return true, nil
}

/***************************************添加设备参数检验**********************************************/
func (d *FUJITSU) getnm(inf interface{}) uint16 {
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
		k, _ := m["_varname"]
		switch k {
		case "室内机":
			{
				d.Quantity = 20
				d.Function_code = 4
				var addr uint16
				if dno, ok := m["_varvalue"]; ok {
					addr = d.getnm(dno)
					if 1 > addr || addr > 128 {
						return nil, errors.New("室内机地址参数错误")
					}
					d.Starting_address = 60*(addr-1) + 51
					log.Debugln("start_address=", d.Starting_address)
					bmdict, berr := d.ModbusRtu.RWDevValue("r", nil)
					if berr == nil {
						btdl := bmdict["Modbus-value"]
						bdl, _ := btdl.([]int)
						log.Debugf("室内机-%d receive data = %d", addr, bdl)
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
			{
				return nil, errors.New("缺少参数")
			}
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
						addr = d.getnm(dno)
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
