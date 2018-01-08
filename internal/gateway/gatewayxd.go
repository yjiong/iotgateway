package gateway

import (
	"bytes"
	"container/list"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"reflect"
	"regexp"
	"strconv"
	"time"
	//	"strings"

	log "github.com/Sirupsen/logrus"
	simplejson "github.com/bitly/go-simplejson"
	"github.com/yjiong/go_tg120/config"
	"github.com/yjiong/go_tg120/gpio"
	"github.com/yjiong/go_tg120/internal/common"
	"github.com/yjiong/go_tg120/internal/device"
	"github.com/yjiong/go_tg120/internal/handler"
	"golang.org/x/net/websocket"
)

func init() {
	Reset, err := gpio.OpenPin(65, gpio.ModeInput)
	if err != nil {
		return
	}
	Run, err := gpio.OpenPin(67, gpio.ModeOutput)
	if err != nil {
		return
	}
	Link, err := gpio.OpenPin(66, gpio.ModeOutput)
	if err != nil {
		return
	}
	go func() {
		for {
			resetdefip(Reset)
		}
	}()
	go func() {
		for {
			ledLink(common.Mqttconnected, Link)
		}
	}()
	go func() {
		for {
			ledRun(Run, 500)
		}
	}()
}

func resetdefip(gp gpio.Pin) {
	if !gp.Get() {
		time.Sleep(5 * time.Second)
		if !gp.Get() {
			ipconstr := "auto lo\n" +
				"iface lo inet loopback\n" +
				"auto eth0\n" +
				"allow-hotplug eth0\n" +
				"iface eth0 inet static\n" +
				"address 192.168.1.188\n" +
				"netmask 255.255.255.0\n" +
				"gateway 192.168.1.1\n" +
				"auto wlan0\n" +
				"iface wlan0 inet static\n" +
				"address 192.168.8.1\n" +
				"netmask 255.255.255.0\n"
			if _, err := os.Stat(common.INTERFACES); err != nil {
				if os.IsNotExist(err) {
					f, _ := os.Create(common.INTERFACES)
					if _, err := f.WriteString(ipconstr); err != nil {
						log.Errorf("reset default ip config failed :%s", err)
						return
					}
					f.Sync()
					f.Close()
				}
			} else {
				os.Remove(common.INTERFACES)
				f, _ := os.OpenFile(common.INTERFACES, os.O_WRONLY|os.O_CREATE, 0666)
				if _, err := io.WriteString(f, ipconstr); err != nil {
					log.Errorf("reset default ip config failed :%s", err)
					return
				}
				f.Sync()
				f.Close()
			}
			cmd := exec.Command("reboot")
			var out bytes.Buffer
			cmd.Stdin = os.Stdin
			cmd.Stdout = &out
			cmd.Run()
		}
	} else {
		time.Sleep(1 * time.Second)
	}
}

func ledRun(gp gpio.Pin, delay int64) {
	time.Sleep(time.Duration(delay) * time.Millisecond)
	if gp.Get() {
		gp.Clear()
	} else {
		gp.Set()
	}
}

func ledLink(ml bool, gp gpio.Pin) {
	var delay int64
	if ml {
		delay = 1000
	} else {
		delay = 200
	}
	time.Sleep(time.Duration(delay) * time.Millisecond)
	if gp.Get() {
		gp.Clear()
	} else {
		gp.Set()
	}
}

// Cmdfp struct
type Cmdfp struct {
	Cmdfunc func(*simplejson.Json) error
	Param   *simplejson.Json
}

type dict map[string]interface{}

// Gateway struct
type Gateway struct {
	DevIfMap   map[string]device.Devicerwer //设备接口
	ConMap     map[string]string            //配置参数
	Handler    handler.Handler              //消息处理
	WsMap      map[int]*websocket.Conn
	Cmdlist    *list.List //命令队列
	Cmdchan    chan Cmdfp //命令chan
	WsNochanr  chan map[int]string
	Devpath    string
	Conpath    string
	loop       bool
	serincount int
	delay      int
}

// Update ..
func (mygw *Gateway) Update() error {
	var err error
	mygw.DevIfMap, err = device.NewDevHandler(mygw.Devpath)
	mygw.ConMap, err = common.NewConMap(mygw.Conpath)
	return err
}

//WsHandle ..
func (mygw *Gateway) WsHandle(ws *websocket.Conn) {
	var err error
	no := len(mygw.WsMap) + 1
	mygw.WsMap[no] = ws
	for {
		var reply string
		if err = websocket.Message.Receive(ws, &reply); err != nil {
			log.Info("Websocket disconnect or Message Can't receive")
			delete(mygw.WsMap, no)
			break
		}
		log.Info("websocket message received = %s", reply)
		//notice := `若您没有接收到请求应答,请发送{"request":{"cmd":"help"}}查看命令帮助`
		//if err = websocket.Message.Send(ws, notice); err != nil {
		//log.Infof("Websocket Message send error : %s", err)
		//delete(mygw.WsMap, no)
		//break
		//}
		wsnochan := map[int]string{
			no: reply,
		}
		mygw.WsNochanr <- wsnochan
	}
}

// Mqttcmdhandler ..
func (mygw *Gateway) Mqttcmdhandler(dpc chan handler.DataDownPayload) {
	for dpj := range dpc {
		go func(dpj handler.DataDownPayload) {
			request := dpj.Pj.Get("request")
			if _, remok := request.Map(); remok == nil {
				mygw.msghandler(request, nil)
			} else {
				mygw.msghandler(dpj.Pj, nil)
			}
		}(dpj)
	}
}

// Wscmdhandler ..
func (mygw *Gateway) Wscmdhandler(req string, ws *websocket.Conn) {
	if reqjs, err := simplejson.NewJson([]byte(req)); err == nil {
		request := reqjs.Get("request")
		mygw.msghandler(request, ws)
	}
}

func (mygw *Gateway) msghandler(request *simplejson.Json, ws *websocket.Conn) {
	cmd := request.Get("cmd")
	cmdstring, err := cmd.String()
	switch {
	case cmdstring == "init/set.do":
		if err := mygw.initset(request, ws); err != nil {
			log.Errorf("init/set.do error :%s", err)
		}

	case cmdstring == "init/get.do":
		mygw.initget(request, ws)

	case cmdstring == "manager/get_suppot_devlist":
		mygw.managerDevSuppotlist(request, ws)

	case cmdstring == "manager/dev/update.do":
		mygw.managerDevUpdate(request, ws)

	case cmdstring == "manager/update_commif.do":
		mygw.managerUpdatecommif(request, ws)

	case cmdstring == "manager/list_commif.do":
		mygw.managerListcommif(request, ws)

	case cmdstring == "manager/dev/list.do":
		mygw.managerDevList(request, ws)

	case cmdstring == "manager/dev/delete.do":
		mygw.managerDevDelete(request, ws)

	case cmdstring == "manager/set_system_time":
		mygw.managerSetSystemTIme(request, ws)

	case cmdstring == "manager/set_interval.do":
		mygw.managerSetInterval(request, ws)

	case cmdstring == "manager/update_drive":
		mygw.managerUpdateDrive(request, ws)

	case cmdstring == "do/getvar":
		mygw.doGetvar(request, ws)

	case cmdstring == "do/setvar":
		mygw.doSetvar(request, ws)

	case cmdstring == "help":
		mygw.doHelp(request, ws)

	case cmdstring == "remoteSerial":
		mygw.remoteSerial(request)
	default:
		log.Errorf("cmd error %s", err)
	}
}

func (mygw *Gateway) updateRstat() {
	conf, _ := config.LoadConfigFile(mygw.Conpath)
	rs, _ := conf.GetValue("other", "runstate")
	rsi, _ := strconv.Atoi(rs)
	urs := strconv.Itoa(rsi + 1)
	conf.SetValue("other", "runstate", urs)
	config.SaveConfigFile(conf, mygw.Conpath)
	mygw.Update()
}

func (mygw *Gateway) initset(req *simplejson.Json, ws *websocket.Conn) error {
	if _, err := os.Stat(mygw.Conpath); err != nil {
		if os.IsNotExist(err) {
			f, _ := os.Create(mygw.Conpath)
			f.WriteString("[mqtt]\n[other]\n[commif]\n")
			f.Sync()
			f.Close()
		}
	}
	conf, _ := config.LoadConfigFile(mygw.Conpath)
	//	defer config.SaveConfigFile(conf, mygw.Conpath)
	data := req.Get("data")
	patternip := regexp.MustCompile(`^(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])\.(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])\.(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])\.(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])$`)
	patternport := regexp.MustCompile(`^[1-9]\d{0,4}$`)
	//修改config文件
	var ack []string
	if serverIP, ok := data.Get("_server_ip").String(); ok == nil {
		if patternip.MatchString(serverIP) {
			conf.SetValue("mqtt", "_server_ip", serverIP)
			ack = append(ack, "_server_ip")
		} else {
			mygw.encoderesponseup(req, fmt.Sprintf("_server_ip 格式错误: %s", serverIP), 1, ws)
			return errors.New("wrong mqttserver ip format")
		}
	}
	if port, ok := data.Get("_server_port").String(); ok == nil {
		if patternport.MatchString(port) {
			conf.SetValue("mqtt", "_server_port", port)
			ack = append(ack, "_server_port")
		} else {
			mygw.encoderesponseup(req, fmt.Sprintf("_server_port 格式错误: %s", port), 1, ws)
			return errors.New("wrong mqtt port format")
		}
	}
	if username, ok := data.Get("_username").String(); ok == nil {
		conf.SetValue("mqtt", "_username", username)
		ack = append(ack, "_username")
	}
	if passwd, ok := data.Get("_password").String(); ok == nil {
		conf.SetValue("mqtt", "_password", passwd)
		ack = append(ack, "_password")
	}
	if topic, ok := data.Get("_server_name").String(); ok == nil {
		conf.SetValue("mqtt", "_server_name", topic)
		ack = append(ack, "_server_name")
	}
	if will, ok := data.Get("_will").String(); ok == nil {
		if iwill, err := strconv.Atoi(will); err == nil && iwill <= 1 {
			conf.SetValue("mqtt", "_will", will)
			ack = append(ack, "_will")
		}
	}
	if keepalive, ok := data.Get("_keepalive").String(); ok == nil {
		if ik, err := strconv.Atoi(keepalive); err == nil && ik >= 0 {
			conf.SetValue("mqtt", "keepalive", keepalive)
			ack = append(ack, "_skeepalive")
		}
	}
	config.SaveConfigFile(conf, mygw.Conpath)
	//设置/etc/network/interfaces
	inet, _ := data.Get("_interface_inet").String()
	if inet != "" {
		if inet == "static" || inet == "dhcp" {
			ack = append(ack, "_interface_inet")
		} else {
			mygw.encoderesponseup(req, fmt.Sprintf("_interface_inet 格式错误: %s [static or dhcp] ", inet), 1, ws)
			return errors.New("wrong client interface inet format")
		}
	}

	address, _ := data.Get("_client_ip").String()
	if address != "" {
		if patternip.MatchString(address) {
			ack = append(ack, "_client_ip")
		} else {
			mygw.encoderesponseup(req, fmt.Sprintf("_client_ip 格式错误: %s", address), 1, ws)
			return errors.New("wrong client ip format")
		}
	}
	netmask, _ := data.Get("_client_netmask").String()
	if netmask != "" {
		if patternip.MatchString(netmask) {
			ack = append(ack, "_client_netmask")
		} else {
			mygw.encoderesponseup(req, fmt.Sprintf("_client_netmask 格式错误: %s", netmask), 1, ws)
			return errors.New("wrong client netmask format")
		}
	}
	gateway, _ := data.Get("_client_gateway").String()
	if gateway != "" {
		if patternip.MatchString(gateway) {
			ack = append(ack, "_client_gateway")
		} else {
			mygw.encoderesponseup(req, fmt.Sprintf("_client_gateway 格式错误: %s", gateway), 1, ws)
			return errors.New("wrong client gateway format")
		}
	}
	if inet == "static" {
		if address == "" || netmask == "" || gateway == "" {
			mygw.encoderesponseup(req, fmt.Sprintf("设定静态ip地址时,address,netmask,gateway 必须设定值"), 1, ws)
			return errors.New("address or netmask or gateway value is empty")
		}
	}
	ipconstr := "auto lo\n" +
		"iface lo inet loopback\n" +
		"auto eth0\n" +
		"allow-hotplug eth0\n" +
		fmt.Sprintf("iface eth0 inet %s\n", inet)
	if inet == "static" {
		ipconstr += fmt.Sprintf("address %s\n", address) +
			fmt.Sprintf("netmask %s\n", netmask) +
			fmt.Sprintf("gateway %s\n", gateway)
	}
	ipconstr += "allow-hotplug wlan0\n" +
		"auto wlan0\n" +
		"iface wlan0 inet static\n" +
		"address 192.168.8.1\n" +
		"netmask 255.255.255.0"

	if _, err := os.Stat(common.INTERFACES); err != nil {
		if os.IsNotExist(err) {
			f, _ := os.Create(common.INTERFACES)
			if _, err := f.WriteString(ipconstr); err != nil {
				log.Errorf("set ip config error :%s", err)
				mygw.encoderesponseup(req, fmt.Sprintf("设置失败,error: %s", err), 1, ws)
				return err
			}
			f.Sync()
			f.Close()
		}
	} else {
		os.Remove(common.INTERFACES)
		f, _ := os.OpenFile(common.INTERFACES, os.O_WRONLY|os.O_CREATE, 0666)
		if _, err := io.WriteString(f, ipconstr); err != nil {
			log.Errorf("set ip config error :%s", err)
			mygw.encoderesponseup(req, fmt.Sprintf("设置失败,error: %s", err), 1, ws)
			return err
		}
		f.Sync()
		f.Close()
	}
	//重启生效
	mygw.encoderesponseup(req, fmt.Sprintf("%s设置成功", ack), 0, ws)
	cmd := exec.Command("reboot")
	var out bytes.Buffer
	cmd.Stdin = os.Stdin
	cmd.Stdout = &out
	cmd.Run()
	return nil
}

func (mygw *Gateway) initget(req *simplejson.Json, ws *websocket.Conn) {
	var ipstr string
	conf, _ := config.LoadConfigFile(mygw.Conpath)
	//	defer config.SaveConfigFile(conf, mygw.Conpath)
	if ifaddrs, err := net.InterfaceAddrs(); err == nil {
		for _, address := range ifaddrs {
			if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					if ipnet.IP.String() != "192.168.8.1" {
						ipstr = ipnet.IP.String()
					}
				}
			}
		}
	}
	var upm map[string]string
	if cm, err := conf.GetSection("mqtt"); err == nil {
		upm = map[string]string{
			"_client_ip": ipstr,
		}
		if common.Mqttconnected {
			cm["net_status"] = "online"
		} else {
			cm["net_status"] = "offline"
		}
		upm = common.Mergemap(cm, upm)
	} else {
		mygw.encoderesponseup(req, fmt.Sprintf("获取配置参数失败 : %s", err), 1, ws)
	}
	mygw.encoderesponseup(req, upm, 0, ws)
}

func (mygw *Gateway) managerUpdatecommif(req *simplejson.Json, ws *websocket.Conn) {
	var ack []string
	conf, _ := config.LoadConfigFile(mygw.Conpath)
	data := req.Get("data")
	if dm, ok := data.Map(); ok == nil {
		for k, v := range dm {
			if vstr, ok := v.(string); ok {
				conf.SetValue("commif", k, vstr)
			}
			ack = append(ack, k)
		}
		config.SaveConfigFile(conf, mygw.Conpath)
		mygw.updateRstat()
	} else {
		mygw.encoderesponseup(req, fmt.Sprintf("設置通信接口失败"), 1, ws)
	}
	mygw.encoderesponseup(req, fmt.Sprintf("通訊接口%s设置成功", ack), 0, ws)
}

func (mygw *Gateway) managerListcommif(req *simplejson.Json, ws *websocket.Conn) {
	conf, _ := config.LoadConfigFile(mygw.Conpath)
	if cm, err := conf.GetSection("commif"); err == nil {
		mygw.encoderesponseup(req, cm, 0, ws)
	} else {
		mygw.encoderesponseup(req, fmt.Sprintf("获取通信接口参数失败 : %s", err), 1, ws)
	}
}

func (mygw *Gateway) managerDevSuppotlist(req *simplejson.Json, ws *websocket.Conn) {
	var keys []string
	for k := range device.RegDevice {
		keys = append(keys, k)
	}
	mygw.encoderesponseup(req, keys, 0, ws)
}

func (mygw *Gateway) managerDevList(req *simplejson.Json, ws *websocket.Conn) {
	var keys []map[string]interface{}
	for _, v := range mygw.DevIfMap {
		velement, _ := v.GetElement()
		keys = append(keys, velement)
	}
	mygw.encoderesponseup(req, keys, 0, ws)
}

func (mygw *Gateway) managerDevUpdate(req *simplejson.Json, ws *websocket.Conn) {
	conf, _ := config.LoadConfigFile(mygw.Devpath)
	datalist := []interface{}{}
	data := req.Get("data").Interface()
	if dl, ok := data.([]interface{}); ok {
		datalist = append(datalist, dl...)
	} else {
		datalist = append(datalist, data)
	}
	for _, data := range datalist {
		func(data interface{}) {
			var ackerr string
			//		dm, err := data.Map()
			dm, _ := data.(map[string]interface{})
			dtype, okDType := dm["_type"].(string)
			if !okDType {
				ackerr = "必须要有_type参数"
			}
			_, exDType := device.RegDevice[dtype]
			if okDType && !exDType {
				ackerr = fmt.Sprintf("不支持的设备类型 :%s", dtype)
			}
			devid, okDEvid := dm["_devid"].(string)
			if okDType && exDType && !okDEvid {
				ackerr = "必须要有_devid参数"
			}
			conn, okCOnn := dm["_conn"].(map[string]interface{})
			if okDType && exDType && okDEvid && !okCOnn {
				ackerr = "必须要有_conn参数"
			}
			commif, okCOmmif := conn["commif"].(string)
			if okDType && exDType && okDEvid && okCOnn && !okCOmmif {
				ackerr = "必须要有commif参数"
			}
			_, exCOmmif := mygw.ConMap[commif]
			if okDType && exDType && okDEvid && okCOnn && okCOmmif && !exCOmmif {
				//ackerr = fmt.Sprintf("不存在或未配置的通信接口 :%s", commif)
				exCOmmif = true
			}
			_, okDEvaddr := conn["devaddr"]
			if okDEvaddr {
				addrtype := reflect.TypeOf(conn["devaddr"])
				okDEvaddr = addrtype.Name() == "Number" || addrtype.Name() == "string"
			}
			if okDType && exDType && okDEvid && okCOnn && okCOmmif && exCOmmif && !okDEvaddr {
				ackerr = "必须要有devaddr参数"
			}
			check := false
			var err error
			if okDType && exDType && okDEvid && okCOnn && okCOmmif && exCOmmif && okDEvaddr {
				check, err = device.RegDevice[dtype].CheckKey(conn)
				if okDType && exDType && okDEvid && okCOnn && okCOmmif && exCOmmif && okDEvaddr && !check {
					ackerr = err.Error()
				}
			}

			if okDType && exDType && okDEvid && okCOnn && okCOmmif && exCOmmif && okDEvaddr && check {
				conf.SetValue(devid, "_type", dtype)
				//				conf.SetValue(devid, "commif", commif)
				//				conf.SetValue(devid, "devaddr", devaddr)
				for k, v := range conn {
					if val, ok := v.(string); ok {
						conf.SetValue(devid, k, val)
					} else if val, ok := v.(json.Number); ok {
						conf.SetValue(devid, k, val.String())
					}
				}
				config.SaveConfigFile(conf, mygw.Devpath)
				mygw.updateRstat()
				log.Infof("设备 : %s更新成功", devid)
				mygw.encoderesponseup(req, fmt.Sprintf("设备 : %s更新成功", devid), 0, ws)
			} else {
				mygw.encoderesponseup(req, ackerr, 1, ws)
				log.Errorf("设备 : %s更新失败 Error : %s", devid, ackerr)
			}
		}(data)
	}
}

func (mygw *Gateway) managerDevDelete(req *simplejson.Json, ws *websocket.Conn) {
	conf, _ := config.LoadConfigFile(mygw.Devpath)
	datalist := []interface{}{}
	data := req.Get("data").Interface()
	if dl, ok := data.([]interface{}); ok {
		datalist = append(datalist, dl...)
	} else {
		datalist = append(datalist, data)
	}
	for _, data := range datalist {
		func(data interface{}) {
			var ackerr string
			dm, _ := data.(map[string]interface{})
			devid, okDEvid := dm["_devid"].(string)
			if !okDEvid {
				ackerr = "必须要有_devid参数"
			}
			okDEl := conf.DeleteSection(devid)
			if okDEvid && !okDEl {
				ackerr = fmt.Sprintf("设备 : %s删除失败或该设备不存在", devid)
			}
			if okDEl && okDEvid {
				config.SaveConfigFile(conf, mygw.Devpath)
				mygw.updateRstat()
				log.Infof("设备 : %s删除成功", devid)
				mygw.encoderesponseup(req, fmt.Sprintf("设备 : %s删除成功", devid), 0, ws)
			} else {
				mygw.encoderesponseup(req, ackerr, 1, ws)
			}
		}(data)
	}
}

func (mygw *Gateway) managerSetSystemTIme(req *simplejson.Json, ws *websocket.Conn) {
	data := req.Get("data")
	patterndate := regexp.MustCompile(`^\d{2}-\d{2}-\d{4}|\d{2}/\d{2}/\d{4}$`)
	patterntime := regexp.MustCompile(`^\d{2}:\d{2}:\d{2}$`)

	da, okDAte := data.Get("date").String()
	ti, okTIme := data.Get("time").String()
	if okDAte == nil && okTIme == nil && patterndate.MatchString(da) && patterntime.MatchString(ti) {
		dt := fmt.Sprintf("%s %s", da, ti)
		//		fmt.Println(time.Unix(1512144000, 0).Format("01/02/2006 15:04:05.999999999"))
		cmd := exec.Command("date", "-s", dt)
		var outbuf bytes.Buffer
		cmd.Stdin = os.Stdin
		cmd.Stdout = &outbuf
		err := cmd.Run()
		if err == nil {
			mygw.encoderesponseup(req, fmt.Sprintf("系统时间设置成功"), 0, ws)
		} else {
			mygw.encoderesponseup(req, fmt.Sprintf("系统时间设置失败"), 1, ws)
		}
	} else {
		mygw.encoderesponseup(req, fmt.Sprintf("Error:请求格式错误"), 1, ws)
	}
}

func (mygw *Gateway) managerSetInterval(req *simplejson.Json, ws *websocket.Conn) {
	data := req.Get("data")
	if interval, ok := data.Get("_interval").Int(); ok == nil {
		conf, _ := config.LoadConfigFile(mygw.Conpath)
		conf.SetValue("mqtt", "_interval", strconv.Itoa(interval))
		config.SaveConfigFile(conf, mygw.Conpath)
		mygw.Update()
		mygw.encoderesponseup(req, fmt.Sprintf("_interval设置成功"), 0, ws)
	} else {
		mygw.encoderesponseup(req, fmt.Sprintf("Error:请求格式错误"), 1, ws)
	}
}

func (mygw *Gateway) managerUpdateDrive(req *simplejson.Json, ws *websocket.Conn) {
	//	cmder := Cmdfp{
	//		Cmdfunc: func(request *simplejson.Json) error {
	//			pjs, _ := request.EncodePretty()
	//			fmt.Printf("%s\n", pjs)
	//			return nil
	//		},
	//		Param: req,
	//	}
	//	//	mygw.Cmdlist.PushBack(cmder)
	//	mygw.Cmdchan <- cmder
}

func (mygw *Gateway) getSetBase(req *simplejson.Json, rw string, ws *websocket.Conn) {
	cmder := Cmdfp{
		Cmdfunc: func(request *simplejson.Json) error {
			data := request.Get("data")
			daif, err := data.Array()
			if err == nil {
				for _, dam := range daif {
					das, ok := dam.(map[string]interface{})
					if ok {
						id, ok := das["_devid"].(string)
						if ok {
							if mf, ok := mygw.DevIfMap[id]; ok {
								if vals, err := mf.RWDevValue(rw, das); err == nil {
									mygw.encoderesponseup(request, vals, 0, ws)
								} else {
									mygw.encoderesponseup(request, dict{"_devid": id, "error": err.Error()}, 1, ws)
								}
							} else {
								mygw.encoderesponseup(request, fmt.Sprintf("%s不存在", id), 1, ws)
							}
						} else {
							mygw.encoderesponseup(request, fmt.Sprintf("Error:请求格式错误,缺少必要字段"), 1, ws)
						}
					}
				}
				return nil
			}
			id, err := data.Get("_devid").String()
			datam, _ := data.Map()
			if err == nil {
				if mf, ok := mygw.DevIfMap[id]; ok {
					if vals, err := mf.RWDevValue(rw, datam); err == nil {
						mygw.encoderesponseup(request, vals, 0, ws)
					} else {
						mygw.encoderesponseup(request, dict{"_devid": id, "error": err.Error()}, 1, ws)
					}
				} else {
					mygw.encoderesponseup(request, fmt.Sprintf("%s不存在", id), 1, ws)
				}
			} else {
				mygw.encoderesponseup(request, fmt.Sprintf("Error:请求格式错误,缺少必要字段"), 1, ws)
			}
			return nil
		},
		Param: req,
	}
	//	mygw.Cmdlist.PushBack(cmder)
	if ws == nil {
		mygw.Cmdchan <- cmder
	} else {
		cmder.Cmdfunc(cmder.Param)
	}
}
func (mygw *Gateway) doGetvar(req *simplejson.Json, ws *websocket.Conn) {
	mygw.getSetBase(req, "r", ws)
}

func (mygw *Gateway) doSetvar(req *simplejson.Json, ws *websocket.Conn) {
	mygw.getSetBase(req, "w", ws)
}

func (mygw *Gateway) doHelp(req *simplejson.Json, ws *websocket.Conn) {
	data, err := req.Get("data").String()
	if err == nil {
		var helpdoc interface{}
		ifhelp, ok := device.RegDevice[data]
		if ok {
			helpdoc = ifhelp.HelpDoc()
		} else {
			helpdoc = fmt.Sprintf("device %s not support", data)
		}
		if ws == nil {
			mygw.Handler.SendDataUp(helpdoc)
		} else {
			if msg, err := json.Marshal(helpdoc); err == nil {
				ws.Write(msg)
			}
		}
		return
	}
	if ws == nil {
		mygw.Handler.SendDataUp(gatewayHelp())
	} else {
		if msg, err := json.Marshal(gatewayHelp()); err == nil {
			ws.Write(msg)
		}
	}
}

// EncodeAutoup ..
func (mygw *Gateway) EncodeAutoup(data map[string]interface{}) error {
	if isup, ok := data["_update"]; ok {
		if bisup, ok := isup.(bool); ok {
			if !bisup {
				return nil
			}
		}
	}
	var errstat error
	uj := simplejson.New()
	devid, _ := mygw.ConMap["_client_id"]
	runstate, _ := mygw.ConMap["runstate"]
	from := map[string]string{
		"_devid":    devid,
		"_model":    common.MODEL,
		"_version":  common.VERSION,
		"_runstate": runstate,
	}
	header := map[string]interface{}{
		"from":    from,
		"msgtype": "update",
	}
	status := 0
	if data["error"] != nil {
		status = 1
	}
	request := map[string]interface{}{
		"cmd":        "do/auto_up_data",
		"data":       data,
		"statuscode": status,
		"timestamp":  time.Now().Unix(),
	}
	uj.Set("header", header)
	uj.Set("request", request)
	if common.Mqttconnected {
		errstat = mygw.Handler.SendDataUp(uj)
	}
	//	if wsuj, ok := uj.MarshalJSON(); ok == nil {
	if wsuj, ok := json.Marshal(uj); ok == nil {
		for _, ws := range mygw.WsMap {
			//			websocket.Message.Send(ws, wsuj)
			_, errstat = ws.Write(wsuj)
		}
	}
	return errstat
}

// OnOfflineMsg ..
func (mygw *Gateway) OnOfflineMsg(da uint) string {
	uj := simplejson.New()
	devid, _ := mygw.ConMap["_client_id"]
	runstate, _ := mygw.ConMap["runstate"]
	from := map[string]string{
		"_devid":    devid,
		"_model":    common.MODEL,
		"_version":  common.VERSION,
		"_runstate": runstate,
	}
	header := map[string]interface{}{
		"from":    from,
		"msgtype": "update",
	}
	req := map[string]interface{}{
		"cmd":        "push/state.do",
		"data":       da,
		"statuscode": 0,
		"timestamp":  time.Now().Unix(),
	}
	uj.Set("header", header)
	uj.Set("request", req)
	retuj, err := json.Marshal(uj)
	if err == nil {
		return string(retuj)
	}
	return err.Error()
}

func (mygw *Gateway) encoderesponseup(req *simplejson.Json, data interface{}, status int, ws *websocket.Conn) error {
	var errstat error
	uj := simplejson.New()
	devid, _ := mygw.ConMap["_client_id"]
	runstate, _ := mygw.ConMap["runstate"]
	from := map[string]string{
		"_devid":    devid,
		"_model":    common.MODEL,
		"_version":  common.VERSION,
		"_runstate": runstate,
	}
	header := map[string]interface{}{
		"from":    from,
		"msgtype": "response",
	}
	ret := req.Get("return")
	retarray, err := ret.StringArray()
	if err == nil {
		request := map[string]interface{}{}
		for _, retv := range retarray {
			request[retv] = req.Get(retv)
		}
		uj.Set("request", request)
	}
	response := map[string]interface{}{
		"cmd":        req.Get("cmd"),
		"data":       data,
		"statuscode": status,
		"timestamp":  time.Now().Unix(),
	}
	uj.Set("header", header)
	uj.Set("response", response)
	if ws == nil {
		errstat = mygw.Handler.SendDataUp(uj)
	} else {
		//		if msg, err := uj.MarshalJSON(); err == nil {
		if msg, err := json.Marshal(uj); err == nil {
			_, errstat = ws.Write(msg)
			log.Infoln("websocket message send = %v", uj)
		}
	}
	return errstat
}

func (mygw *Gateway) remoteSerial(req *simplejson.Json) (err error) {
	defer func() {
		if driveErr := recover(); driveErr != nil {
			log.Errorf("drive programer  error : (%s)", driveErr)
			errstr := fmt.Sprintf("drive programer  error : %s", driveErr)
			err = errors.New(errstr)
		}
	}()
	parsej := req.Get("parse")
	parse, _ := parsej.String()
	data := req.Get("data")
	switch parse {
	case "openser":
		mygw.serincount++
		if mygw.loop == true {
			mygw.Handler.SendSerDataUp([]byte("serial aleady in useing"))
			return nil
		}
		if ok := device.Openser(data); ok == nil {
			mygw.Handler.SendSerDataUp([]byte("open serial successful"))
			mygw.loop = true
		} else {
			mygw.Handler.SendSerDataUp([]byte("open serial failed"))
		}
		if mygw.loop == true && mygw.serincount == 1 {
			for {
				if mygw.loop == true {
					time.Sleep(time.Second)
					mygw.delay++
					if mygw.delay > 300 {
						device.Closeser()
						mygw.loop = false
						mygw.serincount = 0
					}
				} else {
					break
				}
			}
		}
	case "closeser":
		if mygw.serincount > 0 {
			mygw.serincount--
		}
		if mygw.serincount == 0 && mygw.loop == true {
			device.Closeser()
			mygw.loop = false
		}
		log.Info("serincount=", mygw.serincount)
		log.Info("mygw.loop=", mygw.loop)
	case "wser":
		mygw.delay = 0
		if da, ok := data.Interface().(string); ok {
			if deb64, err := base64.StdEncoding.DecodeString(da); err == nil {
				log.Info(deb64)
				err = device.Wser(deb64)
			}
			if err == nil {
				if res, err := device.Rser(); res != nil && err == nil {
					mygw.Handler.SendSerDataUp(res)
				}
			}
		}
	default:
		err = errors.New("available parser")
	}
	return err
}

func gatewayHelp() interface{} {
	data := dict{
		"parameter-1": "參數1",
		"parameter-2": "參數2",
	}
	request := dict{
		"cmd":  "请求命令(见命令列表)",
		"data": data,
	}
	reqformat := dict{
		"request": request,
	}
	cmdlist := dict{
		"init/set.do":                "初始化设置,需要data字段",
		"init/get.do":                "初始化信息获取",
		"manager/get_suppot_devlist": "获取网关所支持的设备",
		"manager/dev/update.do":      "更新设备,需要data字段",
		"manager/dev/delete.do":      "删除设备,需要data字段",
		"manager/update_commif.do":   "设置网关的通信接口,需要data字段,注:出厂前已设定,一般无需设置,供内部调试使用",
		"manager/list_commif.do":     "获取网关的通信接口信息",
		"manager/dev/list.do":        "获取当前设备列表",
		"manager/set_system_time":    "网关校时,需要data字段",
		"manager/set_interval.do":    "设置自动读取设备的间隔时间,单位:second,值=0为不自动循环读取,需要data字段",
		"do/getvar":                  "读取设备实时数据,是否需要data字段,详见设备的帮助信息",
		"do/setvar":                  "控制操作设备,需要data字段,详见设备的帮助信息",
		"help":                       "获取帮助信息,无data字段为网关帮助信息,data字段值为设备名,则为设备的帮助信息",
	}
	forInitset := dict{
		"_client_ip":      "ip地址",
		"_client_gateway": "网络的网关地址",
		"_client_netmask": "网络掩码",
		"_interface_inet": "static或dhcp",
		"_server_ip":      "mqtt服务地址",
		"_server_name":    "mqtt接收topic,比如things",
		"_server_port":    "mqtt服务端口",
		"_username":       "mqtt登录用户名",
		"_password":       "mqtt登录密码",
	}
	forDevUpdate := dict{
		"_devid": "设备id",
		"_conn":  "设备的参数,各种设备不同,详见设备的帮助信息",
		"_type":  "设备类型",
	}
	forDevDelete := dict{
		"_devid": "设备id",
	}
	forUpdateCommif := dict{
		"rs485-xxx":     "通信端口值,比如:/dev/ttyS0",
		"rs232-xxx":     "通信端口值,比如:/dev/ttyS2",
		"interface-xxx": "网络通信接口,比如:enp5s0,wlp4s0",
	}
	forSetTime := dict{
		"date": "格式为:月/日/年,比如 09/02/2017",
		"time": "格式为:时:分:秒,比如 15:08:03",
	}
	forSetInterval := dict{
		"_interval": "值为uint类型数字",
	}
	forGetvar := dict{
		"_devid":   "设备id",
		"value...": "读取设备的参数,各种设备不同,详见设备的帮助信息",
	}

	forSetvar := dict{
		"_devid":   "设备id",
		"value...": "控制操作设备的参数,各种设备不同,详见设备的帮助信息",
	}
	dataexplan := dict{
		"1.data for (init/set.do)":              forInitset,
		"2.data for (manager/dev/update.do)":    forDevUpdate,
		"3.data for (manager/dev/delete.do)":    forDevDelete,
		"4.data for (manager/update_commif.do)": forUpdateCommif,
		"5.data for (manager/set_system_time)":  forSetTime,
		"6.data for (manager/set_interval.do)":  forSetInterval,
		"7,data for (do/getvar)":                forGetvar,
		"8,data for (do/setvar)":                forSetvar,
	}
	hj := simplejson.New()
	hj.Set("1.格式", reqformat)
	hj.Set("2.命令列表", cmdlist)
	hj.Set("3.data参数解释", dataexplan)
	return hj
}

//func yjhttp(w http.ResponseWriter, r *http.Request) {
//	fmt.Println("get", r.URL.Path, " from ", r.RemoteAddr)
//	t, err := template.ParseFiles("templates/index.html")
//	if err != nil {
//		log.Println(err)
//	}
//	t.Execute(w, nil)
//}
