package main

import (
	"container/list"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	assetfs "github.com/elazarl/go-bindata-assetfs"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"github.com/urfave/negroni"
	//"github.com/urfave/negroni"
	"github.com/golang/net/websocket"
	pb "github.com/yjiong/iotgateway/api"
	"github.com/yjiong/iotgateway/internal/common"
	"github.com/yjiong/iotgateway/internal/devapi"
	"github.com/yjiong/iotgateway/internal/device"
	_ "github.com/yjiong/iotgateway/internal/device/ammeter"
	_ "github.com/yjiong/iotgateway/internal/device/sensorcontrol"
	_ "github.com/yjiong/iotgateway/internal/device/watermeter"
	gw "github.com/yjiong/iotgateway/internal/gateway"
	"github.com/yjiong/iotgateway/internal/handler"
	"github.com/yjiong/iotgateway/internal/storage"
	"github.com/yjiong/iotgateway/internal/templates"
	"github.com/yjiong/iotgateway/internal/upgrade"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"
)

func init() {
	//	if runtime.GOOS == "linux" {
	//
	//	}
	grpclog.SetLogger(log.StandardLogger())
	_, err := os.Stat(common.BASEPATH + "httpcert")
	if os.IsNotExist(err) {
		log.Infoln("assets httpcert dir !", upgrade.RestoreAssets(common.BASEPATH, "httpcert"))
		syscall.Sync()
	}
	common.VERSION = VERSION
	fmt.Printf("%s\n%s\n%s\n%s\n%s\n%s\n", "      _           _ ", "__  _(_)_ __   __| | ___  _ __   __ _ ", `\ \/ / | '_ \ / _' |/ _ \| '_ \ / _' |`,
		" >  <| | | | | (_| | (_) | | | | (_| |", `/_/\_\_|_| |_|\__,_|\___/|_| |_|\__, |`, `                                |___/ `)
}

//VERSION ..
var VERSION string

func run(c *cli.Context) error {
	log.SetLevel(log.Level(uint8(c.Int("log-level"))))
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	log.WithFields(log.Fields{
		"version": common.VERSION,
		"docs":    "https://github.com/yjiong/iotgateway",
	}).Info("starting iot gateway programer")
	// 初始化
	gateway := mustGetGateway(c)
	//////////////////////////////////////////////////////////////////////
	//go device.ListenAndServe("udp", "127.0.0.1:2005", &gateway)
	//////////////////////////////////////////////////////////////////////

	gateway.UpdateSchedule()
	go func() {
		router := mux.NewRouter()
		jsonHandler, err := getJSONGateway(ctx, c)
		if err != nil {
			log.Fatal("get jsonhandler failed:", err, jsonHandler)
		}
		router.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
			data, err := templates.Asset("swagger/index.html")
			if err != nil {
				log.Errorf("get swagger template error: %s", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Write(data)
		}).Methods("get")
		router.PathPrefix("/api").Handler(jsonHandler)
		router.PathPrefix("/login").HandlerFunc(gateway.LoginHandler)
		router.PathPrefix("/message").Handler(websocket.Handler(gateway.WsHandle))

		router.PathPrefix("/").Handler(http.FileServer(&assetfs.AssetFS{
			Asset:     templates.Asset,
			AssetDir:  templates.AssetDir,
			AssetInfo: templates.AssetInfo,
			Prefix:    "",
		}))
		port := strings.SplitN(c.String("http-bind"), ":", 2)[1]
		if len(port) == 0 {
			log.Fatal("get port from bind failed")
		}
		grpcHandler, err := getGrpcServer(ctx, &gateway)
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			//log.Infoln(r)
			if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
				grpcHandler.ServeHTTP(w, r)
			} else {
				if router == nil {
					w.WriteHeader(http.StatusNotImplemented)
					return
				}
				router.ServeHTTP(w, r)
			}
		})
		myneg := negroni.New(
			negroni.HandlerFunc(gw.ValidateTokerMiddleware),
			negroni.HandlerFunc(gw.HandlerFuncGetFile),
			negroni.Wrap(handler),
		)
		go http.ListenAndServe(":80", http.HandlerFunc(redirect))
		if err := http.ListenAndServeTLS(":"+port,
			common.BASEPATH+"httpcert/server.crt",
			common.BASEPATH+"httpcert/server.key",
			myneg); err != nil {
			log.Fatal("ListenAndServe:", err)
		}
	}()

	//websocket 消息处理
	go func() {
		for wscmd := range gateway.WsNochanr {
			for k, v := range wscmd {
				go gateway.Wscmdhandler(v, gateway.WsMap[k])
			}
		}
	}()

	go func() {
		for cmd := range gateway.Cmdchan {
			go cmd.Cmdfunc(cmd.Param)
		}
	}()

	if commif, err := gateway.Getcommif(); err == nil {
		commif["ip4"] = `^(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])\.(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])\.(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])\.(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])`
		for cif := range commif {
			patternip := regexp.MustCompile(commif[cif])
			go func(cif string) {
				for {
					delay := 0
					for {
						pubInterval, err := strconv.Atoi(gateway.ConMap[gw.SysInterval])
						if err != nil {
							pubInterval = c.Int("interval")
						}
						delay++
						if delay > pubInterval {
							break
						}
						time.Sleep(time.Second * 1)
					}
					//starttime := time.Now().Unix()
					if gateway.ConMap[gw.SysInterval] != "0" {
						count := 64
						for id, dev := range gateway.DevIfMap {
							if cif != dev.GetCommif() && !patternip.MatchString(dev.GetCommif()) {
								continue
							}
							ret, err := dev.RWDevValue("r", nil)
							if count <= 0 {
								err = errors.New("device offline")
							}
							if err != nil {
								ret = map[string]interface{}{
									"_devid": id,
									"error":  err.Error(),
								}
							}
							go gateway.DB.InsertDevJdoc(id, "do/auto_up_data", ret)
							if err := gateway.EncodeAutoup(ret); err != nil {
								log.Errorf("auto updata error : %s", err)
							}
							count--
							//time.Sleep(time.Second)
						}
					} else {
						time.Sleep(time.Second)
					}
					//log.Debugln("一个周期=", time.Now().Unix()-starttime)
					//break
				}
			}(cif)
		}
	} else {
		log.Fatal("get commif failed", err)
	}

	mqttconnect(c, &gateway)
	go gateway.ScheduleLoop()
	go gateway.LostConnectRestart()
	go gateway.AutoDelOverdueHistoryData()
	//接受消息命令并执行
	go gateway.Mqttcmdhandler(gateway.Handler.DataDownChan())

	sigChan := make(chan os.Signal)
	exitChan := make(chan struct{})
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	log.WithField("signal", <-sigChan).Info("signal received")
	go func() {
		log.Warning("stopping gateway")
		exitChan <- struct{}{}
	}()
	select {
	case a := <-exitChan:
		log.WithField("signal", a).Info("exit signal received, stopping immediately")
	case s := <-sigChan:
		log.WithField("signal", s).Info("singnal signal received, stopping immediately")
	}

	return nil
}

func getGrpcServer(ctx context.Context, gateway *gw.Gateway) (*grpc.Server, error) {
	var validator gw.Validator
	grpcHandler := grpc.NewServer()
	pb.RegisterGatewayServiceServer(grpcHandler, gw.NewGatewayapi(validator, gateway))
	pb.RegisterDLT645_2007Server(grpcHandler, devapi.NewDtl645_2007api(gateway))
	pb.RegisterModbusRtuServer(grpcHandler, devapi.NewModbusRtuapi(gateway))
	pb.RegisterModbusTcpServer(grpcHandler, devapi.NewModbusTcpapi(gateway))
	pb.RegisterElectricMeterServer(grpcHandler, devapi.NewElectricMeterapi(gateway))
	pb.RegisterTC100R8Server(grpcHandler, devapi.NewTcozzreapi(gateway))
	pb.RegisterWaterMeterServer(grpcHandler, devapi.NewWaterMeterapi(gateway))

	return grpcHandler, nil
}

func getJSONGateway(ctx context.Context, c *cli.Context) (http.Handler, error) {
	b, err := upgrade.Asset("httpcert/server.crt")
	if err != nil {
		return nil, errors.Wrap(err, "read http-tls-cert cert error")
	}
	cp := x509.NewCertPool()
	if !cp.AppendCertsFromPEM(b) {
		return nil, errors.Wrap(err, "failed to append certificate")
	}
	grpcDialOpts := []grpc.DialOption{grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
		InsecureSkipVerify: true,
		RootCAs:            cp,
	}))}

	bindParts := strings.SplitN(c.String("http-bind"), ":", 2)
	if len(bindParts) != 2 {
		log.Fatal("get port from bind failed")
	}
	apiEndpoint := fmt.Sprintf("127.0.0.1:%s", bindParts[1])

	mux := runtime.NewServeMux(runtime.WithMarshalerOption(
		runtime.MIMEWildcard,
		&runtime.JSONPb{
			EnumsAsInts:  false,
			EmitDefaults: true,
		},
	))

	if err := pb.RegisterGatewayServiceHandlerFromEndpoint(ctx, mux, apiEndpoint, grpcDialOpts); err != nil {
		return nil, errors.Wrap(err, "register gateway manager handler error")
	}
	if err := pb.RegisterDLT645_2007HandlerFromEndpoint(ctx, mux, apiEndpoint, grpcDialOpts); err != nil {
		return nil, errors.Wrap(err, "register PMC-xxx handler error")
	}
	if err := pb.RegisterModbusRtuHandlerFromEndpoint(ctx, mux, apiEndpoint, grpcDialOpts); err != nil {
		return nil, errors.Wrap(err, "register ModbusRtu handler error")
	}
	if err := pb.RegisterModbusTcpHandlerFromEndpoint(ctx, mux, apiEndpoint, grpcDialOpts); err != nil {
		return nil, errors.Wrap(err, "register ModbusTcp handler error")
	}
	if err := pb.RegisterElectricMeterHandlerFromEndpoint(ctx, mux, apiEndpoint, grpcDialOpts); err != nil {
		return nil, errors.Wrap(err, "register ElectricMeter handler error")
	}
	if err := pb.RegisterTC100R8HandlerFromEndpoint(ctx, mux, apiEndpoint, grpcDialOpts); err != nil {
		return nil, errors.Wrap(err, "register TC100R8 handler error")
	}
	if err := pb.RegisterWaterMeterHandlerFromEndpoint(ctx, mux, apiEndpoint, grpcDialOpts); err != nil {
		return nil, errors.Wrap(err, "register WaterMeter handler error")
	}
	return mux, nil
}

func mustGetGateway(c *cli.Context) gw.Gateway {
	// 初始化设备,也就是设备的驱动接口
	devm, err := device.NewDevHandler(common.DEVFILEPATH)
	if err != nil {
		log.Fatalf("setup device interface map error: %s", err)
	}
	// 初始化配置文件
	conm, err := common.NewConMap(common.CONFILEPATH)
	if err != nil {
		log.Errorf("setup config parameter map error: %s", err)
	}

	if pgdns, ok := conm["postgresql_dns"]; !ok || pgdns == "" {
		conm["postgresql_dns"] = c.String("postgresql_dns")
	}
	gwdb := setPostgreSQLConnection(conm["postgresql_dns"])
	//*************************************************************
	//gwdb = nil
	if gwdb != nil {
		log.Infof("connecting database %s ok !", conm["postgresql_dns"])
		if err := gwdb.CreateDevTable("cmdhistory"); err != nil {
			log.Error(err)
		}
		for devid := range devm {
			if err := gwdb.CreateDevTable(devid); err != nil {
				log.Error(err)
			}
		}
	}
	return gw.Gateway{
		DevIfMap: devm,
		ConMap:   conm,
		//		Handler:   h,
		WsMap:     make(map[int]*websocket.Conn),
		Cmdlist:   list.New(),
		Devpath:   common.DEVFILEPATH,
		Conpath:   common.CONFILEPATH,
		Schedule:  common.SCHEDULEPATH,
		Cmdchan:   make(chan gw.Cmdfp),
		WsNochanr: make(chan map[int]string),
		DB:        gwdb,
	}
}

func setPostgreSQLConnection(pgdns string) *storage.MYDB {
	log.Infof("connecting to  %s ", pgdns)
	db, err := storage.OpenDatabase(pgdns)
	if err != nil {
		log.Error(errors.Wrap(err, "open database or ping error"))
		return nil
	}
	return db
}

func mqttconnect(c *cli.Context, gateway *gw.Gateway) {
	// 初始化mqtt接口
	// 初始化配置文件
	conm, err := common.NewConMap(common.CONFILEPATH)
	var h handler.Handler
	willmsg := gateway.OnOfflineMsg(0)
	onlinemsg := gateway.OnOfflineMsg(1)
	cm := map[string]string{
		"serverIp":   c.String("mqtt-server"),
		"serverPort": "",
		"username":   c.String("mqtt-username"),
		"password":   c.String("mqtt-password"),
		"cafile":     c.String("mqtt-ca-cert"),
		"clientId":   c.String("client_id"),
		"serverName": c.String("server_id"),
		"keepalive":  "60",
	}
	if conm != nil {
		cm = map[string]string{
			"serverIp":   conm[gw.MqttSvrIP],
			"serverPort": conm[gw.MqttSvrPort],
			"username":   conm[gw.MqttUser],
			"password":   conm[gw.MqttPasswd],
			"clientId":   conm[gw.ClientID],
			"serverName": conm[gw.MqttSvrName],
			"keepalive":  conm[gw.MqttKeepAlive],
			"cafile":     conm["cafile"],
			"certfile":   conm["certfile"],
			"keyfile":    conm["keyfile"],
		}
	}
	h, err = handler.NewMQTTHandler(cm, willmsg, onlinemsg)
	if err != nil {
		log.Fatalf("setup mqtt handler error: %s", err)
	}
	gateway.Handler = h
}

func main() {
	app := cli.NewApp()
	app.Name = "GATEWAY"
	app.Usage = "application for IOT gateway"
	app.Version = common.VERSION
	app.Author = "yaojiong"
	app.Email = "yjiong@msn.com"
	app.Copyright = "See https://github.com/yjiong/iotgateway for copyright information"
	app.Action = run
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "mqtt-server",
			Usage:  "mqtt server (e.g. scheme://host:port where scheme is tcp, ssl or ws)",
			Value:  "tcp://211.159.217.108:1883",
			EnvVar: "MQTT_SERVER",
		},
		cli.StringFlag{
			Name:   "mqtt-username",
			Usage:  "mqtt server username (optional)",
			Value:  "yj",
			EnvVar: "MQTT_USERNAME",
		},
		cli.StringFlag{
			Name:   "mqtt-password",
			Usage:  "mqtt server password (optional)",
			Value:  "yj12345",
			EnvVar: "MQTT_PASSWORD",
		},
		cli.StringFlag{
			Name:   "mqtt-ca-cert",
			Usage:  "mqtt CA certificate file used by the gateway backend (optional)",
			EnvVar: "MQTT_CA_CERT",
		},
		cli.IntFlag{
			Name:   "L, log-level",
			Value:  4,
			Usage:  "debug=5, info=4, warning=3, error=2, fatal=1, panic=0",
			EnvVar: "LOG_LEVEL",
		},
		cli.StringFlag{
			Name:   "client_id",
			Value:  "IotGW-GOLANG",
			Usage:  "subscribe publish topic client strings",
			EnvVar: "CLIENT_ID",
		},
		cli.StringFlag{
			Name:   "server_id",
			Value:  "iotserver",
			Usage:  "subscribe publish topic server strings",
			EnvVar: "SERVER_ID",
		},
		cli.IntFlag{
			Name:  "interval",
			Value: 300,
			Usage: "auto updata interval",
			EnvVar: "INTERVAL	",
		},
		cli.StringFlag{
			Name:   "postgresql_dns",
			Usage:  "postgres://user:password@hostname:port/database",
			Value:  `postgres://postgres:yj12345@localhost:5432/postgres?sslmode=disable`,
			EnvVar: "POSTGRESQL_DNS",
		},
		cli.StringFlag{
			Name:   "http-bind",
			Usage:  "ip:port to bind the (user facing) http server to (web-interface and REST / gRPC api)",
			Value:  "0.0.0.0:443",
			EnvVar: "HTTP_BIND",
		},
	}
	app.Run(os.Args)
}
