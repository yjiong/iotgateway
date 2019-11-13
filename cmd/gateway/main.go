//go:generate go-bindata -prefix ../../templates/ -pkg templates -o ../../internal/templates/templates_gen.go ../../templates/...
package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	assetfs "github.com/elazarl/go-bindata-assetfs"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"github.com/urfave/negroni"
	pb "github.com/yjiong/iotgateway/api"
	"github.com/yjiong/iotgateway/internal/common"
	"github.com/yjiong/iotgateway/internal/devapi"
	gw "github.com/yjiong/iotgateway/internal/gateway"
	"github.com/yjiong/iotgateway/internal/templates"
	"github.com/yjiong/iotgateway/internal/upgrade"
	"golang.org/x/net/websocket"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"
)

func init() {
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

	mqttconfig := map[string]string{
		gw.MqttSvrIP:     c.String("mqtt-server"),
		gw.MqttSvrPort:   c.String("mqtt-server-port"),
		gw.MqttUser:      c.String("mqtt-username"),
		gw.MqttPasswd:    c.String("mqtt-password"),
		gw.MqttCaFile:    c.String("mqtt-ca-cert"),
		gw.MqttCert:      c.String("mqtt-client-cert"),
		gw.MqttKey:       c.String("mqtt-client-key"),
		gw.ClientID:      c.String("client_id"),
		gw.MqttSvrName:   c.String("server_id"),
		gw.MqttKeepAlive: "60",
	}
	gateway := gw.NewGateway()
	//////////////////////////////////////////////////////////////////////
	//go device.ListenAndServe("udp", "127.0.0.1:2005", gateway)
	//////////////////////////////////////////////////////////////////////

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
		grpcHandler, err := getGrpcServer(gateway)
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	gateway.Server(ctx,
		gw.WithDevDriveAddrconfig(c.String("ddsvr-addr")),
		gw.WithDBconfig(c.String("postgresql_dns")),
		gw.WithReadDevInterval(c.String("interval")),
		gw.WithDevelopeFlag(c.Bool("develope")),
		gw.WithMqttConfig(mqttconfig))

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

func getGrpcServer(gateway *gw.Gateway) (*grpc.Server, error) {
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
	// dial options for the grpc-gateway
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
			Usage:  "mqtt server",
			Value:  "211.159.217.108",
			EnvVar: "MQTT_SERVER",
		},
		cli.StringFlag{
			Name:   "mqtt-server-port",
			Usage:  "mqtt server port",
			Value:  "1883",
			EnvVar: "MQTT_SVR_PORT",
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
			Name:   "mqtt-cafile",
			Usage:  "mqtt CA certificate file used by the gateway backend (optional)",
			EnvVar: "MQTT_CAFILE",
		},
		cli.StringFlag{
			Name:   "mqtt-client-cert",
			Usage:  "mqtt client cert file used by the gateway backend (optional)",
			EnvVar: "MQTT_CLINET_CERT",
		},
		cli.StringFlag{
			Name:   "mqtt-client-key",
			Usage:  "mqtt client key file used by the gateway backend (optional)",
			EnvVar: "MQTT_CLINET_KEY",
		},
		cli.IntFlag{
			Name:   "L, log-level",
			Value:  4,
			Usage:  "debug=5, info=4, warning=3, error=2, fatal=1, panic=0",
			EnvVar: "LOG_LEVEL",
		},
		cli.StringFlag{
			Name:   "client_id",
			Value:  "IotGW",
			Usage:  "subscribe publish topic client strings",
			EnvVar: "CLIENT_ID",
		},
		cli.StringFlag{
			Name:   "server_id",
			Value:  "IOTSERVER",
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
			Value:  "postgres://postgres:yj12345@localhost:5432/postgres?sslmode=disable",
			EnvVar: "POSTGRESQL_DNS",
		},
		cli.StringFlag{
			Name:   "http-bind",
			Usage:  "ip:port to bind the (user facing) http server to (web-interface and REST / gRPC api)",
			Value:  "0.0.0.0:443",
			EnvVar: "HTTP_BIND",
		},
		cli.StringFlag{
			Name:   "ddsvr-addr",
			Usage:  "ip:port  connect to device drive grpc server",
			Value:  "localhost:9973",
			EnvVar: "DEVADDR",
		},
		cli.BoolFlag{
			Name:   "D,develope",
			Usage:  "open restful web",
			EnvVar: "DECELOPE",
		},
	}
	app.Run(os.Args)
}
