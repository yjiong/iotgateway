package watermeter

import (
	"encoding/json"
	"sync"
	"testing"
	"time"

	simplejson "github.com/bitly/go-simplejson"
	log "github.com/sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
	. "github.com/yjiong/iotgateway/internal/device"
)

func TestDirve(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	Commif["rs485-1"] = "/dev/ttyUSB0"
	Mutex["rs485-1"] = new(sync.Mutex)
	r := LXSZ15F{}
	//r := LXZFL{}
	//r := SENSUS{}
	//r := ZENNER{}
	tval, _ := r.NewDev("TestDevice", map[string]string{
		device.DevAddr: "1100642246",
		//"fcode":   "06",
		//device.DevType: "LXSZ15F",
		"commif": "rs485-1",
		//"Parity": "E",
		//"BaudRate": "2400",
		//"BaudRate": "9600",
		//"DataBits": "8",
		//"StopBits": "1",
	})
	elem := map[string]interface{}{
		/***********伯克利dn15-20********************/
		//"版本号":  "",
		//"当前单价": "",
		//"生产日期": "",
		//"读表地址": "",
		//"写单价": map[string]interface{}{
		//"单价1":  json.Number("3.98"),
		//"单价2":  json.Number("4.98"),
		//"单价3":  json.Number("5.98"),
		//"用量1":  json.Number("9999"),
		//"用量2":  json.Number("999"),
		//"已用量1": json.Number("777"),
		//"已用量2": json.Number("666"),
		//},
		//"充退款": json.Number("222.1"),
		//"开阀": "",
		//"初始化": "",
		//"写反向累计流量": json.Number("5.88"),
		/***********赛达水表********************/
		//"开启预付费": "",
		//"关闭预付费": "",
		//"开阀": "",
		//"关阀": "",
		//"退款": json.Number("711.89"),
		//"充值": json.Number("11.89"),
	}
	//tval.RWDevValue("w", nil)
	Convey("==================测试驱动接口=====================\n", t, func() {
		for {
			if ret, err := tval.RWDevValue("r", elem); err != nil {
				So(err, ShouldBeNil)
				t.Errorf("error=%s,elem=%v", err, elem)
				t.Error(ret)
			} else {
				jret, _ := json.Marshal(ret)
				jsret, _ := simplejson.NewJson(jret)
				prettyret, _ := jsret.EncodePretty()
				log.Debugln(string(prettyret))
			}
			time.Sleep(0 * time.Second)
			break
		}
	})
}
