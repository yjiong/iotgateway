package ammeter

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
	r := DDZY422N{}
	//r := DTZY422{}
	//r := PMC340{}
	tval, _ := r.NewDev("TestDevice", map[string]string{
		device.DevAddr: "1100642246",
		device.DevType:   "DDZY422N",
		"commif":  "rs485-1",
		//"Parity": "E",
		//"BaudRate": "2400",
		//"BaudRate": "9600",
		//"DataBits": "8",
		//"StopBits": "1",
	})
	//initval := map[string]interface{}{
	//"报警金额":   json.Number("1088"),
	//"报警负荷":   json.Number("11.22"),
	//"允许透支金额": json.Number("1111"),
	//"允许囤积金额": json.Number("987654.32"),
	////"金额报警跳闸时间": json.Number("14"),
	//"尖单价": json.Number("4.98"),
	//"峰单价": json.Number("3.76"),
	//"平单价": json.Number("2.54"),
	//"谷单价": json.Number("1.32"),
	////"报警方式": 1,
	//}

	//initval := map[string]interface{}{
	//"报警电量":   json.Number("5"),
	//"报警负荷":   json.Number("11.22"),
	//"允许透支电量": json.Number("0"),
	//"允许囤积电量": json.Number("8888"),
	//}

	elem := map[string]interface{}{
		//"强制断闸": "",
		//"强制合闸": "",
		//"撤销强制": "",
		//"初始化": initval,
		//"充值": json.Number("1.66"),
		//"购电": json.Number("2.73"),
		//"退电": json.Number("4.67"),
		//"退款": json.Number("273.55"),
		//"校时": "2018-6-25 13:44:30",
		/******************PMC340*************/
		//"费率1电价": json.Number("01.05"),
		//"费率2电价":     json.Number("2.22"),
		//"费率3电价":     json.Number("3.33"),
		//"费率4电价":     json.Number("4.44"),
		//"报警金额1限值": json.Number("22"),
		//"报警金额2限值": json.Number("11"),
		//"透支金额限值": json.Number("20.50"),
		//"合闸允许金额限值": json.Number("0"),
		//"囤积金额限值": json.Number("888888.88"),
		//"退费": json.Number("70.01"),
		//"控制": "合闸",
		//"控制": "跳闸",
		//"控制": "报警",
		//"控制": "报警解除",
		//"控制": "保电解除",
		//"广播校时": "19-05-14 15:47:20",
		//"充值": json.Number("2.1"),
		//"all": "",
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
