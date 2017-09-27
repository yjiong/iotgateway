package main

import (
	"github.com/yjiong/go_tg120/internal/device"
)

func main() {
	rsbas := device.RSBAS{}
	rsbas.RWDevValue("r", nil)
}
