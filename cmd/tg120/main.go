package main

import (
	"context"
	"fmt"
//	"io/ioutil"
//	"net"
//	"net/http"
	"os"
	"os/signal"
//	"strings"
	"syscall"
//	"time"
	log "github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
	"google.golang.org/grpc/grpclog"
//	mqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/go_tg120/internal/handler"
)

func init() {
	grpclog.SetLogger(log.StandardLogger())
}

var version string = "v 1.0"// set by the compiler

func run(c *cli.Context) error {
	log.SetLevel(log.Level(uint8(c.Int("log-level"))))

	 ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	log.WithFields(log.Fields{
		"version": version,
		"docs":    "http://www.xingdong.org/",
	}).Info("starting mq programer")

	// get context
	pmh := mustGetContext(c)
	fmt.Printf("%T\n",pmh)
	go mqpbhandler(pmh,pmh.DataDownChan())
	
	sigChan := make(chan os.Signal)
	exitChan := make(chan struct{})
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	log.WithField("signal", <-sigChan).Info("signal received")
	go func() {
		log.Warning("stopping mq")
		// todo: handle graceful shutdown?
		exitChan <- struct{}{}
	}()
//	select {
//	case a := <-exitChan:
//		log.WithField("signal", a).Info("signal received, stopping immediately")
//	case s := <-sigChan:
//		log.WithField("signal", s).Info("signal received, stopping immediately")
//	}

	return nil
}

func mqpbhandler(pmh handler.Handler, dpc chan handler.DataDownPayload) {
	fmt.Println("test chan in mqrecev")
		for pl := range dpc {
		go func(pl handler.DataDownPayload) {
			
			fmt.Printf("%s\n",pl)

		}(pl)
	}
}

func mustGetContext(c *cli.Context) handler.Handler {
	// setup mqtt handler
	h, err := handler.NewMQTTHandler(c.String("mqtt-server"), c.String("mqtt-username"), c.String("mqtt-password"), c.String("mqtt-ca-cert"), c.String("tr_topic"))
	if err != nil {
		log.Fatalf("setup mqtt handler error: %s", err)
	}
	return  h
}

func main() {
	app := cli.NewApp()
	app.Name = "mq"
	app.Usage = "application for TG120 gateway"
	app.Version = version
	app.Copyright = "See http://github.com/yjiong/mq for copyright information"
	app.Action = run
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "mqtt-server",
			Usage:  "mqtt server (e.g. scheme://host:port where scheme is tcp, ssl or ws)",
			Value:  "tcp://192.168.1.160:1883",
			EnvVar: "MQTT_SERVER",
		},
		cli.StringFlag{
			Name:   "mqtt-username",
			Usage:  "mqtt server username (optional)",
			Value:  "jiangnan",
			EnvVar: "MQTT_USERNAME",
		},
		cli.StringFlag{
			Name:   "mqtt-password",
			Usage:  "mqtt server password (optional)",
			Value:  "iloveyou",
			EnvVar: "MQTT_PASSWORD",
		},
		cli.StringFlag{
			Name:   "mqtt-ca-cert",
			Usage:  "mqtt CA certificate file used by the gateway backend (optional)",
			EnvVar: "MQTT_CA_CERT",
		},
		cli.StringFlag{
			Name:   "ca-cert",
			Usage:  "ca certificate used by the api server (optional)",
			EnvVar: "CA_CERT",
		},
		cli.IntFlag{
			Name:   "log-level",
			Value:  4,
			Usage:  "debug=5, info=4, warning=3, error=2, fatal=1, panic=0",
			EnvVar: "LOG_LEVEL",
		},
		cli.BoolFlag{
			Name:   "disable-assign-existing-users",
			Usage:  "when set, existing users can't be re-assigned (to avoid exposure of all users to an organization admin)",
			EnvVar: "DISABLE_ASSIGN_EXISTING_USERS",
		},
		cli.StringFlag{
			Name:   "tr_topic",
			Value:  "device",
			Usage:	 "subscribe publish topic",
			EnvVar: "THINGS_TOPIC",	
		},
	}
	app.Run(os.Args)
}
