package handler

import (
//	"bytes"
	"crypto/tls"
	"crypto/x509"
//	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
//	"regexp"
//	"strconv"
	"sync"
	"time"
	simplejson "github.com/bitly/go-simplejson"
	log "github.com/Sirupsen/logrus"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const Tdelay = time.Millisecond * 100

//var txTopicRegex = regexp.MustCompile(`(\w+)(\w+)/tx`)

// MQTTHandler implements a MQTT handler for sending and receiving data by
type MQTTHandler struct {
	conn         mqtt.Client
	dataDownChan chan DataDownPayload
	wg           sync.WaitGroup
	Topic		  string
}

// NewMQTTHandler creates a new MQTTHandler.
func NewMQTTHandler(server, username, password, cafile ,trtop string) (Handler, error) {
	h := MQTTHandler{
		dataDownChan: make(chan DataDownPayload),
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(server)
	opts.SetUsername(username)
	opts.SetPassword(password)
	opts.SetOnConnectHandler(h.onConnected)
	opts.SetConnectionLostHandler(h.onConnectionLost)
	h.Topic = trtop
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
			log.Info("handeler/mqtt: conneting successfull")
			break
		}
	}
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
	log.WithField("topic", topic).Info("handler/mqtt: publishing data-up")
	if token := h.conn.Publish(topic, 0, false, b); token.Wait() && token.Error() != nil {
		return fmt.Errorf("handler/mqtt: publish data-up error: %s", err)
	}
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

	// get the name of the application and node from the topic
//	match := txTopicRegex.FindStringSubmatch(msg.Topic())
//	if len(match) != 3 {
//		log.WithField("topic", msg.Topic()).Errorf("handler/mqtt: topic regex match error %s",msg.Payload())
//		return
//	}

//	var pl DataDownPayload
//	dec := json.NewDecoder(bytes.NewReader(msg.Payload()))
//	if err := dec.Decode(&pl); err != nil {
//		log.WithFields(log.Fields{
//			"data_base64": base64.StdEncoding.EncodeToString(msg.Payload()),
//		}).Errorf("handler/mqtt: tx payload unmarshal error: %s", err)
//		return
//	}
	 mymsgjson,err := simplejson.NewJson(msg.Payload())
	 if err != nil {
	 	log.WithFields(log.Fields{
	 			"msg2json":msg.Payload(),
	 	}).Errorf("message is not json format: %s", err)
	 	return
	 }
	 pj,_ := mymsgjson.EncodePretty()
//    fmt.Printf("%s\n",pj)
    
//    date := make(map[string]string)
//    date["test"] = "你来搞事情"
//    s := DataupPayload{
//    	Header:Header{
//    		Devid:h.Topic,
//    	},
//    	Request:Request{
//    		Data:date,
//    		Timestamp:time.Now().Unix(),
//    	},
//    }
//    h.SendDataUp(s)

	h.dataDownChan <- DataDownPayload {
		Pj:pj,
	}
}

func (h *MQTTHandler) onConnected(c mqtt.Client) {
	log.Info("handler/mqtt: connected to mqtt broker")
	for {
		log.WithField("topic",  h.Topic + "/things").Info("handler/mqtt: subscribling to tx topic")
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
