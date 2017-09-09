package common

import (
	log "github.com/Sirupsen/logrus"
	"github.com/yjiong/go_tg120/config"
	"os"
)

const (
	VERSION     = "1.0"
	MODEL       = "TG120"
	DEVFILEPATH = "./devlist.ini"
	CONFILEPATH = "./config.ini"
	INTERFACES  = "/etc/network/interfaces"
)

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

func Mergemap(lm ...map[string]string) map[string]string {
	retmap := make(map[string]string)
	for _, m := range lm {
		for k, v := range m {
			retmap[k] = v
		}
	}
	return retmap
}
