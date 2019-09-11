package device

import (
	"encoding/json"
	"sync"
	"testing"
	"time"

	simplejson "github.com/bitly/go-simplejson"
	log "github.com/sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDirve(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	//Commif["rs485-1"] = "/dev/ttyUSB0"
	Commif["rs485-1"] = "/dev/pts/8"
	Mutex["rs485-1"] = new(sync.Mutex)
	r := &ModbusRtu{}
	tval, _ := r.NewDev("TestDevice", map[string]string{
		//device.DevType: "",
		device.DevAddr:         "1",
		"commif":          "rs485-1",
		"Parity":          "E",
		"BaudRate":        "9600",
		"DataBits":        "8",
		"StopBits":        "1",
		"StartingAddress": "0",
		"Quantity":        "10",
		"FunctionCode":    "3",
	})
	//initval := map[string]interface{}{
	//}

	elem := map[string]interface{}{}
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
