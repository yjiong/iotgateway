package handler

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"sync"
	"time"
	//	"bytes"
	//	"encoding/base64"
	//	"regexp"
	//	"strconv"
	simplejson "github.com/bitly/go-simplejson"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	log "github.com/sirupsen/logrus"
	"github.com/yjiong/iotgateway/internal/common"
)

// DataDownPayload ...
type DataDownPayload struct {
	Pj *simplejson.Json
}

// MQTTHandler ...
type MQTTHandler struct {
	conn            mqtt.Client
	dataDownChan    chan DataDownPayload
	wg              sync.WaitGroup
	ClientID        string
	ServerID        string
	onlinemsg       string
	LostConnectTime time.Time
	Firmware        chan []byte
}

// NewMQTTHandler creates a new MQTTHandler.
func NewMQTTHandler(conm map[string]string, willmsg, onlinemsg string) (Handler, error) {
	h := MQTTHandler{
		dataDownChan: make(chan DataDownPayload),
		Firmware:     make(chan []byte),
	}

	opts := mqtt.NewClientOptions()
	//opts.AddBroker(server)
	server := conm["_server_ip"] + ":" + conm["_server_port"]
	opts.SetUsername(conm["_username"])
	opts.SetPassword(conm["_password"])
	opts.SetOnConnectHandler(h.onConnected)
	opts.SetConnectionLostHandler(h.onConnectionLost)
	kplv, _ := strconv.Atoi(conm["_keepalive"])
	opts.SetKeepAlive(time.Duration(kplv) * time.Second)
	h.ClientID = conm["_client_id"]
	h.ServerID = conm["_server_name"]
	h.onlinemsg = onlinemsg
	opts.SetWill(h.ServerID+"/"+h.ClientID, willmsg, 1, true)
	optserver := "tcp://" + server
	if conm["cafile"] != "" {
		tlsconfig, err := newTLSConfig(conm)
		if err != nil {
			log.Errorf("Error with the mqtt CA certificate: %s", err)
		} else {
			opts.SetTLSConfig(tlsconfig)
			optserver = "ssl://" + server
		}
	}
	opts.AddBroker(optserver)

	log.WithField("server", server).Info("handler/mqtt: connecting to mqtt broker")
	h.conn = mqtt.NewClient(opts)
	for {
		if token := h.conn.Connect(); token.Wait() && token.Error() != nil {
			log.Errorf("handler/mqtt: connecting to broker error, will retry in 2s: %s", token.Error())
			time.Sleep(2 * time.Second)
		} else {
			log.Info("handler/mqtt: conneting successfull")
			break
		}
	}
	return &h, nil
}

func newTLSConfig(cm map[string]string) (*tls.Config, error) {
	// Import trusted certificates from CAfile.pem.
	cafile := cm["cafile"]
	cert, err := ioutil.ReadFile(cafile)
	//log.Infof("print this ca file :\n %s", cert)
	if err != nil {
		log.Errorf("backend: couldn't load cafile: %s", err)
		return nil, err
	}

	certpool := x509.NewCertPool()
	certpool.AppendCertsFromPEM(cert)

	// Create tls.Config with desired tls properties
	if cm["certfile"] != "" && cm["keyfile"] != "" {
		certpair, err := tls.LoadX509KeyPair(cm["certfile"], cm["keyfile"])
		if err != nil {
			log.Errorf("get cert error :%s", err)
			return nil, err
			//log.Fatalf("get cert error :%s", err)
		}
		return &tls.Config{
			RootCAs:            certpool,
			Certificates:       []tls.Certificate{certpair},
			InsecureSkipVerify: true,
		}, nil
	}
	return &tls.Config{
		// RootCAs = certs used to verify server cert.
		RootCAs:            certpool,
		InsecureSkipVerify: true,
	}, nil
}

// Close stops the handler.
func (h *MQTTHandler) Close() error {
	log.Info("handler/mqtt: closing handler")
	if token := h.conn.Unsubscribe(h.ClientID + "/" + h.ServerID); token.Wait() && token.Error() != nil {
		return fmt.Errorf("handler/mqtt: unsubscribe from %s error: %s", h.ClientID, token.Error())
	}
	log.Info("handler/mqtt: handling last items in queue")
	h.wg.Wait()
	close(h.dataDownChan)
	return nil
}

// SendDataUp sends a DataUpPayload.
func (h *MQTTHandler) SendDataUp(payload interface{}) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("handler/mqtt: data-up payload marshal error: %s", err)
	}

	topic := h.ServerID + "/" + h.ClientID
	if token := h.conn.Publish(topic, 0, false, b); token.Wait() && token.Error() != nil {
		return fmt.Errorf("handler/mqtt: publish data-up error: %s", err)
	}
	mymsg, err := simplejson.NewJson(b)
	logsb, _ := mymsg.EncodePretty()
	log.WithFields(log.Fields{"topic": topic, "messageType": "handler/mqtt: publishing data-send"}).Debugf(string(logsb))
	return nil
}

// SendSerDataUp sends a DataUpPayload.
func (h *MQTTHandler) SendSerDataUp(b []byte) error {

	topic := "serial" + "/" + h.ClientID
	if token := h.conn.Publish(topic, 0, false, b); token.Wait() && token.Error() != nil {
		return fmt.Errorf("handler/mqtt: publish data-up error: %s", token.Error())
	}
	log.WithFields(log.Fields{"topic": topic}).Info(b)
	return nil
}

// DataDownChan returns the channel containing the received DataDownPayload.
func (h *MQTTHandler) DataDownChan() chan DataDownPayload {
	return h.dataDownChan
}

func (h *MQTTHandler) rxmsgHandler(c mqtt.Client, msg mqtt.Message) {
	h.wg.Add(1)
	defer h.wg.Done()

	//	dec := json.NewDecoder(bytes.NewReader(msg.Payload()))
	//	if err := dec.Decode(&pl); err != nil {
	//		return
	//	}
	mymsgjson, err := simplejson.NewJson(msg.Payload())
	if err != nil {
		log.WithFields(log.Fields{
			"msg_payload": string(msg.Payload()),
		}).Errorf("message is not json format: %s", err)
		return
	}
	logsb, _ := mymsgjson.EncodePretty()
	log.WithFields(log.Fields{"topic": msg.Topic(), "messageType": "received cmd", "Qos": msg.Qos()}).Info(string(logsb))
	h.dataDownChan <- DataDownPayload{Pj: mymsgjson}
}

func (h *MQTTHandler) onConnected(c mqtt.Client) {
	log.Info("handler/mqtt: connected to mqtt broker")
	for {
		log.WithField("topic", h.ClientID+"/"+h.ServerID).Info("handler/mqtt: subscribling to things topic")
		if token := h.conn.Subscribe(h.ClientID+"/"+h.ServerID, 2, h.rxmsgHandler); token.Wait() && token.Error() != nil {
			log.WithField("topic", h.ClientID+"/"+h.ServerID).Errorf("handler/mqtt: subscribe error: %s", token.Error())
			time.Sleep(time.Second)
			continue
		}
		common.Mqttconnected = true
		h.conn.Publish(h.ServerID+"/"+h.ClientID, 1, true, h.onlinemsg)
		return
	}
}

// SubFirmware ..
func (h *MQTTHandler) SubFirmware(topic string) chan []byte {
	log.Infof("subscrib for %s message", topic)
	if token := h.conn.Subscribe(topic, 1, h.rxmfireware); token.Wait() && token.Error() != nil {
		log.Warning("subscrib for fireware message failed")
		return nil
	}
	return h.Firmware
}

func (h *MQTTHandler) rxmfireware(c mqtt.Client, msg mqtt.Message) {
	h.wg.Add(1)
	defer h.wg.Done()
	h.Firmware <- msg.Payload()
	h.conn.Unsubscribe(msg.Topic())
}

func (h *MQTTHandler) onConnectionLost(c mqtt.Client, reason error) {
	log.Errorf("handler/mqtt: mqtt connection error: %s", reason)
	common.Mqttconnected = false
	h.LostConnectTime = time.Now()
}

//IsConnected ..
func (h *MQTTHandler) IsConnected() bool {
	return h.conn.IsConnected()
}

// GetLostConnectTime ..
func (h *MQTTHandler) GetLostConnectTime() time.Time {
	return h.LostConnectTime
}
