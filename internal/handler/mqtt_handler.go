package handler

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"
	"sync"
	"strconv"
//	"bytes"
//	"encoding/base64"
//	"regexp"
//	"strconv"
	simplejson "github.com/bitly/go-simplejson"
	log "github.com/Sirupsen/logrus"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)


// MQTTHandler implements a MQTT handler for sending and receiving data by
type DataDownPayload  struct{
	Pj		*simplejson.Json
}

type MQTTHandler struct {
	conn         mqtt.Client
	dataDownChan chan DataDownPayload
	wg           sync.WaitGroup
	Topic		  string
}


// NewMQTTHandler creates a new MQTTHandler.
func NewMQTTHandler(server, username, password, cafile ,trtop ,kl ,willmsg ,onlinemsg string) (Handler, error) {
	h := MQTTHandler{
		dataDownChan: make(chan DataDownPayload),
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(server)
	opts.SetUsername(username)
	opts.SetPassword(password)
	opts.SetOnConnectHandler(h.onConnected)
	opts.SetConnectionLostHandler(h.onConnectionLost)
	kplv,_ := strconv.Atoi(kl)
	opts.SetKeepAlive(time.Duration(kplv) * time.Second)
	h.Topic = trtop
	opts.SetWill("things/"+h.Topic, willmsg, 1, true)
	if cafile != "" {
		tlsconfig, err := newTLSConfig(cafile)
		if err != nil {
			log.Fatalf("Error with the mqtt CA certificate: %s", err)
		} else {
			opts.SetTLSConfig(tlsconfig)
		}
	}

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
	h.conn.Publish("things/"+h.Topic, 1, true, onlinemsg)
	return &h, nil
}

func newTLSConfig(cafile string) (*tls.Config, error) {
	// Import trusted certificates from CAfile.pem.

	cert, err := ioutil.ReadFile(cafile)
	if err != nil {
		log.Errorf("backend: couldn't load cafile: %s", err)
		return nil, err
	}

	certpool := x509.NewCertPool()
	certpool.AppendCertsFromPEM(cert)

	// Create tls.Config with desired tls properties
	return &tls.Config{
		// RootCAs = certs used to verify server cert.
		RootCAs: certpool,
	}, nil
}

// Close stops the handler.
func (h *MQTTHandler) Close() error {
	log.Info("handler/mqtt: closing handler")
	if token := h.conn.Unsubscribe(h.Topic); token.Wait() && token.Error() != nil {
		return fmt.Errorf("handler/mqtt: unsubscribe from %s error: %s", h.Topic, token.Error())
	}
	log.Info("handler/mqtt: handling last items in queue")
	h.wg.Wait()
	close(h.dataDownChan)
	return nil
}

// SendDataUp sends a DataUpPayload.
func (h *MQTTHandler) SendDataUp(payload interface {}) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("handler/mqtt: data-up payload marshal error: %s", err)
	}

	topic := "things/" + h.Topic
	if token := h.conn.Publish(topic, 0, false, b); token.Wait() && token.Error() != nil {
		return fmt.Errorf("handler/mqtt: publish data-up error: %s", err)
	}
	log.WithField("topic", topic).Info("handler/mqtt: publishing data-up")
	return nil
}
// DataDownChan returns the channel containing the received DataDownPayload.
func (h *MQTTHandler) DataDownChan() chan DataDownPayload {
	return h.dataDownChan
}

func (h *MQTTHandler) rxmsgHandler(c mqtt.Client, msg mqtt.Message) {
	h.wg.Add(1)
	defer h.wg.Done()

	log.WithField("topic", msg.Topic()).Info("payload received"+ fmt.Sprintf(" Qos=%d",msg.Qos()))
//	dec := json.NewDecoder(bytes.NewReader(msg.Payload()))
//	if err := dec.Decode(&pl); err != nil {
//		return
//	}
	 mymsgjson,err := simplejson.NewJson(msg.Payload())
	 if err != nil {
	 	log.WithFields(log.Fields{
	 			"msg2json":msg.Payload(),
	 	}).Errorf("message is not json format: %s", err)
	 	return
	 }
	h.dataDownChan <- DataDownPayload{ Pj:mymsgjson, }
}

func (h *MQTTHandler) onConnected(c mqtt.Client) {
	log.Info("handler/mqtt: connected to mqtt broker")
	for {
		log.WithField("topic",  h.Topic + "/things").Info("handler/mqtt: subscribling to things topic")
		if token := h.conn.Subscribe(h.Topic + "/things", 2, h.rxmsgHandler); token.Wait() && token.Error() != nil {
			log.WithField("topic", h.Topic + "/things").Errorf("handler/mqtt: subscribe error: %s", token.Error())
			time.Sleep(time.Second)
			continue
		}
		return
	}
}

func (h *MQTTHandler) onConnectionLost(c mqtt.Client, reason error) {
	log.Errorf("handler/mqtt: mqtt connection error: %s", reason)
}

func (h *MQTTHandler) IsConnected() bool{
	return h.conn.IsConnected()
}
