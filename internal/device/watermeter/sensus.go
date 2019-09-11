package watermeter

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/yjiong/iotgateway/internal/device"
	"github.com/yjiong/iotgateway/serial"
	"strconv"
	"time"
)

//SENSUS ..
type SENSUS struct {
	//组合Device
	device.Device
	/**************按不同设备自定义*************************/
	//BaudRate int
	//DataBits int
	//StopBits int
	//Parity   string
	/**************按不同设备自定义*************************/
}

func init() {
	device.RegDevice["SENSUS"] = &SENSUS{}
}

// NewDev ..
func (d *SENSUS) NewDev(id string, ele map[string]string) (device.Devicerwer, error) {
	ndev := new(SENSUS)
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
func (d *SENSUS) GetElement() (device.Dict, error) {
	conn := device.Dict{
		/***********************设备的特有的参数*****************************/
		device.DevAddr: d.Devaddr,
		"commif":       d.Commif,
		//"BaudRate": d.BaudRate,
		//"DataBits": d.DataBits,
		//"StopBits": d.StopBits,
		//"Parity":   d.Parity,
		/***********************设备的特有的参数*****************************/
	}
	data := device.Dict{
		device.DevID:   d.Devid,
		device.DevType: d.Devtype,
		device.DevConn: conn,
	}
	return data, nil
}

/***********************设备的参数说明帮助***********************************/

// HelpDoc ..
func (d *SENSUS) HelpDoc() interface{} {
	conn := device.Dict{
		device.DevAddr: "设备地址",
		/***********TOSHIBA设备的参数*****************************/
		"commif": "通信接口,比如(rs485-1)",
		/***********TOSHIBA设备的参数*****************************/
	}
	rParameter := device.Dict{
		device.DevID: "被读取设备对象的id",
		/***********读取设备的参数*****************************/
		/***********读取设备的参数*****************************/
	}
	wParameter := device.Dict{
		device.DevID: "被操作设备对象的id",
		/***********操作设备的参数*****************************/
		//"初始化":  initval,
	}
	data := device.Dict{
		device.DevID:   "添加设备对象的id",
		device.DevType: "SENSUS", //设备类型
		device.DevConn: conn,
	}
	devUpdate := device.Dict{
		"request": device.Dict{
			"cmd":  device.UpdateDevItem,
			"data": data,
		},
	}
	readdev := device.Dict{
		"request": device.Dict{
			"cmd":  device.GetDevVar,
			"data": rParameter,
		},
	}
	writedev := device.Dict{
		"request": device.Dict{
			"cmd":  device.SetDevVar,
			"data": wParameter,
		},
	}
	helpdoc := device.Dict{
		"1.添加设备": devUpdate,
		"2.读取设备": readdev,
		"3.操作设备": writedev,
	}
	return helpdoc
}

/***********************设备的参数说明帮助***********************************/

/***************************************添加设备参数检验**********************************************/

// CheckKey ..
func (d *SENSUS) CheckKey(ele device.Dict) (bool, error) {
	im, imok := ele["commif"].(string)
	if !imok {
		return false, fmt.Errorf("device must have string type element 通讯接口 :commif")
	}
	if _, ok := device.Commif[im]; !ok {
		return false, fmt.Errorf("通讯接口:%s不存在", im)
	}
	return true, nil
}

/***************************************添加设备参数检验**********************************************/
func (d *SENSUS) rDateSum(data []byte) byte {
	sum := 0
	for i := 0; i < len(data); i++ {
		sum += int(data[i])
	}
	return device.IntToBytes(sum & 0xff)[3]
}

func (d *SENSUS) baseCmd(cmdid []byte) []byte {
	saddr := fmt.Sprintf("%03s", d.Devaddr)
	cmd := cmdid
	temp, _ := strconv.Atoi(saddr)
	cmd = append(cmd, device.IntToBytes(temp)[3])
	sum := d.rDateSum(cmd[1:])
	cmd = append(cmd, sum)
	cmd = append(cmd, 0x16)
	return cmd
}

func (d *SENSUS) readCmd() []byte {
	return d.baseCmd([]byte{0x10, 0x5B})
}

func (d *SENSUS) activeCmd() []byte {
	return d.baseCmd([]byte{0x10, 0x40})
}

func (d *SENSUS) rdvalue(rsport serial.Port, cmd int, ret map[string]interface{}) (err error) {
	results := make([]byte, 256)
	for i := 0; i < 2; i++ {
		if _, err = rsport.Write(d.activeCmd()); err != nil {
			log.Errorf("send active register error  %s", err.Error())
			return err
		}
		_, err = device.SerialRead(rsport, time.Second, results)
		log.Debugf("active receive = %x", results[0])
		if results[0] != 0xe5 {
			return errors.New("active register faild")
		}
		log.Debugf("send cmd = % x", d.readCmd())
		if _, err = rsport.Write(d.readCmd()); err != nil {
			log.Errorf("send cmd error  %s", err.Error())
			return err
		}
		var rlen int
		time.Sleep(600 * time.Millisecond)
		rlen, err = device.SerialRead(rsport, time.Second, results)
		if err == nil {
			var startbyte int
			count := 0
			for i, v := range results {
				if v == 0x68 {
					count++
					if count == 2 {
						startbyte = 1 + i
						break
					}
				}
			}
			log.Debugf("receive data = % x rlen = %d sum = %x ", results[:rlen], rlen, d.rDateSum(results[startbyte:rlen-2]))
			if d.rDateSum(results[startbyte:rlen-2]) == results[rlen-2] && results[startbyte] == 0x08 && results[rlen-1] == 0x16 && rlen > 9 {
				log.Debugf("校验正确")
				valb := results[startbyte : results[1]+4]
				log.Debugf("08-1f=% x", valb)
				if d.Devtype == "ZENNER-M" {
					tempint, _ := strconv.ParseInt(device.HexStringReverse(fmt.Sprintf("%x", valb[29:33])), 16, 0)
					ret["当前读数"] = float64(tempint) / 1000
					//ret["当前读数"] = (float64(valb[32])*0x1000000 + float64(valb[31])*0x10000 + float64(valb[30])*0x100 + float64(valb[29])) / 1000
				} else {
					ret["水表ID"] = device.HexStringReverse(fmt.Sprintf("%x", valb[3:7]))
					tempint, _ := strconv.Atoi(device.HexStringReverse(fmt.Sprintf("%x", valb[17:21])))
					ret["正向流量"] = float64(tempint) / 100
					tempint, _ = strconv.Atoi(device.HexStringReverse(fmt.Sprintf("%x", valb[24:28])))
					ret["反向流量"] = float64(tempint) / 100
					tempint, _ = strconv.Atoi(device.HexStringReverse(fmt.Sprintf("%x", valb[30:34])))
					ret["流速"] = float64(tempint) / 100
					ret["制造编号"] = device.HexStringReverse(fmt.Sprintf("%x", valb[36:40]))
					ret["用户地点"] = device.HexStringReverse(fmt.Sprintf("%x", valb[43:47]))
					ret["表内时钟"] = fmt.Sprintf("2%03d年%d月%d日%d时%d分",
						(valb[52]&0xf0>>4)*8+(valb[51]&0xf0>>4)/2,
						valb[52]&0xf, ((valb[51]&0xf0)%0x20)+valb[51]&0xf, valb[50], valb[49])
					//ret["报警状态代码"] = fmt.Sprintf("%02x%02x%02x%02x", valb[59], valb[58], valb[57], valb[56])
					ret["报警状态代码"] = fmt.Sprintf("%02x%02x", valb[57], valb[56])
					hisDate1 := fmt.Sprintf("2%03d年%d月%d日%d时%d分",
						(valb[66]&0xf0>>4)*8+(valb[65]&0xf0>>4)/2,
						valb[66]&0xf, ((valb[65]&0xf0)%0x20)+valb[65]&0xf, valb[64], valb[63])
					tempint, _ = strconv.Atoi(device.HexStringReverse(fmt.Sprintf("%x", valb[70:74])))
					//log.Debugln("hisDate1=", hisDate1, "流量=", tempint)
					ret[`历史记录(`+hisDate1+`)正向流量`] = float64(tempint) / 100
					hisDate2 := fmt.Sprintf("2%03d年%d月%d日%d时%d分",
						(valb[80]&0xf0>>4)*8+(valb[79]&0xf0>>4)/2,
						valb[80]&0xf, ((valb[79]&0xf0)%0x20)+valb[79]&0xf, valb[78], valb[77])
					//log.Debugln("hisDate2", hisDate2)
					tempint, _ = strconv.Atoi(device.HexStringReverse(fmt.Sprintf("%x", valb[85:89])))
					ret[`历史记录(`+hisDate2+`)正向流量`] = float64(tempint) / 100
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

// RWDevValue ..
func (d *SENSUS) RWDevValue(rw string, m device.Dict) (ret device.Dict, err error) {
	serMutex := device.Mutex[d.Commif]
	serMutex.Lock()
	defer serMutex.Unlock()
	defer func() {
		if driveErr := recover(); driveErr != nil {
			log.Errorf("drive programer  error : (%s)", driveErr)
			errstr := fmt.Sprintf("drive programer  error : %s", driveErr)
			err = errors.New(errstr)
		}
	}()
	serconfig := serial.Config{}
	serconfig.Address = device.Commif[d.Commif]
	serconfig.BaudRate = 2400 //d.BaudRate
	serconfig.DataBits = 8    //d.DataBits
	serconfig.Parity = "E"    //d.Parity
	serconfig.StopBits = 1    // d.StopBits
	serconfig.Timeout = time.Microsecond * time.Duration(500000000/serconfig.BaudRate)
	ret = map[string]interface{}{}
	ret[device.DevID] = d.Devid
	rsport, err := serial.Open(&serconfig)
	if err != nil {
		log.Errorf("open serial error %s", err.Error())
		return nil, err
	}
	defer rsport.Close()
	if rw == "r" {
		err = d.rdvalue(rsport, 1, ret)
		if err != nil {
			log.Errorf("read device %s-%s faild %s", d.Devtype, d.Devid, err.Error())
			return nil, err
		}
	}
	//log.Debugln(ret)
	return ret, nil
}

/***************************************读写接口实现**************************************************/
