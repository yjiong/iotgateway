package device

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	//	"sync"
	log "github.com/Sirupsen/logrus"
	//simplejson "github.com/bitly/go-simplejson"
	"github.com/yjiong/iotgateway/serial"
)

//DDZY422N ..
type DDZY422N struct {
	//继承于Device
	Device
	/**************按不同设备自定义*************************/
	//BaudRate int
	//DataBits int
	//StopBits int
	//Parity   string

	/**************按不同设备自定义*************************/
}

func init() {
	RegDevice["DDZY422N"] = &DDZY422N{}
}

// NewDev ..
func (d *DDZY422N) NewDev(id string, ele map[string]string) (Devicerwer, error) {
	ndev := new(DDZY422N)
	ndev.Device = d.Device.NewDev(id, ele)
	/***********************初始化设备的特有的参数*****************************/
	//ndev.BaudRate, _ = strconv.Atoi(ele["BaudRate"])
	//ndev.DataBits, _ = strconv.Atoi(ele["DataBits "])
	//ndev.StopBits, _ = strconv.Atoi(ele["StopBits"])
	//ndev.Parity, _ = ele["Parity"]
	/***********************初始化设备的特有的参数*****************************/
	return ndev, nil
}

// GetElement ..
func (d *DDZY422N) GetElement() (dict, error) {
	conn := dict{
		/***********************设备的特有的参数*****************************/
		"devaddr": d.devaddr,
		"commif":  d.commif,
		//"BaudRate": d.BaudRate,
		//"DataBits": d.DataBits,
		//"StopBits": d.StopBits,
		//"Parity":   d.Parity,
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
func (d *DDZY422N) HelpDoc() interface{} {
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

	initval := map[string]interface{}{
		"报警金额":   "int类型0-9999",
		"报警负荷":   "float类型0.01-99.99",
		"允许透支金额": "int类型0-9999",
		"允许囤积金额": "float类型 0.01<=X<=999999.99",
		//"金额报警跳闸时间": json.Number("14"),
		"尖单价": "float类型0.01-99.99",
		"峰单价": "float类型0.01-99.99",
		"平单价": "float类型0.01-99.99",
		"谷单价": "float类型0.01-99.99",
		//"报警方式": 1,
	}
	wParameter := dict{
		"_devid": "被操作设备对象的id",
		/***********操作设备的参数*****************************/
		"强制合闸": "无需值参数",
		"强制断闸": "无需值参数",
		"撤销强制": "无需值参数",
		"初始化":  initval,
		"充值":   "0.01<=X<=9999.99",
		"退款":   "0.01<=X<=9999.99",
		"注意":   "一次执行一条命令",
	}
	data := dict{
		"_devid": "添加设备对象的id",
		"_type":  "DDZY422N", //设备类型
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

// CheckKey ..
func (d *DDZY422N) CheckKey(ele dict) (bool, error) {
	return true, nil
}

/***************************************添加设备参数检验**********************************************/
func (d *DDZY422N) plus33(data []byte) []byte {
	rd := make([]byte, len(data))
	for i := range data {
		rd[len(data)-i-1] = (data[i] + 0x33) % 0xff
	}
	return rd
}

func (d *DDZY422N) sub33(data []byte) []byte {
	rd := make([]byte, len(data))
	for i := range data {
		rd[len(data)-i-1] = (data[i] - 0x33) % 0xff
	}
	return rd
}

/***************************************读写接口实现**************************************************/
func (d *DDZY422N) crackbuf(inbuf []byte) (retbuf []byte) {
	log.Debugf("inbuf=% x", inbuf)
	for i, val := range inbuf {
		if i%2 == 0 {
			tbf := int(val) - 0x33 - 0x48 - i
			if tbf < 0 {
				tbf = tbf + 0x100
			}
			retbuf = append(retbuf, IntToBytes(tbf)[3])
		} else {
			tbf := int(val) - 0x33 - 0x54 - i
			if tbf < 0 {
				tbf = tbf + 0x100
			}
			retbuf = append(retbuf, IntToBytes(tbf)[3])
		}
	}
	log.Debugf("retbuf=% x", retbuf)
	return
}

func (d *DDZY422N) readCmd(taddr string, di []byte) []byte {
	saddr := fmt.Sprintf("%012s", taddr)
	cmd := []byte{0x68}
	for i := 12; i >= 2; i -= 2 {
		temp, _ := strconv.Atoi(saddr[i-2 : i])
		cmd = append(cmd, Bcd2Hex(IntToBytes(temp)[3]))
	}
	cmd = append(cmd, []byte{0x68, 0x01, 0x02}...)
	//cdi := d.plus33(di)
	cmd = append(cmd, di...)
	sum := d.rDateSum(cmd)
	cmd = append(cmd, sum)
	cmd = append(cmd, 0x16)
	rcmd := append([]byte{0xFE, 0xFE, 0xFE, 0xFE}, cmd...)
	return rcmd
}

func (d *DDZY422N) writeCmd(taddr string, di, value []byte) []byte {
	saddr := fmt.Sprintf("%012s", taddr)
	cmd := []byte{0x68}
	for i := 12; i >= 2; i -= 2 {
		temp, _ := strconv.Atoi(saddr[i-2 : i])
		cmd = append(cmd, Bcd2Hex(IntToBytes(temp)[3]))
	}
	cmd = append(cmd, []byte{0x68, 0x04}...)
	blen := 5 + len(value)
	cmd = append(cmd, IntToBytes(blen)[3])
	//cdi := d.plus33(di)
	cmd = append(cmd, di...)
	cmd = append(cmd, []byte{0xE2, 0xCD, 0xA0}...)
	var bvi int
	for j, v := range value {
		if (j % 2) == 0 {
			bvi = int(v) + 0x33 + 0x54 + j + 5
		} else {
			bvi = int(v) + 0x33 + 0x48 + j + 5
		}
		if bvi > 255 {
			bvi -= 256
		}
		cmd = append(cmd, IntToBytes(bvi)[3])
	}
	sum := d.rDateSum(cmd)
	cmd = append(cmd, sum)
	cmd = append(cmd, 0x16)
	rcmd := append([]byte{0xFE, 0xFE, 0xFE, 0xFE}, cmd...)
	return rcmd
}

func (d *DDZY422N) rDateSum(data []byte) byte {
	sum := 0
	for i := 0; i < len(data); i++ {
		sum += int(data[i])
	}
	return IntToBytes(sum & 0xff)[3]
}

func (d *DDZY422N) rdvalue(rsport serial.Port, cmd int, ret map[string]interface{}) (err error) {
	var di []byte
	results := make([]byte, 40)
	switch cmd {
	case 1:
		di = []byte{0xBA, 0x78}
	case 2:
		di = []byte{0x6A, 0x78}
	case 3:
		di = []byte{0x0A, 0x7C}
	case 4:
		di = []byte{0x2A, 0x7C}
	}
	for i := 0; i < 2; i++ {
		log.Debugf("send cmd = %x", d.readCmd(d.devaddr, di))
		if _, err = rsport.Write(d.readCmd(d.devaddr, di)); err != nil {
			log.Errorf("send cmd error  %s", err.Error())
			return err
		}
		time.Sleep(800 * time.Millisecond)
		var len int
		len, err = rsport.Read(results)
		if err == nil {
			var startbyte int
			for i, v := range results {
				if v == 0x68 {
					startbyte = i
					break
				}
			}
			log.Debugf("receive data = %x len = %d sum = %x ", results, len, d.rDateSum(results[startbyte:len-2]))
			if d.rDateSum(results[startbyte:len-2]) == results[len-2] {
				log.Debugf("校验正确")
				switch cmd {
				case 1:
					valb := d.crackbuf(results[startbyte+10 : startbyte+27])
					ret["剩余金额"] = float64(Hex2Bcd(valb[5]))*10000 +
						float64(Hex2Bcd(valb[4]))*100 +
						float64(Hex2Bcd(valb[3])) +
						float64(Hex2Bcd(valb[2]))/100
					ret["透支金额"] = float64(Hex2Bcd(valb[9]))*10000 +
						float64(Hex2Bcd(valb[8]))*100 +
						float64(Hex2Bcd(valb[7])) +
						float64(Hex2Bcd(valb[6]))/100
					ret["电量"] = float64(Hex2Bcd(valb[13]))*10000 +
						float64(Hex2Bcd(valb[12]))*100 +
						float64(Hex2Bcd(valb[11])) +
						float64(Hex2Bcd(valb[10]))/100
					//if valb[14]&0x40 == 0x40 {
					//ret["闸状态"] = "off"
					//} else {
					//ret["闸状态"] = "on"
					//}
				case 2:
					valb := d.crackbuf(results[startbyte+10 : startbyte+37])
					ret["报警金额"] = float64(Hex2Bcd(valb[3]))*100 +
						float64(Hex2Bcd(valb[2]))
					ret["报警负荷"] = float64(Hex2Bcd(valb[5])) +
						float64(Hex2Bcd(valb[4]))/100
					ret["允许透支金额"] = float64(Hex2Bcd(valb[7]))*100 +
						float64(Hex2Bcd(valb[6]))
					ret["允许囤积金额"] = float64(Hex2Bcd(valb[11]))*10000 +
						float64(Hex2Bcd(valb[10]))*100 +
						float64(Hex2Bcd(valb[9])) +
						float64(Hex2Bcd(valb[8]))/100
					ret["金额报警跳闸时间"] = float64(Hex2Bcd(valb[13]))*100 +
						float64(Hex2Bcd(valb[12]))
					ret["尖单价"] = float64(Hex2Bcd(valb[17])) +
						float64(Hex2Bcd(valb[16]))/100
					ret["峰单价"] = float64(Hex2Bcd(valb[20])) +
						float64(Hex2Bcd(valb[19]))/100
					ret["平单价"] = float64(Hex2Bcd(valb[23])) +
						float64(Hex2Bcd(valb[22]))/100
					ret["谷单价"] = float64(Hex2Bcd(valb[26])) +
						float64(Hex2Bcd(valb[25]))/100
				case 3:
					valb := d.crackbuf(results[startbyte+10 : startbyte+25])
					ret["网点编码"] = valb[2:4]
					ret["购电次数"] = int(Hex2Bcd(valb[14]))*100 + int(Hex2Bcd(valb[13]))
				case 4:
					valb := d.crackbuf(results[startbyte+10 : startbyte+25])
					ret["网点编码"] = valb[2:4]
					ret["购电次数"] = int(Hex2Bcd(valb[14]))*100 + int(Hex2Bcd(valb[13]))
				}
				err = nil
				break
			} else {
				err = errors.New("校验错误")
			}
		}
		if i < 1 {
			time.Sleep(300 * time.Millisecond)
		}
	}
	return
}

func (d *DDZY422N) decodeVal(cmd string, val interface{}) (retb []byte, err error) {
	switch cmd {
	case "初始化":
		if vm, ok := val.(map[string]interface{}); ok {
			if vf, ok := vm["报警金额"].(json.Number); ok {
				if tvf, err := vf.Int64(); err == nil {
					retb = append(retb, Bcd2Hex(IntToBytes(int(tvf) % 100)[3]))
					retb = append(retb, Bcd2Hex(IntToBytes(int(tvf) / 100)[3]))
				} else {
					return nil, err
				}
			} else {
				return nil, errors.New(`缺少字段"报警金额"`)
			}
		}
		if vm, ok := val.(map[string]interface{}); ok {
			if vf, ok := vm["报警负荷"].(json.Number); ok {
				if tvf, err := vf.Float64(); err == nil {
					retb = append(retb, Bcd2Hex(IntToBytes(int(tvf*100) % 100)[3]))
					retb = append(retb, Bcd2Hex(IntToBytes(int(tvf*100) / 100)[3]))
				} else {
					return nil, err
				}
			} else {
				return nil, errors.New(`缺少字段"报警负荷"`)
			}
		}
		if vm, ok := val.(map[string]interface{}); ok {
			if vf, ok := vm["允许透支金额"].(json.Number); ok {
				if tvf, err := vf.Int64(); err == nil {
					retb = append(retb, Bcd2Hex(IntToBytes(int(tvf) % 100)[3]))
					retb = append(retb, Bcd2Hex(IntToBytes(int(tvf) / 100)[3]))
				} else {
					return nil, err
				}
			} else {
				return nil, errors.New(`缺少字段"允许透支金额"`)
			}
		}
		if vm, ok := val.(map[string]interface{}); ok {
			if vf, ok := vm["允许囤积金额"].(json.Number); ok {
				if ftvf, err := vf.Float64(); err == nil {
					tvf := int(ftvf * 100)
					//retb = append(retb, Bcd2Hex(IntToBytes(int(tvf % 1000000))[3]))
					retb = append(retb, Bcd2Hex(IntToBytes(int(tvf % 100))[3]))
					retb = append(retb, Bcd2Hex(IntToBytes(int(tvf % 10000 / 100))[3]))
					retb = append(retb, Bcd2Hex(IntToBytes(int(tvf % 1000000 / 10000))[3]))
					retb = append(retb, Bcd2Hex(IntToBytes(int(tvf / 1000000))[3]))
				} else {
					return nil, err
				}
			} else {
				return nil, errors.New(`缺少字段"允许囤积金额"`)
			}
		}
		retb = append(retb, []byte{0x60, 0x0, 0x3, 0x0}...) //金额报警跳闸时间
		if vm, ok := val.(map[string]interface{}); ok {
			if vf, ok := vm["尖单价"].(json.Number); ok {
				if tvf, err := vf.Float64(); err == nil {
					retb = append(retb, Bcd2Hex(IntToBytes(int(tvf*100) % 100)[3]))
					retb = append(retb, Bcd2Hex(IntToBytes(int(tvf*100) / 100)[3]))
				} else {
					return nil, err
				}
			} else {
				return nil, errors.New(`缺少字段"尖单价"`)
			}
		}
		retb = append(retb, 0x1)
		if vm, ok := val.(map[string]interface{}); ok {
			if vf, ok := vm["峰单价"].(json.Number); ok {
				if tvf, err := vf.Float64(); err == nil {
					retb = append(retb, Bcd2Hex(IntToBytes(int(tvf*100) % 100)[3]))
					retb = append(retb, Bcd2Hex(IntToBytes(int(tvf*100) / 100)[3]))
				} else {
					return nil, err
				}
			} else {
				return nil, errors.New(`缺少字段"峰单价"`)
			}
		}
		retb = append(retb, 0x2)
		if vm, ok := val.(map[string]interface{}); ok {
			if vf, ok := vm["平单价"].(json.Number); ok {
				if tvf, err := vf.Float64(); err == nil {
					retb = append(retb, Bcd2Hex(IntToBytes(int(tvf*100) % 100)[3]))
					retb = append(retb, Bcd2Hex(IntToBytes(int(tvf*100) / 100)[3]))
				} else {
					return nil, err
				}
			} else {
				return nil, errors.New(`缺少字段"平单价"`)
			}
		}
		retb = append(retb, 0x3)
		if vm, ok := val.(map[string]interface{}); ok {
			if vf, ok := vm["谷单价"].(json.Number); ok {
				if tvf, err := vf.Float64(); err == nil {
					retb = append(retb, Bcd2Hex(IntToBytes(int(tvf*100) % 100)[3]))
					retb = append(retb, Bcd2Hex(IntToBytes(int(tvf*100) / 100)[3]))
				} else {
					return nil, err
				}
			} else {
				return nil, errors.New(`缺少字段"谷单价"`)
			}
		}
		log.Debugf("初始化 send = %x", retb)

	default:
		if vm, ok := val.(map[string]interface{}); ok {
			if vf, ok := vm["网点编码"].([]byte); ok {
				retb = append(retb, vf...)
			}
			ty, tm, td := time.Now().Date()
			th, tf, _ := time.Now().Clock()
			retb = append(retb, []byte{
				Bcd2Hex(IntToBytes(tf)[3]),
				Bcd2Hex(IntToBytes(th)[3]),
				Bcd2Hex(IntToBytes(td)[3]),
				Bcd2Hex(IntToBytes(int(tm))[3]),
				Bcd2Hex(IntToBytes(ty % 100)[3]),
				Bcd2Hex(IntToBytes(ty / 100)[3])}...)
			if vf, ok := vm["value"].(json.Number); ok {
				if tvf, err := vf.Float64(); err == nil {
					retb = append(retb, Bcd2Hex(IntToBytes(int(tvf*100) % 100)[3]))
					retb = append(retb, Bcd2Hex(IntToBytes(int(tvf*100) % 10000 / 100)[3]))
					retb = append(retb, Bcd2Hex(IntToBytes(int(tvf*100) / 10000)[3]))
				}
			} else {
				return nil, errors.New(`0.01<=x<=9999.99`)
			}
			if vf, ok := vm["购电次数"].(int); ok {
				retb = append(retb, Bcd2Hex(IntToBytes(int(vf+1) % 100)[3]))
				retb = append(retb, Bcd2Hex(IntToBytes(int(vf+1) / 100)[3]))
			}
		} else {
			return nil, errors.New(`参数错误`)
		}
		log.Debugf("充值或退款send = %x", retb)
	}
	return
}

func (d *DDZY422N) wtvalue(rsport serial.Port, cmd string, value interface{}, ret map[string]interface{}) (err error) {
	var di []byte
	var bval []byte
	var decodeval interface{}
	results := make([]byte, 40)
	switch cmd {
	case "初始化":
		di = []byte{0x6A, 0x78}
		decodeval = value
		bval, err = d.decodeVal(cmd, decodeval)
		if err != nil {
			return
		}
	case "充值":
		di = []byte{0x0A, 0x7C}
		err = d.rdvalue(rsport, 3, ret)
		if err != nil {
			return
		}
		ret["value"] = value
		decodeval = ret
		bval, err = d.decodeVal(cmd, decodeval)
		if err != nil {
			return
		}
		delete(ret, "value")
	case "退款":
		di = []byte{0x2A, 0x7C}
		err = d.rdvalue(rsport, 4, ret)
		if err != nil {
			return
		}
		ret["value"] = value
		decodeval = ret
		bval, err = d.decodeVal(cmd, decodeval)
		if err != nil {
			return
		}
		delete(ret, "value")
	case "强制合闸":
		di = []byte{0xDC, 0x78}
	case "强制断闸":
		di = []byte{0xDD, 0x78}
	case "撤销强制":
		di = []byte{0xDE, 0x78}
	default:
		return errors.New("错误的命令参数")
	}
	for i := 0; i < 2; i++ {
		log.Debugf("send cmd = %x", d.writeCmd(d.devaddr, di, bval))
		if _, err = rsport.Write(d.writeCmd(d.devaddr, di, bval)); err != nil {
			log.Errorf("send cmd error  %s", err.Error())
			return err
		}
		time.Sleep(800 * time.Millisecond)
		var len int
		len, err = rsport.Read(results)
		if err == nil {
			var startbyte int
			for i, v := range results {
				if v == 0x68 {
					startbyte = i
					break
				}
			}
			log.Debugf("receive data = %x len = %d sum = %x ", results, len, d.rDateSum(results[startbyte:len-2]))
			if d.rDateSum(results[startbyte:len-2]) == results[len-2] {
				log.Debugf("校验正确")
				err = nil
				break
			} else {
				err = errors.New("校验错误")
			}
		}
		if i < 1 {
			time.Sleep(300 * time.Millisecond)
		}
	}
	if cmd == "初始化" {
		err = d.rdvalue(rsport, 2, ret)
	} else if cmd == "充值" || cmd == "退款" {
		err = d.rdvalue(rsport, 1, ret)
	}
	return
}

// RWDevValue ..
func (d *DDZY422N) RWDevValue(rw string, m dict) (ret dict, err error) {
	sermutex := Mutex[d.commif]
	sermutex.Lock()
	defer sermutex.Unlock()
	defer func() {
		if driveErr := recover(); driveErr != nil {
			log.Errorf("drive programer  error : (%s)", driveErr)
			errstr := fmt.Sprintf("drive programer  error : %s", driveErr)
			err = errors.New(errstr)
		}
	}()
	serconfig := serial.Config{}
	serconfig.Address = Commif[d.commif]
	serconfig.BaudRate = 1200 //d.BaudRate
	serconfig.DataBits = 8    //d.DataBits
	serconfig.Parity = "E"    //d.Parity
	serconfig.StopBits = 1    // d.StopBits
	serconfig.Timeout = 2 * time.Second
	ret = map[string]interface{}{}
	ret["_devid"] = d.devid
	rsport, err := serial.Open(&serconfig)
	if err != nil {
		log.Errorf("open serial error %s", err.Error())
		return nil, err
	}
	defer rsport.Close()
	if rw == "r" {
		err = d.rdvalue(rsport, 1, ret)
		if err != nil {
			log.Errorf("read DDZY422N faild %s", err.Error())
			return nil, err
		}
		err = d.rdvalue(rsport, 2, ret)
		if err != nil {
			log.Errorf("read DDZY422N faild %s", err.Error())
			return nil, err
		}
	} else {
		for k, v := range m {
			if k != "_devid" {
				err = d.wtvalue(rsport, k, v, ret)
				if err != nil {
					log.Errorf("read DDZY422N faild %s", err.Error())
					return nil, err
				}
			}
		}
	}
	log.Info(ret)
	return ret, nil
}

/***************************************读写接口实现**************************************************/
