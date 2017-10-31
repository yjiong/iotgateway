package main

import (
	//"fmt"
	//"math"
	//log "github.com/Sirupsen/logrus"
	"github.com/yjiong/go_tg120/internal/device"
)

func main() {
	//fmt.Printf("%x\n", math.Float32bits(100.0))

	//First it needs to be stated the bit-length of the input. Since the hex representation has 4 bytes (8 hex digits), it is most likely a float32 (needs clarification from the asker).
	//
	//You can parse the bytes from the hex representation into an uint32 using strconv.ParseUint(). ParseUint() always returns uint64 which uses 8 bytes in memory so you have to convert it to uint32 which uses 4 bytes just like float32:
	//
	//s := "C40C5253"
	//n, err := strconv.ParseUint(s, 16, 32)
	//if err != nil {
	//    panic(err)
	//}
	//n2 = uint32(n)
	//
	//Now you have the bytes but they are stored in a variable of type uint32 and therefore interpreted as the bytes of an integer. And you want to interpret them as the bytes of a IEEE-754 floating point number, you can use the unsafe package to do that:
	//
	//f := *(*float32)(unsafe.Pointer(&n2))
	//fmt.Println(f)
	//
	//Output (try it on the Go Playground):
	//
	//-561.2863
	//
	//Note:
	//
	//As JimB noted, for the 2nd part (translating uint32 to float32) the math package has a built-in function math.Float32frombits() which does exactly this under the hood:
	//
	//f := math.Float32frombits(n2)

	r := device.DTSD422{}
	tval, _ := r.NewDev("dev-0", map[string]string{
		"devaddr": "3300027014",
		"commif":  "rs485-1",
		//"BaudRate": "19200",
		//"DataBits": "8",
		//"StopBits": "1",
	})
	tval.RWDevValue("r", nil)

}
