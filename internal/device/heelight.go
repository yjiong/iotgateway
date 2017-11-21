package device

import (
	"encoding/json"
	//"errors"
	//"strconv"
	"fmt"
	"time"
	//	"sync"
	log "github.com/Sirupsen/logrus"
	simplejson "github.com/bitly/go-simplejson"
	"github.com/yjiong/go_tg120/serial"
)

//var mutex sync.Mutex
type HEELIGHT struct {
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
	RegDevice["HEELIGHT"] = &HEELIGHT{}
}

func (d *HEELIGHT) NewDev(id string, ele map[string]string) (DeviceRWer, error) {
	ndev := new(HEELIGHT)
	ndev.Device = d.Device.NewDev(id, ele)
	/***********************初始化设备的特有的参数*****************************/
	//ndev.BaudRate, _ = strconv.Atoi(ele["BaudRate"])
	//ndev.DataBits, _ = strconv.Atoi(ele["DataBits "])
	//ndev.StopBits, _ = strconv.Atoi(ele["StopBits"])
	//ndev.Parity, _ = ele["Parity"]
	/***********************初始化设备的特有的参数*****************************/
	return ndev, nil
}

func (d *HEELIGHT) GetElement() (dict, error) {
	conn := dict{
		/***********************设备的特有的参数*****************************/
		"commif": d.commif,
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
func (d *HEELIGHT) HelpDoc() interface{} {
	conn := dict{
		"devaddr": "设备地址",
		/***********HEELIGHT设备的参数*****************************/
		/***********读取设备的参数*****************************/
	}
	r_parameter := dict{
		"_devid": "被读取设备对象的id",
		/***********读取设备的参数*****************************/
		/***********读取设备的参数*****************************/
	}
	data := dict{
		"_devid": "添加设备对象的id",
		"_type":  "HEELIGHT", //设备类型
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
func (d *HEELIGHT) CheckKey(ele dict) (bool, error) {
	return true, nil
}

/***************************************添加设备参数检验**********************************************/

/***************************************读写接口实现**************************************************/

func (d *HEELIGHT) RWDevValue(rw string, m dict) (ret dict, err error) {
	//log.SetLevel(log.DebugLevel)
	sermutex := Mutex[d.commif]
	sermutex.Lock()
	defer sermutex.Unlock()
	serconfig := serial.Config{}
	serconfig.Address = Commif[d.commif]
	serconfig.BaudRate = 9600 //d.BaudRate
	serconfig.DataBits = 8    //d.DataBits
	serconfig.Parity = "N"    //d.Parity
	serconfig.StopBits = 1    // d.StopBits
	serconfig.Timeout = 30 * time.Second
	ret = map[string]interface{}{}
	ret["_devid"] = d.devid
	rsport, err := serial.Open(&serconfig)
	if err != nil {
		log.Errorf("open serial error %s", err.Error())
		return nil, err
	}
	defer rsport.Close()

	if rw == "r" {
		results := make([]byte, 32)
		var len int
		var rstr string
		len, err = rsport.Read(results)
		if err == nil && len > 3 {
			log.Debugf("receive data = %x len = %d ", results, len)
			for i := 2; i < len; i++ {
				rstr += fmt.Sprintf("%x", results[i])
			}
			ret["cmd_id"] = rstr
		} else {
			log.Errorf("read HEELIGHT faild ")
			ret["_update"] = false
			return ret, nil
		}
		jsret, _ := json.Marshal(ret)
		inforet, _ := simplejson.NewJson(jsret)
		pinforet, _ := inforet.EncodePretty()
		log.Info(string(pinforet))
	}
	return ret, nil
}

/***************************************读写接口实现**************************************************/
