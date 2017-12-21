package device

import (
	"encoding/json"
	//"math/rand"
	//"errors"
	//"fmt"
	//"strconv"
	//"time"

	//	"sync"
	log "github.com/Sirupsen/logrus"
	simplejson "github.com/bitly/go-simplejson"
	//"github.com/yjiong/go_tg120/serial"
)

//var mutex sync.Mutex
type TEST_GO struct {
	//继承于Device
	Device
	/**************按不同设备自定义*************************/
	add   int
	setnm interface{}
	//DataBits int
	//StopBits int
	//Parity   string

	/**************按不同设备自定义*************************/
}

func init() {
	RegDevice["TEST_GO"] = &TEST_GO{}
}

func (d *TEST_GO) NewDev(id string, ele map[string]string) (Devicerwer, error) {
	ndev := new(TEST_GO)
	ndev.Device = d.Device.NewDev(id, ele)
	/***********************初始化设备的特有的参数*****************************/
	d.add = 0
	d.setnm = "hahaha"
	//ndev.BaudRate, _ = strconv.Atoi(ele["BaudRate"])
	//ndev.DataBits, _ = strconv.Atoi(ele["DataBits "])
	//ndev.StopBits, _ = strconv.Atoi(ele["StopBits"])
	//ndev.Parity, _ = ele["Parity"]
	/***********************初始化设备的特有的参数*****************************/
	return ndev, nil
}

func (d *TEST_GO) GetElement() (dict, error) {
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
func (d *TEST_GO) HelpDoc() interface{} {
	conn := dict{
		"devaddr": "设备地址",
		/***********TEST_GO设备的参数*****************************/
		/***********读取设备的参数*****************************/
	}
	r_parameter := dict{
		"_devid": "被读取设备对象的id",
		/***********读取设备的参数*****************************/
		/***********读取设备的参数*****************************/
	}
	data := dict{
		"_devid": "添加设备对象的id",
		"_type":  "TEST_GO", //设备类型
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
func (d *TEST_GO) CheckKey(ele dict) (bool, error) {
	return true, nil
}

/***************************************添加设备参数检验**********************************************/

/***************************************读写接口实现**************************************************/
func (d *TEST_GO) read_cmd(taddr byte) []byte {
	sum := (0x81 + int(taddr)) & 0xff
	cmd := []byte{0xa5, 0x81, taddr, 0x00, 0x00, IntToBytes(sum)[3], 0x5a}
	return cmd
}

func (d *TEST_GO) r_data_sum(data []byte) byte {
	sum := 0
	for i := 1; i < 9; i++ {
		sum += int(data[i])
	}
	return IntToBytes(sum & 0xff)[3]
}
func (d *TEST_GO) RWDevValue(rw string, m dict) (ret dict, err error) {
	//log.SetLevel(log.DebugLevel)
	sermutex := Mutex[d.commif]
	sermutex.Lock()
	defer sermutex.Unlock()
	//serconfig := serial.Config{}
	//serconfig.Address = Commif[d.commif]
	//serconfig.BaudRate = 9600 //d.BaudRate
	//serconfig.DataBits = 8    //d.DataBits
	//serconfig.Parity = "N"    //d.Parity
	//serconfig.StopBits = 1    // d.StopBits
	//slaveid, _ := strconv.Atoi(d.devaddr)
	//taddr := byte(slaveid)
	//serconfig.Timeout = 2 * time.Second
	ret = map[string]interface{}{}
	ret["_devid"] = d.devid
	//rsport, err := serial.Open(&serconfig)
	//if err != nil {
	//log.Errorf("open serial error %s", err.Error())
	//return nil, err
	//}
	//defer rsport.Close()

	if rw == "r" {
		//results := make([]byte, 11)
		//for i := 0; i < 2; i++ {
		//log.Debugf("send cmd = %x", d.read_cmd(taddr))
		//if _, ok := rsport.Write(d.read_cmd(taddr)); ok != nil {
		//log.Errorf("send cmd error  %s", ok.Error())
		//return nil, ok
		//}
		//time.Sleep(100 * time.Millisecond)
		//var len int
		//len, err = rsport.Read(results)
		////if err != nil {
		////log.Errorf("serial read error  %s", err.Error())
		////}
		//if err == nil && len == 11 && results[9] == d.r_data_sum(results) {
		//log.Debugf("receive data = %x len = %d sum = %x", results, len, d.r_data_sum(results))
		//break
		//}
		//if i < 1 {
		//time.Sleep(10 * time.Second)
		//}
		//}
		//if err != nil {
		//log.Errorf("read TEST_GO faild %s", err.Error())
		//return nil, err
		//} else {
		d.add += 1
		if d.add > 100 {
			d.add = 0
		}
		ret["递增数"] = d.add
		ret["设置数"] = d.setnm
		//}
		jsret, _ := json.Marshal(ret)
		inforet, _ := simplejson.NewJson(jsret)
		pinforet, _ := inforet.EncodePretty()
		log.Info(string(pinforet))
	} else {
		d.setnm = m["_varvalue"]
	}
	return ret, nil
}

/***************************************读写接口实现**************************************************/
