package sensorcontrol

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
	//r := RSBAS{}
	r := TUF2000SW{}
	//r := ZD6W1L{}
	tval, _ := r.NewDev("TestDevice", map[string]string{
		device.DevAddr: "1",
		"commif":  "rs485-1",
		//"IndoorNum": "8",
		//"mtype":     "inM",
		//"subAddr": "1",
		//"Parity":    "E",
		//"BaudRate":  "9600",
		//"DataBits":  "8",
		//"StopBits":  "1",
	})

	elem := map[string]interface{}{
		//"_varname":  "运行模式设置",
		//"_varvalue": "送风",
		//"_varname":  "运行开关设置",
		//"_varvalue": "运行",
		//"_varname":  "设置温度设定值",
		//"_varvalue": "28",
		//"_varname":  "气流设置",
		//"_varvalue": "中",
		//"_varname":  "垂直空气方向位置状态",
		//"_varvalue": "位置4",
		//"_varname":  "水平空气方向位置状态",
		//"_varvalue": "位置2",
		//"_varname":  "遥控器运行禁止设置",
		//"_varvalue": "允许",
		//"_varname":  "过滤网标志重置",
		//"_varvalue": "重置",
		//"_varname":  "经济运行模式设置",
		//"_varvalue": "节能运行",
		//"_varname":  "防冻液运行设置",
		//"_varvalue": "释放",
		//"_varname":  "制冷/干燥温度上限设置",
		//"_varvalue": "16",
		//"_varname":  "制冷/干燥温度下限设置",
		//"_varvalue": "16",
		//"_varname":  "加热温度上限设置",
		//"_varvalue": "16",
		//"_varname": "加热温度下限设置",
		//"_varvalue": "16",
		//"_varname":  "自动温度上限设置",
		//"_varvalue": "16",
		//"_varname":  "自动温度下限设置",
		//"_varvalue": "6",
		//"_varname":  "外部关热设置",
		//"_varvalue": "关热",
		//"_varname":  "紧急停止",
		//"_varvalue": "紧急停止请求",
		//"_varname":  "室外机低噪音运行设置",
		//"_varvalue": "性能优先3",
		//"_varname":  "室外机额定容量节省指令",
		//"_varvalue": "70%",
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
