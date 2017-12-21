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

var MASTERSLAVE = map[int]string{
	0: "主机",
	1: "副机",
}
var RUNSTATUS = map[int]string{
	0: "不变",
	1: "自动",
	2: "制冷",
	3: "除湿",
	4: "制热",
	5: "送风",
}
var ONOFF = map[int]string{
	0: "不变",
	1: "停止",
	2: "运行",
}
var ARIFLOWSTATUS = map[int]string{
	0: "不变",
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
var VERTICALHORIZONTAL = map[int]string{
	0: "不变",
	1: "摆动",
	2: "位置1",
	3: "位置2",
	4: "位置3",
	5: "位置4",
	6: "位置5",
}
var REMOTESET = map[int]string{
	0: "不抑制",
	1: "抑制",
}
var FILTERSTATUS = map[int]string{
	0: "无标志",
	1: "过滤网标志",
}
var NORMALENEREGYSAVE = map[int]string{
	0: "无变化",
	1: "正常运行",
	2: "节能运行",
}
var NORMALFANGDONG = map[int]string{
	1: "正常运行",
	2: "防冻液运行",
}
var NORMALSPACIAL = map[int]string{
	0: "特殊状态",
	1: "正常状态",
}
var CHUSHUANG = map[int]string{
	0: "无除霜状态",
	1: "除霜状态",
}
var YOUHUISHOU = map[int]string{
	0: "无油回收状态",
	1: "油回收状态",
}
var BENGGUZHANG = map[int]string{
	0: "无泵故障状态",
	1: "泵故障状态",
}
var WAIBUGUANRE = map[int]string{
	0: "无变化",
	1: "释放",
	2: "关热",
}

var SETFANGDONG = map[int]string{
	0: "无变化",
	1: "释放",
	2: "防冻液运行",
}

//outM
var LOWNOISE = map[int]string{
	0: "性能优先无效",
	1: "性能优先有效",
}
var LOWNOISELEVEL = map[int]string{
	0: "释放",
	1: "第1级",
	2: "第2级",
	3: "第3级",
}
var LOWNOISESET = map[int]string{
	1:  "非性能优先0",
	5:  "非性能优先1",
	9:  "非性能优先2",
	13: "非性能优先3",
	3:  "性能优先0",
	7:  "性能优先1",
	11: "性能优先2",
	15: "性能优先3",
}
var EDRLJSYX = map[int]string{
	1: "释放",
	2: "100%",
	3: "90%",
	4: "80%",
	5: "70%",
	6: "60%",
	7: "50%",
	8: "40%",
}

//allInM
var ALLONOFF = map[int]string{
	0: "不变",
	1: "所有室内机均停止",
	2: "有些室内机正在运行",
}
var ALLINMALFUNCTION = map[int]string{
	0: "所有室内机无故障",
	1: "有些室内机处于故障状态",
}
var JINGJITINGZHI = map[int]string{
	0: "不变",
	1: "释放请求",
	2: "紧急停止请求",
}

// FUJITSU ..
type FUJITSU struct {
	//继承于ModebusRtu
	ModbusRtu
	/**************按不同设备自定义*************************/
	subAddr string
	mtype   string
	/**************按不同设备自定义*************************/
}

func init() {
	RegDevice["FUJITSU"] = &FUJITSU{}
}

// NewDev ..
func (d *FUJITSU) NewDev(id string, ele map[string]string) (Devicerwer, error) {
	ndev := new(FUJITSU)
	ndev.Device = d.Device.NewDev(id, ele)
	/***********************初始化设备的特有的参数*****************************/
	ndev.subAddr = ele["subAddr"]
	ndev.mtype = ele["mtype"]
	ndev.BaudRate = 19200
	ndev.DataBits = 8
	ndev.StopBits = 1
	ndev.Parity = "E"
	//	saint, _ := strconv.Atoi(ele["StartingAddress"])
	//ndev.StartingAddress = 2
	//	qint, _ := strconv.Atoi(ele["Quantity"])
	//ndev.Quantity = 22
	/***********************初始化设备的特有的参数*****************************/
	return ndev, nil
}

// GetElement ..
func (d *FUJITSU) GetElement() (dict, error) {
	conn := dict{
		"devaddr": d.devaddr,
		/***********************设备的特有的参数*****************************/
		"mtype":   d.mtype,
		"subAddr": d.subAddr,
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

// HelpDoc ..
func (d *FUJITSU) HelpDoc() interface{} {
	conn := dict{
		"devaddr": "设备地址",
		/***********FUJITSU设备的参数*****************************/
		"commif":  "通信接口,比如(rs485-1)",
		"mtype":   "(inM|outM|allInM|VRF)",
		"subAddr": "子地址",
		/***********FUJITSU设备的参数*****************************/
	}
	rParameter := dict{
		"_devid": "被读取设备对象的id",
		/***********读取设备的参数*****************************/
		/***********读取设备的参数*****************************/
	}

	wParameter := dict{
		"_devid": "被操作设备对象的id",
		/***********操作设备的参数*****************************/
		"_varname.1":   "运行模式设置",
		"_varvalue.1":  "(自动|制冷|除湿|制热|送风)",
		"_varname.2":   "运行模式设置",
		"_varvalue.2":  "(自动|制冷|除湿|制热|送风)",
		"_varname.3":   "运行开关设置",
		"_varvalue.3":  "(停止|运行)",
		"_varname.4":   "设置温度设定值",
		"_varvalue.4":  "0-63",
		"_varname.5":   "气流设置",
		"_varvalue.5":  "(自动|安静|低|中|高|中低|中高)",
		"_varname.6":   "垂直空气方向位置状态",
		"_varvalue.6":  "(摆动|位置1|位置2|位置3|位置4)",
		"_varname.7":   "水平空气方向位置状态",
		"_varvalue.7":  "(摆动|位置1|位置2|位置3|位置4|位置5)",
		"_varname.8":   "遥控器运行禁止设置",
		"_varvalue.8":  "(允许|禁止)",
		"_varname.9":   "过滤网标志重置",
		"_varvalue.9":  "(重置)",
		"_varname.10":  "经济运行模式设置",
		"_varvalue.10": "(正常运行|节能运行)",
		"_varname.11":  "防冻液运行设置",
		"_varvalue.11": "(释放|防冻液运行)",
		"_varname.12":  "制冷/干燥温度上限设置",
		"_varvalue.12": "0-63",
		"_varname.13":  "制冷/干燥温度下限设置",
		"_varvalue.13": "0-63",
		"_varname.14":  "加热温度上限设置",
		"_varvalue.14": "0-63",
		"_varname.15":  "加热温度下限设置",
		"_varvalue.15": "0-63",
		"_varname.16":  "自动温度上限设置",
		"_varvalue.16": "0-63",
		"_varname.17":  "自动温度下限设置",
		"_varvalue.17": "0-63",
		"_varname.18":  "外部关热设置",
		"_varvalue.18": "(释放|关热)",
		"_varname.19":  "紧急停止",
		"_varvalue.19": "(释放请求|紧急停止请求)",
		"_varname.20":  "室外机低噪音运行设置",
		"_varvalue.20": "(性能优先0|性能优先1|性能优先2|性能优先3|非性能优先0|非性能优先1|非性能优先2|非性能优先3)",
		"_varname.21":  "室外机额定容量节省指令",
		"_varvalue.21": "(释放|40%|50%|60%|70%|80%|90%|100%)",
		"解释": "one cmd set one value," +
			"inM(_varname.1----_varname.18)," +
			"outM(_varname.20----_varname.21)," +
			"allInM(_varname.1----_varname.5," +
			"_varname.8----_varname.19)," +
			"VRF(_varname.19)",
		/***********操作设备的参数*****************************/
	}
	data := dict{
		"_devid": "添加设备对象的id",
		"_type":  "FUJITSU", //设备类型
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
func (d *FUJITSU) CheckKey(ele dict) (bool, error) {
	_, sbOk := ele["subAddr"].(json.Number)
	if !sbOk {
		return false, errors.New("FUJITSU device must have int type element 室内外机编号: subAddr")
	}
	_, mtOk := ele["mtype"].(string)
	if !mtOk {
		return false, errors.New("FUJITSU device must have  element 室内还是室外机{inM|outM}): mtype")
	}
	return true, nil
}

/***************************************添加设备参数检验**********************************************/

func (d *FUJITSU) insideStatus(ret dict, iarray []int) {
	ret["VRF地址"] = fmt.Sprintf("%d-%d", iarray[1], iarray[0])
	ret["主副机信息"] = MASTERSLAVE[iarray[5]]
	ret["运行模式状态"] = RUNSTATUS[iarray[7]]
	ret["运行开关状态"] = ONOFF[iarray[9]]
	ret["设置温度状态"] = (iarray[10]*0x100 + iarray[11]) / 4
	ret["气流状态"] = ARIFLOWSTATUS[iarray[13]]
	ret["室内温度状态"] = (iarray[14]*0x100 + iarray[15]) / 4
	ret["故障监控"] = MALFUNCTION[iarray[17]]
	ret["垂直空气方向位置状态"] = VERTICALHORIZONTAL[iarray[19]]
	ret["水平空气方向位置状态"] = VERTICALHORIZONTAL[iarray[21]]
	ret["遥控器运行禁止设置状态"] = map[string]string{
		"全部运行设置":    REMOTESET[iarray[23]&0x01],
		"定时器设置":     REMOTESET[iarray[23]&0x02/0x02],
		"室温设置":      REMOTESET[iarray[23]&0x04/0x04],
		"运行模式设置":    REMOTESET[iarray[23]&0x08/0x08],
		"启动/停止运行设置": REMOTESET[iarray[23]&0x10/0x10],
		"启动行设置":     REMOTESET[iarray[23]&0x20/0x20],
		"过滤网重置运行":   REMOTESET[iarray[23]&0x40/0x40],
	}
	ret["过滤网标志状态"] = FILTERSTATUS[iarray[25]]
	ret["经济模式运行状态"] = NORMALENEREGYSAVE[iarray[27]]
	ret["防冻液运行状态"] = NORMALFANGDONG[iarray[29]]
	ret["温度上下限设置状态(制冷/干燥)"] = fmt.Sprintf("上限=%0.1f,下限=%0.1f", float64(iarray[31])/2, float64(iarray[30])/2)
	ret["温度上下限设置状态(加热)"] = fmt.Sprintf("上限=%0.1f,下限=%0.1f", float64(iarray[33])/2, float64(iarray[32])/2)
	ret["温度上下限设置状态(自动)"] = fmt.Sprintf("上限=%0.1f,下限=%0.1f", float64(iarray[35])/2, float64(iarray[34])/2)
	ret["室内机状态"] = map[string]string{
		"正常状态": NORMALSPACIAL[iarray[37]&0x01],
		"除霜":   CHUSHUANG[iarray[37]&0x02/0x02],
		"油回收":  YOUHUISHOU[iarray[37]&0x04/0x04],
		"泵故障":  BENGGUZHANG[iarray[37]&0x08/0x08],
	}
	ret["外部关热状态"] = WAIBUGUANRE[iarray[39]]
}
func (d *FUJITSU) outsideStatus(ret dict, iarray []int) {
	ret["室外机低噪音运行状态监控"] = map[string]string{
		"性能优先": LOWNOISE[iarray[5]&0x01],
		"级别":   LOWNOISELEVEL[iarray[5]>>1],
	}
	ret["室外机额定容量节省运行监控"] = EDRLJSYX[iarray[7]]
	ret["主副机信息"] = MASTERSLAVE[iarray[3]]
	ret["VRF地址"] = fmt.Sprintf("%d-%d", iarray[1], iarray[0])

}

func (d *FUJITSU) allInStatus(ret dict, iarray []int) {
	ret["所有室内机故障监控"] = ALLINMALFUNCTION[iarray[1]]
	ret["所有室内机开/关状态"] = ALLONOFF[iarray[3]]
}

func (d *FUJITSU) encode(m dict) (json.Number, error) {
	name, _ := m["_varname"]
	var results json.Number = "0"
	if d.mtype == "outM" {
		if !(name == "室外机低噪音运行设置" || name == "室外机额定容量节省指令") {
			return results, errors.New("_varname错误")
		}
	}
	if d.mtype == "VRF" {
		if name != "紧急停止" {
			return results, errors.New("_varname错误")
		}
	}
	switch name {
	case "运行模式设置":
		{
			if val, ok := m["_varvalue"]; ok {
				if sval, ok := val.(string); ok {
					for k, v := range RUNSTATUS {
						if v == sval {
							sk := strconv.Itoa(k)
							results = json.Number(sk)
							d.StartingAddress += 0
							log.Debugln("set 运行模式状态 = ", results)
						}
					}
				}
			}
		}
	case "运行开关设置":
		{
			if val, ok := m["_varvalue"]; ok {
				if sval, ok := val.(string); ok {
					for k, v := range ONOFF {
						if v == sval {
							results = json.Number(strconv.Itoa(k))
							d.StartingAddress++
							log.Debugln("set 运行开关设置 = ", results)
						}
					}
				}
			}
		}
	case "设置温度设定值":
		{
			if val, ok := m["_varvalue"]; ok {
				if sval, ok := val.(string); ok {
					if isvalm, err := strconv.Atoi(sval); err == nil {
						si := isvalm*8 + 1
						results = json.Number(strconv.Itoa(si))
						d.StartingAddress += 2
						log.Debugln("set 设置温度设定值 = ", results)
					}
				}
			}
		}
	case "气流设置":
		{
			if val, ok := m["_varvalue"]; ok {
				if sval, ok := val.(string); ok {
					for k, v := range ARIFLOWSTATUS {
						if v == sval {
							sk := strconv.Itoa(k)
							results = json.Number(sk)
							d.StartingAddress += 3
							log.Debugln("set 气流设置 = ", results)
						}
					}
				}
			}
		}
	case "垂直空气方向位置状态":
		{
			if val, ok := m["_varvalue"]; ok && d.mtype == "inM" {
				if sval, ok := val.(string); ok {
					for k, v := range VERTICALHORIZONTAL {
						if v == sval {
							sk := strconv.Itoa(k)
							results = json.Number(sk)
							d.StartingAddress += 4
							log.Debugln("set 垂直空气方向位置状态 = ", results)
						}
					}
				}
			} else {
				return results, errors.New("设置参数错误")
			}
		}
	case "水平空气方向位置状态":
		{
			if val, ok := m["_varvalue"]; ok && d.mtype == "inM" {
				if sval, ok := val.(string); ok {
					for k, v := range VERTICALHORIZONTAL {
						if v == sval {
							sk := strconv.Itoa(k)
							results = json.Number(sk)
							d.StartingAddress += 5
							log.Debugln("set 水平空气方向位置状态 = ", results)
						}
					}
				}
			} else {
				return results, errors.New("设置参数错误")
			}
		}
	case "遥控器运行禁止设置":
		{
			if val, ok := m["_varvalue"]; ok {
				if sval, ok := val.(string); ok {
					if sval == "允许" {
						results = json.Number("255")
					} else {
						results = json.Number("0")
					}
					d.StartingAddress += 6
					if d.mtype == "allInM" {
						d.StartingAddress -= 2
					}
					log.Debugln("set 遥控器运行禁止设置 = ", results)
				}
			}
		}
	case "过滤网标志重置":
		{
			if val, ok := m["_varvalue"]; ok && d.mtype == "inM" {
				if sval, ok := val.(string); ok {
					if sval == "重置" {
						results = json.Number("1")
					} else {
						results = json.Number("0")
					}
					d.StartingAddress += 7
					log.Debugln("set 过滤网标志重置 = ", results)
				}
			} else {
				return results, errors.New("设置参数错误")
			}
		}
	case "经济运行模式设置":
		{
			if val, ok := m["_varvalue"]; ok && d.mtype == "inM" {
				if sval, ok := val.(string); ok {
					for k, v := range NORMALENEREGYSAVE {
						if v == sval {
							sk := strconv.Itoa(k)
							results = json.Number(sk)
							d.StartingAddress += 8
							log.Debugln("set 经济运行模式设置 = ", results)
						}
					}
				}
			} else {
				return results, errors.New("设置参数错误")
			}
		}
	case "防冻液运行设置":
		{
			if val, ok := m["_varvalue"]; ok && d.mtype == "inM" {
				if sval, ok := val.(string); ok {
					for k, v := range SETFANGDONG {
						if v == sval {
							sk := strconv.Itoa(k)
							results = json.Number(sk)
							d.StartingAddress += 9
							log.Debugln("set 防冻液运行设置 = ", results)
						}
					}
				}
			} else {
				return results, errors.New("设置参数错误")
			}
		}
	case "制冷/干燥温度上限设置":
		{
			if val, ok := m["_varvalue"]; ok {
				if sval, ok := val.(string); ok {
					if isvalm, err := strconv.Atoi(sval); err == nil {
						si := isvalm*8 + 1
						results = json.Number(strconv.Itoa(si))
						d.StartingAddress += 10
						if d.mtype == "allInM" {
							d.StartingAddress -= 5
						}
						log.Debugln("set 制冷/干燥温度上限设置 = ", results)
					}
				}
			}
		}
	case "制冷/干燥温度下限设置":
		{
			if val, ok := m["_varvalue"]; ok {
				if sval, ok := val.(string); ok {
					if isvalm, err := strconv.Atoi(sval); err == nil {
						si := isvalm*8 + 1
						results = json.Number(strconv.Itoa(si))
						d.StartingAddress += 11
						if d.mtype == "allInM" {
							d.StartingAddress -= 5
						}
						log.Debugln("set 制冷/干燥温度下限设置 = ", results)
					}
				}
			}
		}
	case "加热温度上限设置":
		{
			if val, ok := m["_varvalue"]; ok {
				if sval, ok := val.(string); ok {
					if isvalm, err := strconv.Atoi(sval); err == nil {
						si := isvalm*8 + 1
						results = json.Number(strconv.Itoa(si))
						d.StartingAddress += 12
						if d.mtype == "allInM" {
							d.StartingAddress -= 5
						}
						log.Debugln("set 加热温度上限设置 = ", results)
					}
				}
			}
		}
	case "加热温度下限设置":
		{
			if val, ok := m["_varvalue"]; ok {
				if sval, ok := val.(string); ok {
					if isvalm, err := strconv.Atoi(sval); err == nil {
						si := isvalm*8 + 1
						results = json.Number(strconv.Itoa(si))
						d.StartingAddress += 13
						if d.mtype == "allInM" {
							d.StartingAddress -= 5
						}
						log.Debugln("set 加热温度下限设置 = ", results)
					}
				}
			}
		}
	case "自动温度上限设置":
		{
			if val, ok := m["_varvalue"]; ok {
				if sval, ok := val.(string); ok {
					if isvalm, err := strconv.Atoi(sval); err == nil {
						si := isvalm*8 + 1
						results = json.Number(strconv.Itoa(si))
						d.StartingAddress += 14
						if d.mtype == "allInM" {
							d.StartingAddress -= 5
						}
						log.Debugln("set 自动温度上限设置 = ", results)
					}
				}
			}
		}
	case "自动温度下限设置":
		{
			if val, ok := m["_varvalue"]; ok {
				if sval, ok := val.(string); ok {
					if isvalm, err := strconv.Atoi(sval); err == nil {
						si := isvalm*8 + 1
						results = json.Number(strconv.Itoa(si))
						d.StartingAddress += 15
						if d.mtype == "allInM" {
							d.StartingAddress -= 5
						}
						log.Debugln("set 自动温度下限设置 = ", results)
					}
				}
			}
		}
	case "外部关热设置":
		{
			if val, ok := m["_varvalue"]; ok {
				if sval, ok := val.(string); ok {
					for k, v := range WAIBUGUANRE {
						if v == sval {
							sk := strconv.Itoa(k)
							results = json.Number(sk)
							d.StartingAddress += 16
							if d.mtype == "allInM" {
								d.StartingAddress -= 5
							}
							log.Debugln("set 外部关热设置 = ", results)
						}
					}
				}
			}
		}
	case "紧急停止":
		{
			if val, ok := m["_varvalue"]; ok && (d.mtype == "allInM" || d.mtype == "VRF") {
				if sval, ok := val.(string); ok {
					for k, v := range JINGJITINGZHI {
						if v == sval {
							sk := strconv.Itoa(k)
							results = json.Number(sk)
							d.StartingAddress += 12
							if d.mtype == "VRF" {
								d.StartingAddress = 9210
							}
							log.Debugln("set 紧急停止说明 = ", results)
						}
					}
				}
			} else {
				return results, errors.New("设置参数错误")
			}
		}
	case "室外机低噪音运行设置":
		{
			if val, ok := m["_varvalue"]; ok && d.mtype == "outM" {
				if sval, ok := val.(string); ok {
					for k, v := range LOWNOISESET {
						if v == sval {
							sk := strconv.Itoa(k)
							results = json.Number(sk)
							d.StartingAddress += 0
							log.Debugln("set 室外机低噪音运行设置 = ", results)
						}
					}
				}
			} else {
				return results, errors.New("设置参数错误")
			}
		}
	case "室外机额定容量节省指令":
		{
			if val, ok := m["_varvalue"]; ok && d.mtype == "outM" {
				if sval, ok := val.(string); ok {
					for k, v := range EDRLJSYX {
						if v == sval {
							sk := strconv.Itoa(k)
							results = json.Number(sk)
							d.StartingAddress++
							log.Debugln("set 室外机额定容量节省指令 = ", results)
						}
					}
				}
			} else {
				return results, errors.New("设置参数错误")
			}
		}
	default:
		{
			return json.Number("0"), errors.New("错误的_varname")
		}
	}

	return results, nil
}

/***************************************读写接口实现**************************************************/

// RWDevValue ..
func (d *FUJITSU) RWDevValue(rw string, m dict) (ret dict, err error) {
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
	ret["mtype"] = d.mtype
	if rw == "r" {
		if d.mtype == "inM" {
			ret["subAddr"] = d.subAddr
			d.Quantity = 25
			d.FunctionCode = 4
			var addr uint16
			if dno, err := strconv.Atoi(d.subAddr); err == nil {
				addr = uint16(dno)
				if 1 > addr || addr > 128 {
					return nil, errors.New("室内机地址参数错误")
				}
				d.StartingAddress = 60*(addr-1) + 50
				log.Debugln("startAddress=", d.StartingAddress)
				bmdict, berr := d.ModbusRtu.RWDevValue("r", nil)
				if berr != nil {
					bmdict, berr = d.ModbusRtu.RWDevValue("r", nil)
				}
				if berr == nil {
					btdl := bmdict["Modbus-value"]
					bdl, _ := btdl.([]int)
					log.Debugf("室内机-%d receive data = %d", addr, bdl)
					d.insideStatus(ret, bdl)
				} else {
					ret["error"] = berr.Error()
					log.Debugln(ret)
					return ret, nil
				}
			}
		} else if d.mtype == "outM" {
			d.Quantity = 4
			d.FunctionCode = 4
			ret["subAddr"] = d.subAddr
			var addr uint16
			if dno, err := strconv.Atoi(d.subAddr); err == nil {
				addr = uint16(dno)
				if 1 > addr || addr > 100 {
					return nil, errors.New("室外机地址参数错误")
				}
				d.StartingAddress = 15*(addr-1) + 7740
				log.Debugln("startAddress=", d.StartingAddress)
				bmdict, berr := d.ModbusRtu.RWDevValue("r", nil)
				if berr == nil {
					btdl := bmdict["Modbus-value"]
					bdl, _ := btdl.([]int)
					log.Debugf("室外机-%d receive data = %d", addr, bdl)
					d.outsideStatus(ret, bdl)

				} else {
					ret["error"] = berr.Error()
					log.Debugln(ret)
					return ret, nil
				}
			}
		} else if d.mtype == "allInM" {
			d.Quantity = 2
			d.FunctionCode = 4
			var addr uint16
			d.StartingAddress = 7730
			log.Debugln("startAddress=", d.StartingAddress)
			bmdict, berr := d.ModbusRtu.RWDevValue("r", nil)
			if berr == nil {
				log.Debugln(bmdict)
				btdl := bmdict["Modbus-value"]
				bdl, _ := btdl.([]int)
				log.Debugf("ALL室内机-%d receive data = %d", addr, bdl)
				d.allInStatus(ret, bdl)
			} else {
				ret["error"] = berr.Error()
				log.Debugln(ret)
				return ret, nil
			}
		} else if d.mtype == "VRF" {
			ret["note"] = "it is a virtual device , write only"
		}
		/****************************************write device**********************************************/
	} else {
		if d.mtype == "inM" {
			d.FunctionCode = 6
			var addr uint16
			if dno, err := strconv.Atoi(d.subAddr); err == nil {
				addr = uint16(dno)
				if 1 > addr || addr > 128 {
					return nil, errors.New("室内机地址参数错误")
				}
				//if dno, ok := m["_varvalue"]; ok {
				//addr = getnm(dno)
				d.StartingAddress = 60*(addr-1) + 1
				wval, werr := d.encode(m)
				if werr != nil {
					ret["error"] = werr.Error()
					return ret, nil
				}
				if wval == "0" {
					ret["error"] = "_varvalue 有误,未执行写操作"
					return ret, nil
				}
				log.Debugln("wval", wval)
				log.Debugln("startAddress=", d.StartingAddress)
				bmdict, berr := d.ModbusRtu.RWDevValue("w", dict{"value": wval})
				if berr == nil {
					log.Infof("设置-%s-%d receive data = %v", m["_varname"], addr, bmdict)
					ret["cmdStatus"] = "successful"
				} else {
					ret["error"] = berr.Error()
					log.Debugln(ret)
					return ret, nil
				}
			} else {
				return nil, errors.New("地址参数错误")
			}
		} else if d.mtype == "allInM" || d.mtype == "VRF" {
			d.FunctionCode = 6
			d.StartingAddress = 7680
			wval, werr := d.encode(m)
			if werr != nil {
				ret["error"] = werr.Error()
				log.Debugf("设置%s-(%s)-%v", d.mtype, m["_varname"], werr)
				return ret, nil
			}
			if wval == "0" {
				ret["error"] = "_varvalue 有误,未执行写操作"
				return ret, nil
			}
			log.Debugln("wval", wval)
			log.Debugln("startAddress=", d.StartingAddress)
			bmdict, berr := d.ModbusRtu.RWDevValue("w", dict{"value": wval})
			if berr == nil {
				log.Infof("设置%s-%s receive data = %v", d.mtype, m["_varname"], bmdict)
				ret["cmdStatus"] = "successful"
			} else {
				ret["error"] = berr.Error()
				log.Debugln(ret)
				return ret, nil
			}
		} else if d.mtype == "outM" {
			d.FunctionCode = 6
			var addr uint16
			if dno, err := strconv.Atoi(d.subAddr); err == nil {
				addr = uint16(dno)
				if 1 > addr || addr > 100 {
					return nil, errors.New("室外机地址参数错误")
				}
				d.StartingAddress = 15*(addr-1) + 7711
				wval, werr := d.encode(m)
				if werr != nil {
					ret["error"] = werr.Error()
					return ret, nil
				}
				if wval == "0" {
					ret["error"] = "_varvalue 有误,未执行写操作"
					return ret, nil
				}
				log.Debugln("wval", wval)
				log.Debugln("startAddress=", d.StartingAddress)
				bmdict, berr := d.ModbusRtu.RWDevValue("w", dict{"value": wval})
				if berr == nil {
					log.Infof("设置-%s-%d receive data = %v", m["_varname"], addr, bmdict)
					ret["cmdStatus"] = "successful"
				} else {
					ret["error"] = berr.Error()
					log.Debugln(ret)
					return ret, nil
				}
			} else {
				return nil, errors.New("地址参数错误")
			}
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
	var ruint uint16
	ruint = 0
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
