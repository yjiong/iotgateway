package device

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	//	"sync"
	log "github.com/Sirupsen/logrus"
	simplejson "github.com/bitly/go-simplejson"
)

//var mutex sync.Mutex
type QDSL_SM510 struct {
	//继承于ModebusRtu
	ModbusRtu /**************按不同设备自定义*************************/

	/**************按不同设备自定义*************************/
}

func init() {
	RegDevice["QDSL_SM510"] = &QDSL_SM510{}
}

func (d *QDSL_SM510) NewDev(id string, ele map[string]string) (DeviceRWer, error) {
	ndev := new(QDSL_SM510)
	ndev.Device = d.Device.NewDev(id, ele)
	/***********************初始化设备的特有的参数*****************************/
	ndev.BaudRate = 19200 //, _ = strconv.Atoi(ele["BaudRate"])
	ndev.DataBits = 8     //, _ = strconv.Atoi(ele["DataBits "])
	ndev.StopBits = 1     //, _ = strconv.Atoi(ele["StopBits"])
	ndev.Parity = "N"     //, _ = ele["Parity"]
	//ndev.Function_code = 3
	//	saint, _ := strconv.Atoi(ele["Starting_address"])
	//ndev.Starting_address = 0
	//	qint, _ := strconv.Atoi(ele["Quantity"])
	//ndev.Quantity = 12
	/***********************初始化设备的特有的参数*****************************/
	return ndev, nil
}

func (d *QDSL_SM510) GetElement() (dict, error) {
	conn := dict{
		/***********************设备的特有的参数*****************************/
		"devaddr": d.devaddr,
		"commif":  d.commif,
		//"BaudRate": 19200,
		//"DataBits": 8,
		//"StopBits": 1,
		//"Parity":   "N",
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
func (d *QDSL_SM510) HelpDoc() interface{} {
	conn := dict{
		"devaddr": "设备地址",
		/***********QDSL_SM510设备的参数*****************************/
		"备注":       "由于寄存器不连续,请一条命令设置一个字段",
		"定时交换分钟数":  "单位分钟,大于0小于9999,int类型",
		"无水停机压力":   "单位MP,大于0小于9999,int类型",
		"无水停机延时":   "单位秒,大于0小于9999,int类型",
		"有水开机压力":   "单位MP,大于0小于9999,int类型",
		"有水开机延时":   "单位秒,大于0小于9999,int类型",
		"设备通讯地址":   "大于0小于9999,int类型",
		"设定压力":     "单位MP,初始化值0.300MP,大于0,小于10,小数位3位,float类型",
		"软件超压保护偏差": "单位MP,初始化值0.300MP,大于0,小于10,小数位3位,float类型",
		"远程启动停止":   "=1停止,=0启动",
		//"BaudRate": "波特率,比如(9600)",
		//"DataBits": "数据位,比如(8)",
		//"Parity":   "校验,(N,E,O)",
		//"StopBits": "停止位,比如(1)",
		/***********QDSL_SM510设备的参数*****************************/
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
		"_type":  "QDSL_SM510", //设备类型
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
func (d *QDSL_SM510) CheckKey(ele dict) (bool, error) {

	//if _, ok := ele["BaudRate"].(json.Number); !ok {
	//return false, errors.New(fmt.Sprintf("QDSL_SM510 device must have int type element 波特率 :BaudRate"))
	//}
	//if _, ok := ele["DataBits"].(json.Number); !ok {
	//return false, errors.New(fmt.Sprintf("QDSL_SM510 device must have int type element 数据位 :DataBits"))
	//}
	//if _, ok := ele["StopBits"].(json.Number); !ok {
	//return false, errors.New(fmt.Sprintf("QDSL_SM510 device must have int type element 停止位 :StopBits"))
	//}
	//if _, ok := ele["Parity"].(string); !ok {
	//return false, errors.New(fmt.Sprintf("QDSL_SM510 device must have string type element 校验 :Parity"))
	//}
	return true, nil
}

/***************************************添加设备参数检验**********************************************/
func (d *QDSL_SM510) getfloat(m dict) (dict, error) {
	var wval json.Number
	if k, ok := m["_varvalue"]; ok {
		log.Debugln("varvalue=", k)
		if jn, ok := k.(json.Number); ok {
			k = string(jn)
		}
		if str, ok := k.(string); ok {
			fstr, _ := strconv.ParseFloat(str, 64)
			istr := int(fstr * 1000)
			log.Debugln(wval)
			if istr < 0 || istr > 9999 {
				return nil, errors.New("wrong varvalue")
			}
			istr = (istr/999)*0x1000 + (istr%1000/100)*0x100 + (istr%100/10)*0x10 + (istr % 10)
			wval = json.Number(fmt.Sprintf("%d", istr))
		} else {
			return nil, errors.New("wrong varvalue")
		}
		log.Debugln("send modbus data = ", wval)
	}
	return dict{"value": wval}, nil
}

func (d *QDSL_SM510) getint(m dict) (dict, error) {
	var wval json.Number
	if k, ok := m["_varvalue"]; ok {
		log.Debugln("varvalue=", k)
		if jn, ok := k.(json.Number); ok {
			k = string(jn)
		}
		if str, ok := k.(string); ok {
			istr, _ := strconv.Atoi(str)
			log.Debugln(wval)
			if istr < 0 || istr > 9999 {
				return nil, errors.New("wrong varvalue")
			}
			istr = (istr/999)*0x1000 + (istr%1000/100)*0x100 + (istr%100/10)*0x10 + (istr % 10)
			wval = json.Number(fmt.Sprintf("%d", istr))
		} else {
			return nil, errors.New("wrong varvalue")
		}
		log.Debugln("send modbus data = ", wval)
	}
	return dict{"value": wval}, nil

}

/***************************************读写接口实现**************************************************/
func (d *QDSL_SM510) RWDevValue(rw string, m dict) (ret dict, err error) {
	ret = make(dict)
	var mdict dict
	defer func() {
		if drive_err := recover(); drive_err != nil {
			log.Errorf("drive programer  error : (%s)", drive_err)
			errstr := fmt.Sprintf("drive programer  error : (%s)", drive_err)
			err = errors.New(errstr)
		}
	}()
	d.mutex.Lock()
	defer d.mutex.Unlock()
	//log.SetLevel(log.DebugLevel)
	ret["_devid"] = d.devid
	mdata := make([]int, 32)
	if rw == "r" {
		d.Starting_address = 2
		d.Quantity = 22
		d.Function_code = 3
		mdict, err = d.ModbusRtu.RWDevValue("r", nil)
		if err == nil {
			tdl := mdict["Modbus-value"]
			dl, ok := tdl.([]int)
			//log.Info(dl)
			if ok {
				ret["设定压力"] = Bcd2_2f(dl[0], dl[1]) / 1000.0
				ret["有水开机压力"] = Bcd2_2f(dl[2], dl[3]) / 1000.0
				ret["无水停机压力"] = Bcd2_2f(dl[4], dl[5]) / 1000.0
				ret["有水开机延时"] = Bcd2_2f(dl[6], dl[7])
				ret["无水停机延时"] = Bcd2_2f(dl[8], dl[9])
				ret["设备通讯地址"] = Bcd2_2f(dl[18], dl[19])
				ret["定时交换分钟数"] = Bcd2_2f(dl[22], dl[23])
				ret["软件超压保护偏差"] = Bcd2_2f(dl[24], dl[25]) / 1000.0
			} else {
				ret["error"] = "部分参数读取失败"
			}
		} else {
			ret["_status"] = "offline"
			return ret, nil
		}
		d.Starting_address = 339
		d.Quantity = 32
		d.Function_code = 3
		mdict, err = d.ModbusRtu.RWDevValue("r", nil)
		if err == nil {
			tdl := mdict["Modbus-value"]
			dl, ok := tdl.([]int)
			for i := 0; i < 32; i += 2 {
				mdata[i/2] = dl[i]*0x100 + dl[i+1]
			}
			//log.Info(dl)
			if ok {
				ret["过程和报警序号"] = mdata[0]
				ret["过程和报警序号"] = mdata[0]
				ret["2#泵手动/自动"] = (mdata[1] & 0x0001) / 0x0001
				ret["2#泵变频指示"] = (mdata[1] & 0x0002) / 0x0002
				ret["2#泵工频指示"] = (mdata[1] & 0x0004) / 0x0004
				ret["2#泵故障指示"] = (mdata[1] & 0x0008) / 0x0008
				ret["1#泵手动/自动"] = (mdata[1] & 0x0100) / 0x0100
				ret["1#泵变频指示"] = (mdata[1] & 0x0200) / 0x0200
				ret["1#泵工频指示"] = (mdata[1] & 0x0400) / 0x0400
				ret["1#泵故障指示"] = (mdata[1] & 0x0800) / 0x0800

				ret["4#泵手动/自动"] = (mdata[2] & 0x0001) / 0x0001
				ret["4#泵变频指示"] = (mdata[2] & 0x0002) / 0x0002
				ret["4#泵工频指示"] = (mdata[2] & 0x0004) / 0x0004
				ret["4#泵故障指示"] = (mdata[2] & 0x0008) / 0x0008
				ret["3#泵手动/自动"] = (mdata[2] & 0x0100) / 0x0100
				ret["3#泵变频指示"] = (mdata[2] & 0x0200) / 0x0200
				ret["3#泵工频指示"] = (mdata[2] & 0x0400) / 0x0400
				ret["3#泵故障指示"] = (mdata[2] & 0x0800) / 0x0800

				ret["手动/自动"] = (mdata[4] & 0x0001) / 0x0001
				ret["启动/停止"] = (mdata[4] & 0x0002) / 0x0002
				ret["变频器报警"] = (mdata[4] & 0x0004) / 0x0004
				ret["电源故障"] = (mdata[4] & 0x0008) / 0x0008
				ret["无水停机"] = (mdata[4] & 0x0010) / 0x0010
				ret["水箱水位超高"] = (mdata[4] & 0x0020) / 0x0020
				ret["硬件超压保护"] = (mdata[4] & 0x0040) / 0x0040
				ret["补偿器报警"] = (mdata[4] & 0x0080) / 0x0080
				//ret["预留"] = (mdata[4] & 0x0100)/0x0100
				ret["最大供水能力停机"] = (mdata[4] & 0x0200) / 0x0200
				ret["软件超压保护"] = (mdata[4] & 0x0400) / 0x0400
				ret["水池液位超低"] = (mdata[4] & 0x0800) / 0x0800
				ret["小流量停机标志"] = (mdata[4] & 0x1000) / 0x1000
				ret["巡检状态"] = (mdata[4] & 0x2000) / 0x2000
				ret["进水压大于设定压停机"] = (mdata[4] & 0x4000) / 0x4000
				ret["所有泵故障"] = (mdata[4] & 0x8000) / 0x8000

				ret["进水压大于出水压停机"] = (mdata[5] & 0x0001) / 0x0001
				ret["小流量破坏"] = (mdata[5] & 0x0002) / 0x0002
				ret["变频器复位失败"] = (mdata[5] & 0x0004) / 0x0004
				ret["软启故障标志"] = (mdata[5] & 0x0008) / 0x0008
				ret["定时分时段启动停止标志"] = (mdata[5] & 0x0010) / 0x0010
				ret["紧急停止标志"] = (mdata[5] & 0x0020) / 0x0020
				//ret["预留"] = (mdata[5] & 0x0040)/0x0040
				ret["定时密码停机标志"] = (mdata[5] & 0x0080) / 0x0080
				ret["远程遥控停机标志"] = (mdata[5] & 0x0100) / 0x0100
				ret["后备电池失效"] = (mdata[5] & 0x0200) / 0x0200
				ret["正在处理小流量标志"] = (mdata[5] & 0x0400) / 0x0400
				ret["消防状态标志"] = (mdata[5] & 0x0800) / 0x0800
				ret["小流量间隔标志"] = (mdata[5] & 0x1000) / 0x1000

				ret["当前设定压力/MP"] = Bcd2_2f(dl[16], dl[17]) / 1000.0
				ret["变频器频率/HZ"] = Bcd2_2f(dl[18], dl[19]) / 10.0
				ret["进水口压力/MP"] = Bcd2_2f(dl[20], dl[21]) / 1000.0
				ret["出水口压力/MP"] = Bcd2_2f(dl[22], dl[23]) / 1000.0
				ret["1#泵电流/A"] = Bcd2_2f(dl[26], dl[27]) / 10.0
				ret["2#泵电流/A"] = Bcd2_2f(dl[28], dl[29]) / 10.0
				ret["3#泵电流/A"] = Bcd2_2f(dl[30], dl[31]) / 10.0
				ret["4#泵电流/A"] = Bcd2_2f(dl[32], dl[33]) / 10.0
				ret["5#泵电流/A"] = Bcd2_2f(dl[34], dl[35]) / 10.0
				ret["液位深度"] = Bcd2_2f(dl[36], dl[37])
				ret["系统电压/V"] = Bcd2_2f(dl[38], dl[39]) / 10.0
				ret["累计流量"] = Bcd2_2f(dl[42], dl[43])*10000 + Bcd2_2f(dl[40], dl[41])
				ret["浊度"] = Bcd2_2f(dl[44], dl[45])
				ret["余氯"] = Bcd2_2f(dl[46], dl[47])
				ret["PH值"] = Bcd2_2f(dl[48], dl[49])
				y := (dl[58]*0x100 + dl[59]) / 100
				m := (dl[58]*0x100 + dl[59]) % 100
				d := (dl[60]*0x100 + dl[61]) / 100
				h := (dl[60]*0x100 + dl[61]) % 100
				M := (dl[62]*0x100 + dl[63]) / 100
				s := (dl[62]*0x100 + dl[63]) % 100
				ret["时间"] = fmt.Sprintf("%x-%x-%x %x:%x:%x", y, m, d, h, M, s)

			}
		} else {
			ret["error"] = err.Error()
			log.Debugln(ret)
			return ret, nil
		}
	} else {
		d.Function_code = 6
		var method func(dict) (dict, error)
		if k, ok := m["_varname"]; ok {
			switch k {
			case "设定压力":
				d.Starting_address = 2
				method = d.getfloat
			case "有水开机压力":
				d.Starting_address = 3
				method = d.getfloat
			case "无水停机压力":
				d.Starting_address = 4
				method = d.getfloat
			case "有水开机延时":
				d.Starting_address = 5
				method = d.getint
			case "无水停机延时":
				d.Starting_address = 6
				method = d.getint
			case "设备通讯地址":
				d.Starting_address = 11
				method = d.getint
			case "定时交换分钟数":
				d.Starting_address = 13
				method = d.getint
			case "软件超压保护偏差":
				d.Starting_address = 24
				method = d.getfloat
			case "远程启动停止":
				d.Starting_address = 313
				method = d.getint
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
	log.Debugln(string(pinforet))
	return ret, err
}

/***************************************读写接口实现**************************************************/
