package main

import (
	//"fmt"
	//"math"
	log "github.com/Sirupsen/logrus"
	"github.com/yjiong/go_tg120/internal/device"
	"time"
)

func main() {
	log.SetLevel(log.DebugLevel)
	//r := device.QDSL_SM510{}
	//r := device.TEST_GO{}
	//r := device.RSBAS{}
	r := device.FUJITSU{}
	//r := device.DTSD422{}
	tval, _ := r.NewDev("fujit", map[string]string{
		//"devaddr": "3300027014",
		"devaddr":  "1",
		"commif":   "rs485-1",
		"mtype":    "in_m",
		"sub_addr": "1",
		//"BaudRate": "19200",
		//"DataBits": "8",
		//"StopBits": "1",
	})
	elem := map[string]interface{}{
		"_varvalue": "1",
	}
	//tval.RWDevValue("w", nil)
	for {

		if _, err := tval.RWDevValue("r", elem); err != nil {
			log.Debugf("error=%s", err)
		} else {
			log.Debugf("ok!!!")
		}
		time.Sleep(2 * time.Second)
		//break
	}
}
