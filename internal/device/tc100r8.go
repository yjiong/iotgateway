package device

import (
	"encoding/json"
	"errors"
	"fmt"

	log "github.com/Sirupsen/logrus"
	//simplejson "github.com/bitly/go-simplejson"
	"strconv"
	"strings"
	//"sync"
)

//var mutex sync.Mutex
type TC100R8 struct {
	//继承于ModebusRtu
	ModbusRtu
	/**************按不同设备自定义*************************/

	/**************按不同设备自定义*************************/
}

func init() {
	RegDevice["TC100R8"] = &TC100R8{}
}

func (d *TC100R8) NewDev(id string, ele map[string]string) (Devicerwer, error) {
	ndev := new(TC100R8)
	ndev.Device = d.Device.NewDev(id, ele)
	/***********************初始化设备的特有的参数*****************************/
	ndev.BaudRate = 2400
	ndev.DataBits = 8
	ndev.StopBits = 1
	ndev.Parity = "N"
	//	saint, _ := strconv.Atoi(ele["StartingAddress"])
	ndev.StartingAddress = 0
	//	qint, _ := strconv.Atoi(ele["Quantity"])
	ndev.Quantity = 12
	/***********************初始化设备的特有的参数*****************************/
	return ndev, nil
}

func (d *TC100R8) GetElement() (dict, error) {
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
func (d *TC100R8) HelpDoc() interface{} {
	conn := dict{
		"devaddr": "设备地址",
		/***********TC100R8设备的参数*****************************/
		"commif": "通信接口,比如(rs485-1)",
		/***********TC100R8设备的参数*****************************/
	}
	rParameter := dict{
		"_devid": "被读取设备对象的id",
		/***********读取设备的参数*****************************/
		"_varname.1": "设备编号",
		"_varname.2": "策略",
		"解释":         "如果无_varname字段,将读取默认参数",
		/***********读取设备的参数*****************************/
	}
	va1_d1 := dict{
		"回路":   "1",
		"初始状态": "on",
		"时间点":  "0:10,4:20,5:22,7:33,8:44,15:55,16:55,19:55,20:33,21:36,22:37,23:38",
	}
	va1_d2 := dict{
		"回路":   "值为1-8,string类型",
		"初始状态": "值为(on|off),string",
		"时间点":  "时间点,小于12个",
		"解释":   "必须8路都写",
	}
	va1_1 := []dict{va1_d2, va1_d1}
	va2_d1 := dict{
		"回路": "1",
		"状态": "自动",
		"开关": "off",
	}
	va2_d2 := dict{
		"回路": "值为1-8,string类型",
		"状态": "值为(自动|手动|不操作),string类型",
		"开关": "值为(on|off),sting类型",
		"解释": "必须8路都写",
	}
	va2_2 := []dict{va2_d2, va2_d1}
	wParameter := dict{
		"_devid": "被操作设备对象的id",
		/***********操作设备的参数*****************************/
		"_varname.1":  "自动策略",
		"_varvalue.1": va1_1,
		"_varname.2":  "手动策略",
		"_varvalue.2": va2_2,
		"_varname.3":  "设置时间",
		"_varvalue.3": `设置时间格式举例:"17-11-5 18:30:28"`,
		"_varname.4":  "单回路策略",
		"_varvalue.4": "格式同自动策略,单条发送",
		/***********操作设备的参数*****************************/
	}
	data := dict{
		"_devid": "添加设备对象的id",
		"_type":  "TC100R8", //设备类型
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

/***************************************添加设备参数检验**********************************************/
func (d *TC100R8) CheckKey(ele dict) (bool, error) {

	return true, nil
}

/***************************************添加设备参数检验**********************************************/
func (d *TC100R8) autostrate(m dict) (dict, error) {
	var wbyte []string = []string{"0"}
	d.Quantity = 13
	var vallist []interface{}
	//vallist = append(vallist, json.Number(strconv.Itoa(0)))
	startstat := map[string]string{"on": "1", "off": "0"}
	//key := fmt.Sprintf("回路%d自动策略", i)
	if k, ok := m["_varvalue"]; ok {
		if km, ok := k.(map[string]interface{}); ok {
			log.Debugln("varvalue=", km)
			if ks1, ok := km["loop"]; ok {
				log.Debugln("回路=", ks1)
				if nm, ok := ks1.(json.Number); ok {
					nmi64, _ := nm.Int64()
					if 1 > nmi64 || nmi64 > 8 {
						return nil, errors.New("自动策略回路参数错误")
					}
					d.StartingAddress = 13 * (uint16(nmi64) - 1)
					log.Debugln("startAddress=", d.StartingAddress)
				} else {
					if nm, ok := ks1.(string); ok {
						snm, _ := strconv.Atoi(nm)
						if 1 > snm || snm > 8 {
							return nil, errors.New("自动策略回路参数错误")
						}
						d.StartingAddress = 13 * (uint16(snm) - 1)
						log.Debugln("startAddress=", d.StartingAddress)
					}
				}
			}
			if ks2, ok := km["startSwitch"]; ok {
				log.Debugln("初始状态=", ks2)
				if str, ok := ks2.(string); ok {
					if onoff, ok := startstat[str]; ok {
						wbyte = append(wbyte, onoff)
						log.Debugln("send modbusbyte=", wbyte)
					}
				}
			}
			if ks3, ok := km["timeSet"]; ok {
				if str, ok := ks3.(string); ok {
					log.Debugln("时间点=", str)
					wbyte = append(wbyte, strings.Split(strings.Replace(str, ":", ",", -1), ",")...)
					if len(wbyte) < 26 {
						for {
							wbyte = append(wbyte, string("0"))
							if len(wbyte) >= 26 {
								break
							}
						}
					}
				}
			}
		}
	}
	log.Debugln("send modbus data = ", wbyte)
	if len(wbyte) != 26 {
		return nil, errors.New("自动策略参数错误")
	}
	for _, v := range wbyte {
		jn := json.Number(v)
		vallist = append(vallist, jn)
	}
	wval := dict{"value": vallist}
	return wval, nil
}

func (d *TC100R8) settime(m dict) (dict, error) {
	var wbyte []string
	d.Quantity = 3
	d.StartingAddress = 123
	var vallist []interface{}
	if k, ok := m["_varvalue"]; ok {
		log.Debugln("varvalue=", k)
		if str, ok := k.(string); ok {
			log.Debugln(str)
			tstr := str
			for _, symbol := range []string{"-", " ", ":", "/"} {
				tstr = strings.Replace(tstr, symbol, ",", -1)
			}
			wbyte = append(wbyte, strings.Split(tstr, ",")...)
		}
	}
	log.Debugln("send modbus data = ", wbyte)
	if len(wbyte) != 6 {
		return nil, errors.New("设置时间参数错误")
	}
	for _, v := range wbyte {
		jn := json.Number(v)
		vallist = append(vallist, jn)
	}
	return dict{"value": vallist}, nil
}

func (d *TC100R8) handlerstrte(m dict) (dict, error) {
	wbyte := make([]string, 32)
	d.Quantity = 16
	d.StartingAddress = 104
	status := map[string]string{"自动": "1", "手动": "2", "不操作": "0"}
	onOff := map[string]string{"on": "1", "off": "0"}
	var vallist []interface{}
	var test int = 0
	if kif, ok := m["_varvalue"]; ok {
		if klist, ok := kif.([]interface{}); ok {
			for _, k := range klist {
				if kmap, ok := k.(map[string]interface{}); ok {
					line, ok1 := kmap["回路"].(string)
					stat, ok2 := kmap["状态"].(string)
					delay, ok3 := kmap["分钟"].(string)
					onoff, ok4 := kmap["开关"].(string)
					if ok1 && ok2 && ok3 && ok4 {
						lineint, _ := strconv.Atoi(line)
						log.Debugln("line=", lineint)
						log.Debugln("status=", status[stat])
						log.Debugln("onOff", onOff[onoff])
						tdelay, _ := strconv.Atoi(delay)
						log.Debugln("delay=", tdelay)
						lineint -= 1
						wbyte[lineint*4] = status[stat]
						wbyte[lineint*4+1] = onOff[onoff]
						wbyte[lineint*4+2] = strconv.Itoa(tdelay / 0xff)
						wbyte[lineint*4+3] = strconv.Itoa(tdelay & 0xff)
						test += 1
					}
				}
			}
		}
	}
	log.Debugln("send modbus data = ", wbyte, test)
	if test != 8 {
		return nil, errors.New("设置手动策略参数错误")
	}
	for _, v := range wbyte {
		jn := json.Number(v)
		vallist = append(vallist, jn)
	}
	return dict{"value": vallist}, nil
}

func (d *TC100R8) allstrate(m dict) (dict, error) {
	var vallist []interface{}
	temp := make([]interface{}, 8)
	if kif, ok := m["_varvalue"]; ok {
		if klist, ok := kif.([]interface{}); ok {
			for _, kl := range klist {
				if km, ok := kl.(map[string]interface{}); ok {
					if ks1, ok := km["loop"]; ok {
						if nm, ok := ks1.(string); ok {
							snm, _ := strconv.Atoi(nm)
							vl, err := d.autostrate(dict{"_varvalue": km})
							if err == nil {
								temp[snm-1] = vl["value"]
							} else {
								return nil, errors.New("自动策略参数错误")
							}
							log.Debugln("subvalue=", temp[snm-1])
						}
					}
				}
			}
		}
	}
	for i := 0; i < 8; i++ {
		for j := 0; j < 26; j++ {
			tmif, _ := temp[i].([]interface{})
			vallist = append(vallist, tmif[j])
		}
	}
	log.Debugln("所有策略=", vallist)
	d.Quantity = 104
	d.StartingAddress = 0
	return dict{"value": vallist}, nil
}

// --------------------------- 华丽的分割线 --------------------------------------

// 读 厂家编号
func (d *TC100R8) doReadFactoryNo(m dict) (ret dict, err error) {

	d.StartingAddress = 126
	d.Quantity = 4
	d.FunctionCode = 3

	rstDict, rstErr := d.ModbusRtu.RWDevValue("r", nil)
	if rstErr != nil {
		return nil, errors.New("read error")
	} else {
		btdl := rstDict["Modbus-value"]
		bdl, _ := btdl.([]int)

		ret = make(dict)
		ret["factoryNo"] = fmt.Sprintf("%d%d%d%d%d%d%d%d", bdl[0], bdl[1], bdl[2], bdl[3], bdl[4], bdl[5], bdl[6], bdl[7])

		return ret, nil
	}
}

// 读 策略
func (d *TC100R8) doReadStrategy(m dict) (ret dict, err error) {

	d.StartingAddress = 0
	d.Quantity = 123
	d.FunctionCode = 3

	rstDict, rstErr := d.ModbusRtu.RWDevValue("r", nil)
	if rstErr != nil {
		return nil, errors.New("read error")
	} else {
		tdl := rstDict["Modbus-value"]
		dl, _ := tdl.([]int)
		log.Debugf("receive data = %d", dl)

		status := map[int]string{1: "on", 0: "off"}

		ret = make(dict)
		ret["localTime"] = fmt.Sprintf("20%d-%d-%d %d:%d:%d", dl[240], dl[241], dl[242], dl[243], dl[244], dl[245])

		for i := 0; i < 8; i++ {
			timeList := ""
			for j := 0; j < 12; j++ {
				timeList += fmt.Sprintf("%d:%d", dl[i*26+j*2+2], dl[i*26+j*2+3])
			}
			if timeList != "" {
				timeList = strings.TrimLeft(timeList, ",")
			}
			ret[fmt.Sprintf("begin%d", i+1)] = status[dl[i*26+1]]
			ret[fmt.Sprintf("timeList%d", i+1)] = timeList
		}

		return ret, nil
	}
}

// 读 主动上报
func (d *TC100R8) doReadRealData(m dict) (ret dict, err error) {

	d.StartingAddress = 120
	d.Quantity = 42
	d.FunctionCode = 3

	rstDict, rstErr := d.ModbusRtu.RWDevValue("r", nil)
	if rstErr != nil {
		return nil, errors.New("read error")
	} else {
		tdl := rstDict["Modbus-value"]
		dl, _ := tdl.([]int)
		log.Debugf("receive data = %d", dl)

		ret = make(dict)

		ret["localTime"] = fmt.Sprintf("20%d-%d-%d %d:%d:%d", dl[0], dl[1], dl[2], dl[3], dl[4], dl[5])

		status := map[int]string{1: "on", 0: "off"}
		handcl := map[int]string{0: "C", 1: "A", 2: "M", 0xff: "?"}

		var offSet int
		offSet = 20

		for i := 0; i < 8; i++ {
			ret[fmt.Sprintf("switch%d", i+1)] = status[dl[offSet+i*8+1]]
			ret[fmt.Sprintf("ctrl%d", i+1)] = handcl[dl[offSet+i*8]]
			if ret[fmt.Sprintf("ctrl%d", i+1)] == "M" {
				ret[fmt.Sprintf("endTime%d", i+1)] = fmt.Sprintf("20%d-%d-%d %d:%d:%d", dl[offSet+i*8+2], dl[offSet+i*8+3], dl[offSet+i*8+4], dl[offSet+i*8+5], dl[offSet+i*8+6], dl[offSet+i*8+7])
			} else {
				ret[fmt.Sprintf("endTime%d", i+1)] = ""
			}

		}

		return ret, nil
	}
}

// 读参数
func (d *TC100R8) doGetVar(m dict) (ret dict, err error) {
	varName, _ := m["_varname"]
	switch varName {
	// 厂家编号
	case "factoryNo":
		{
			return d.doReadFactoryNo(m)
		}
	// 策略
	case "strategy":
		{
			return d.doReadStrategy(m)
		}
	// 主动上报
	default:
		{
			return d.doReadRealData(m)
		}

	}

	return nil, errors.New("参数未")
}

// 写 校时
func (d *TC100R8) doTiming(m dict) (ret dict, err error) {

	d.StartingAddress = 123
	d.Quantity = 3
	d.FunctionCode = 16

	var wbyte []string
	var vallist []interface{}
	var timeValue string

	timeValue = m["_varvalue"].(string)
	for _, symbol := range []string{"-", " ", ":", "/"} {
		timeValue = strings.Replace(timeValue, symbol, ",", -1)
	}
	wbyte = append(wbyte, strings.Split(timeValue, ",")...)

	if len(wbyte) != 6 {
		return nil, errors.New("设置时间参数错误")
	}
	for _, v := range wbyte {
		jn := json.Number(v)
		vallist = append(vallist, jn)
	}

	ret, err = d.ModbusRtu.RWDevValue("w", dict{"value": vallist})

	return ret, err
}

// 写 手动
func (d *TC100R8) doManual(m dict) (ret dict, err error) {

	d.StartingAddress = 104
	d.Quantity = 16
	d.FunctionCode = 16

	status := map[string]string{"A": "1", "M": "2"}
	onOff := map[string]string{"on": "1", "off": "0"}

	wbyte := make([]string, 32)
	var vallist []interface{}

	// 初始化 8个回路都不操作
	for i := 0; i < 8; i++ {
		wbyte[i*4] = "255"
		wbyte[i*4+1] = "0"
		wbyte[i*4+2] = "0"
		wbyte[i*4+3] = "0"
	}

	manualValue := m["_varvalue"]
	if klist, ok := manualValue.([]interface{}); ok {
		for _, k := range klist {
			if kmap, ok := k.(map[string]interface{}); ok {
				line, ok1 := kmap["loop"].(string)
				stat, ok2 := kmap["ctrl"].(string)
				delay, ok3 := kmap["delay"].(string)
				onoff, ok4 := kmap["switch"].(string)
				if ok1 && ok2 && ok3 && ok4 {
					//log.Debugln("xxxxxxxxxxxxxxxxxx", manualValue)
					lineint, _ := strconv.Atoi(line)
					log.Debugln("line=", lineint)
					log.Debugln("status=", status[stat])
					log.Debugln("onOff", onOff[onoff])
					tdelay, _ := strconv.Atoi(delay)
					log.Debugln("delay=", tdelay)
					lineint -= 1
					wbyte[lineint*4] = status[stat]
					wbyte[lineint*4+1] = onOff[onoff]
					wbyte[lineint*4+2] = strconv.Itoa(tdelay / 0xff)
					wbyte[lineint*4+3] = strconv.Itoa(tdelay & 0xff)
				}
			}
		}
	}

	for _, v := range wbyte {
		jn := json.Number(v)
		vallist = append(vallist, jn)
	}

	ret, err = d.ModbusRtu.RWDevValue("w", dict{"value": vallist})
	return ret, err
}

// 写 一个策略
func (d *TC100R8) doIssueOneStrategy(m dict) (ret dict, err error) {

	d.FunctionCode = 16

	var wval dict
	wval, err = d.autostrate(m)
	if err == nil {
		ret, err = d.ModbusRtu.RWDevValue("w", wval)
	}

	return ret, err
}

// 写 所有策略
func (d *TC100R8) doIssueAllStrategy(m dict) (ret dict, err error) {

	d.FunctionCode = 16

	var wval dict
	wval, err = d.allstrate(m)
	if err == nil {
		d.StartingAddress = 0
		d.Quantity = 104
		ret, err = d.ModbusRtu.RWDevValue("w", wval)
	}

	return ret, err
}

// 写参数
func (d *TC100R8) doSetVar(m dict) (ret dict, err error) {
	if varName, ok := m["_varname"]; ok {

		if varValue, ok := m["_varvalue"]; ok {

			switch varName {
			// 单回路策略
			case "oneStrategy":
				{
					ret, err = d.doIssueOneStrategy(m)
				}
			// 自动策略
			case "allStrategy":
				{
					ret, err = d.doIssueAllStrategy(m)
				}
			// 校时
			case "timing":
				{
					ret, err = d.doTiming(m)
				}
			// 手动控制
			case "manual":
				{
					ret, err = d.doManual(m)
				}
			default:
				{
					return nil, errors.New("错误的_varname")
				}
			}
		} else {
			return nil, errors.New(fmt.Sprintf("错误的_varvalue,%s", varValue))
		}

		return ret, err
	}

	return nil, errors.New("未指定_varname")
}

/***************************************读写接口实现**************************************************/
// 读写参数入口
func (d *TC100R8) RWDevValue(rw string, m dict) (ret dict, err error) {
	defer func() {
		if driveErr := recover(); driveErr != nil {
			log.Errorf("drive programer  error : %s", driveErr)
			errstr := fmt.Sprintf("drive programer  error : %s", driveErr)
			err = errors.New(errstr)
		}
	}()
	d.mutex.Lock()
	defer d.mutex.Unlock()
	//log.SetLevel(log.DebugLevel)

	if rw == "r" {
		ret, err = d.doGetVar(m)
	} else {
		ret, err = d.doSetVar(m)
	}

	if err != nil {
		ret = make(dict)
		ret["_status"] = "offline"
	}

	ret["_devid"] = d.devid

	return ret, err
}
