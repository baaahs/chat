package main

import (
	"encoding/json"
	"fmt"
	"github.com/denisbrodbeck/machineid"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/eyethereal/go-config"
	"os"
	"regexp"
	"time"
)

type UIRefresher func()
type MessageRecver func(*Message)

type Message struct {
	Msg string  `json:"msg,omitempty"`
	From string `json:"from,omitempty"`
	Sent int64  `json:"sent,omitempty"`
	me bool
}


type Network struct {
	c mqtt.Client

	url string
	id string
	name string

	err error

	// External interface methods to be set by the UI
	RefreshUI   UIRefresher
	RecvMessage MessageRecver

	TimeToDie chan bool

	publishTopic string
}

func NewNetwork(cfg *config.AclNode) *Network {
	url := cfg.DefChildAsString("tcp://localhost:18830", "url")
	id := cfg.ChildAsString("id")
	if len(id) == 0 {
		var err error
		id, err = machineid.ProtectedID("bchat")
		log.Errorf("Failed to read machine id: %v", err)
		if err != nil {
			id, _ = os.Hostname()
		}
	}

	name := cfg.DefChildAsString(id, "name")

	// Make a Regex to say we only want letters and numbers
	reg, _ := regexp.Compile("[^a-zA-Z0-9]+")
	name = reg.ReplaceAllString(name, "-")

	net := &Network {
		url: url,
		id: id,
		name: name,
		TimeToDie: make(chan bool),
		publishTopic: fmt.Sprintf("bchat/rooms/main/%v/messages", name),
	}

	return net
}

func (net *Network) start() {
	log.Infof("Starting network %v %v %v", net.url, net.id, net.name)

	var opts = mqtt.NewClientOptions().AddBroker(net.url).SetClientID(net.id)
	opts.SetCleanSession(false)

	opts.SetConnectionLostHandler(func(client mqtt.Client, e error) {
		net.disconnect(e)
	})
	opts.SetOnConnectHandler(func(client mqtt.Client) {
		net.connect()
	})

	net.c = mqtt.NewClient(opts)

	go net.connectAndSubscribe()
}

func (net *Network)SendText(txt string) {
	m := Message{
		From: net.name,
		Msg: txt,
		Sent: time.Now().Unix(),
	}

	data, err := json.Marshal(m)
	if err != nil {
		log.Errorf("Can not send: %v", err)
		return
	}

	net.c.Publish(net.publishTopic, 2, false, data)
}

func (net *Network)disconnect(e error) {
	log.Infof("disconnect event %v", e)
	if net.RefreshUI != nil {
		net.RefreshUI()
	}
}

func (net *Network)connect() {
	log.Info("connect event")
	if net.RefreshUI != nil {
		net.RefreshUI()
	}
}

func (net *Network) connectAndSubscribe() {

	var numErrors = 0

	for {
		select {
		case <- net.TimeToDie:
			log.Info("Giving up on starting a network connection")
			return

		default:
			token := net.c.Connect()
			token.WaitTimeout(30 * time.Second)
			if token.Error() != nil {
				numErrors++
				log.Errorf("Connection Failure #%v: %v", numErrors, token.Error())

				if numErrors > 10 {
					log.Errorf("Giving up")
					panic("Too many connection errors")
				}

				// Always delay a few seconds before trying again
				time.Sleep(5 * time.Second)
			} else {
				// Hey we are connected now!
				log.Infof("Starting subscription call id=%v", net.id)
				t := net.c.Subscribe("bchat/rooms/main/+/messages", 2, func(client mqtt.Client, message mqtt.Message) {
					net.handleChat(message)
				})
				// This should be fast, but don't wait forever
				t.WaitTimeout(5 * time.Second)
				if t.Error() != nil {
					log.Errorf("Error subscribing: %v", t.Error())
					net.c.Disconnect(250)
				} else {
					// Presumably auto connect will take it from here so we can exit the outer loop
					log.Info("Subscribed")
					return
				}
			}
		}
	}
}

func (net *Network) handleChat(message mqtt.Message) {
	// With QOS 2 we don't have to worry about dupes
	msg := Message{}
	err := json.Unmarshal(message.Payload(), &msg)
	if err != nil {
		log.Warningf("Error unmarshalling a message: %v", err)
		return
	}

	if msg.From == net.name {
		msg.me = true
	}

	if net.RecvMessage != nil {
		net.RecvMessage(&msg)
	}
}