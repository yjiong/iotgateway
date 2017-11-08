package device

import (
	//"encoding/json"
	//"errors"
	"errors"
	"fmt"
	"strconv"
	"time"
	//	"sync"
	log "github.com/Sirupsen/logrus"
	//simplejson "github.com/bitly/go-simplejson"
	"github.com/yjiong/go_tg120/serial"
)

//var mutex sync.Mutex
type DTSD422 struct {
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
	RegDevice["DTSD422"] = &DTSD422{}
}

func (d *DTSD422) NewDev(id string, ele map[string]string) (DeviceRWer, error) {
	ndev := new(DTSD422)
	ndev.Device = d.Device.NewDev(id, ele)
	/***********************初始化设备的特有的参数*****************************/
	//ndev.BaudRate, _ = strconv.Atoi(ele["BaudRate"])
	//ndev.DataBits, _ = strconv.Atoi(ele["DataBits "])
	//ndev.StopBits, _ = strconv.Atoi(ele["StopBits"])
	//ndev.Parity, _ = ele["Parity"]
	/***********************初始化设备的特有的参数*****************************/
	return ndev, nil
}

func (d *DTSD422) GetElement() (dict, error) {
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
func (d *DTSD422) HelpDoc() interface{} {
	conn := dict{
		"devaddr": "设备地址",
		/***********DTSD422设备的参数*****************************/
		/***********读取设备的参数*****************************/
	}
	r_parameter := dict{
		"_devid": "被读取设备对象的id",
		/***********读取设备的参数*****************************/
		/***********读取设备的参数*****************************/
	}
	data := dict{
		"_devid": "添加设备对象的id",
		"_type":  "DTSD422", //设备类型
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
	helpdoc := dict{
		"1.添加设备": dev_update,
		"2.读取设备": readdev,
	}
	return helpdoc
}

/***********************设备的参数说明帮助***********************************/

/***************************************添加设备参数检验**********************************************/
func (d *DTSD422) CheckKey(ele dict) (bool, error) {
	return true, nil
}

/***************************************添加设备参数检验**********************************************/
func (d *DTSD422) plus33(data []byte) []byte {
	rd := make([]byte, len(data))
	for i, _ := range data {
		rd[len(data)-i-1] = (data[i] + 0x33) % 0xff
	}
	return rd
}

func (d *DTSD422) sub33(data []byte) []byte {
	rd := make([]byte, len(data))
	for i, _ := range data {
		rd[len(data)-i-1] = (data[i] - 0x33) % 0xff
	}
	return rd
}

/***************************************读写接口实现**************************************************/
func (d *DTSD422) read_cmd(taddr string, di []byte) []byte {
	saddr := fmt.Sprintf("%012s", taddr)
	cmd := []byte{0x68}
	for i := 12; i >= 2; i -= 2 {
		temp, _ := strconv.Atoi(saddr[i-2 : i])
		cmd = append(cmd, Bcd2Hex(IntToBytes(temp)[3]))
	}
	cmd = append(cmd, []byte{0x68, 0x11, 0x04}...)
	cdi := d.plus33(di)
	cmd = append(cmd, cdi...)
	sum := d.r_data_sum(cmd)
	cmd = append(cmd, sum)
	cmd = append(cmd, 0x16)
	rcmd := append([]byte{0xFE, 0xFE, 0xFE, 0xFE}, cmd...)
	return rcmd
}

func (d *DTSD422) r_data_sum(data []byte) byte {
	sum := 0
	for i := 0; i < len(data); i++ {
		sum += int(data[i])
	}
	return IntToBytes(sum & 0xff)[3]
}
func (d *DTSD422) RWDevValue(rw string, m dict) (ret dict, err error) {
	log.SetLevel(log.DebugLevel)
	di_Forward_active_power := []byte{0x00, 0x01, 0x00, 0x00} //正向有功电能数据(ff块)
	//di_Voltages := []byte{0x02, 0x01, 0xFF, 0x00}             //电压数据块
	//di_Currents := []byte{0x02, 0x02, 0xFF, 0x00}             //电流数据块
	//di_PEs := []byte{0x02, 0x06, 0xFF, 0x00}                  //功率因素数据块
	//di_Psums := []byte{0x02, 0x03, 0xFF, 0x00}                //舜时总有功率数据块
	sermutex := Mutex[d.commif]
	sermutex.Lock()
	defer sermutex.Unlock()
	serconfig := serial.Config{}
	serconfig.Address = Commif[d.commif]
	serconfig.BaudRate = 2400 //d.BaudRate
	serconfig.DataBits = 8    //d.DataBits
	serconfig.Parity = "E"    //d.Parity
	serconfig.StopBits = 1    // d.StopBits
	//slaveid, _ := strconv.Atoi(d.devaddr)
	taddr := d.devaddr
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
		results := make([]byte, 40)
		for i := 0; i < 2; i++ {
			log.Debugf("send cmd = %x", d.read_cmd(taddr, di_Forward_active_power))
			if _, ok := rsport.Write(d.read_cmd(taddr, di_Forward_active_power)); ok != nil {
				log.Errorf("send cmd error  %s", ok.Error())
				return nil, ok
			}
			time.Sleep(500 * time.Millisecond)
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
				log.Debugf("receive data = %x len = %d sum = %x ", results, len, d.r_data_sum(results[startbyte:len-2]))
				if d.r_data_sum(results[startbyte:len-2]) == results[len-2] {
					log.Debugf("校验正确")
					valb := d.sub33(results[startbyte+14 : len-2])
					log.Debugf("%x", valb)
					ret["正向有功电能"] = float64(Hex2Bcd(valb[0]))*10000 +
						float64(Hex2Bcd(valb[1]))*100 +
						float64(Hex2Bcd(valb[2])) +
						float64(Hex2Bcd(valb[3]))/100
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
		if err != nil {
			log.Errorf("read DTSD422 faild %s", err.Error())
			return nil, err
		}
		log.Info(ret)
	}
	return ret, nil
}

/***************************************读写接口实现**************************************************/
