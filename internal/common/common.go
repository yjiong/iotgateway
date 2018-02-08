package common

import (
	log "github.com/Sirupsen/logrus"
	"github.com/yjiong/iotgateway/config"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// versin
const (
	VERSION    = "1.0"
	MODEL      = "IOT-GATEWAY"
	INTERFACES = "/etc/network/interfaces"
)

// CONFILEPATH ..
var CONFILEPATH = "./config.ini"

// DEVFILEPATH ..
var DEVFILEPATH = "./devlist.ini"

// Mqttconnected ..
var Mqttconnected = false

func init() {
	var pathfs string
	if runtime.GOOS == "Linux" {
		pathfs = "\\"
	} else {
		pathfs = "/"
	}
	if execfile, err := exec.LookPath(os.Args[0]); err == nil {
		//		fmt.Printf("%s\n", execfile)
		if path, err := filepath.Abs(execfile); err == nil {
			//			fmt.Printf("%s\n", path)
			i := strings.LastIndex(path, pathfs)
			basepath := string(path[0 : i+1])
			//			fmt.Printf("%s\n", path[0:i+1])
			CONFILEPATH = basepath + CONFILEPATH[2:]
			DEVFILEPATH = basepath + DEVFILEPATH[2:]
		}
	}
}

// NewConMap ..
func NewConMap(confile string) (map[string]string, error) {
	_, err := os.Stat(confile)
	if os.IsNotExist(err) {
		return nil, nil
	}
	con, err := config.LoadConfigFile(confile)
	if err != nil {
		log.WithFields(log.Fields{
			"config": con,
		}).Errorf("load config file failed: %s", err)
		return nil, err
	}
	retm := make(map[string]string)
	for _, sec := range con.GetSectionList() {
		if m, err := con.GetSection(sec); err == nil {
			retm = Mergemap(retm, m)
		} else {
			log.Errorf("get config element failed: %s", err)
			return nil, err
		}
	}
	return retm, nil
}

// Mergemap ..
func Mergemap(lm ...map[string]string) map[string]string {
	retmap := make(map[string]string)
	for _, m := range lm {
		for k, v := range m {
			retmap[k] = v
		}
	}
	return retmap
}
