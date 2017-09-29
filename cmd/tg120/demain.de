package main

import (
	"github.com/yjiong/go_tg120/internal/device"
)

func main() {
	r := device.RSBAS{}
	rsbas, _ := r.NewDev("rsbas-0", map[string]string{"devaddr": "2"})
	rsbas.RWDevValue("r", nil)
}
