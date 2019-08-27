package common

import (
	log "github.com/sirupsen/logrus"
	"github.com/yjiong/iotgateway/config"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// INTERFACES ...
const (
	INTERFACES = "/etc/network/interfaces"
)

var (
	//VERSION ...
	VERSION string

	// MODEL ...
	MODEL = "IOT-GATEWAY"

	// BASEPATH ..
	BASEPATH string

	// Mqttconnected ..
	Mqttconnected = false

	// CONFILEPATH ..
	CONFILEPATH = "./config.ini"

	// DEVFILEPATH ..
	DEVFILEPATH = "./devlist"

	// SCHEDULEPATH ..
	SCHEDULEPATH = "./schedule"
)

func init() {
	var pathfs string
	if runtime.GOOS == "Linux" {
		pathfs = "\\"
	} else {
		pathfs = "/"
	}
	if execfile, err := exec.LookPath(os.Args[0]); err == nil {
		if path, err := filepath.Abs(execfile); err == nil {
			i := strings.LastIndex(path, pathfs)
			BASEPATH = string(path[0 : i+1])
			CONFILEPATH = BASEPATH + CONFILEPATH[2:]
			DEVFILEPATH = BASEPATH + DEVFILEPATH[2:]
			SCHEDULEPATH = BASEPATH + SCHEDULEPATH[2:]
		}
	}
	if _, err := os.Stat("/etc/default/iotdconf"); err == nil {
		CONFILEPATH = "/etc/default/iotdconf"
	}
	if _, err := os.Stat(DEVFILEPATH + `.ini`); err == nil {
		DEVFILEPATH = DEVFILEPATH + `.ini`
	}
	for _, fp := range []string{DEVFILEPATH, SCHEDULEPATH} {
		_, err := os.Stat(fp)
		if os.IsNotExist(err) {
			f, _ := os.Create(fp)
			//		f.WriteString("[xxxx]")
			f.Sync()
			f.Close()
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
	if model, ok := retm["model"]; ok {
		MODEL = model
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
