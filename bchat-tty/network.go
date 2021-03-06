package main

import (
	"encoding/json"
	"fmt"
	"github.com/denisbrodbeck/machineid"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/eyethereal/go-archercl"
	"os"
	"regexp"
	"time"
)

type UIRefresher func()
type MessageRecver func(*Message)

type Message struct {
	Msg  string `json:"msg,omitempty"`
	From string `json:"from,omitempty"`
	Sent int64  `json:"sent,omitempty"`
	me   bool
}

type Network struct {
	c mqtt.Client

	url  string
	id   string
	name string

	err error

	// External interface methods to be set by the UI
	RefreshUI   UIRefresher
	RecvMessage MessageRecver

	TimeToDie chan bool

	publishTopic string
}

func NewNetwork(cfg *archercl.AclNode) *Network {
	url := cfg.DefChildAsString("tcp://localhost:1883", "url")
	id := cfg.ChildAsString("id")

	if len(id) == 0 {
		var err error
		id, err = machineid.ProtectedID("bchat")
		if err != nil {
			log.Errorf("Failed to read machine id. Will use hostname: %v", err)
			id, _ = os.Hostname()
		}
	}

	defName := id
	if len(defName) > 6 {
		defName = defName[len(defName)-6:]
	}
	name := cfg.DefChildAsString(defName, "name")

	// Make a Regex to say we only want letters and numbers
	reg, _ := regexp.Compile("[^a-zA-Z0-9]+")
	name = reg.ReplaceAllString(name, "-")

	net := &Network{
		url:          url,
		id:           id,
		name:         name,
		TimeToDie:    make(chan bool),
		publishTopic: fmt.Sprintf("bchat/rooms/main/%v/messages", name),
	}

	return net
}

func (net *Network) start() {
	log.Infof("Starting network %v %v %v", net.url, net.id, net.name)

	var opts = mqtt.NewClientOptions().AddBroker(net.url).SetClientID(net.id)
	opts.SetCleanSession(false)

	opts.SetConnectionLostHandler(net.evtDisconnect)
	opts.SetOnConnectHandler(net.evtConnect)

	net.c = mqtt.NewClient(opts)

	go net.tryToConnect()
}

func (net *Network) SendText(txt string) {
	m := Message{
		From: net.name,
		Msg:  txt,
		Sent: time.Now().Unix(),
	}

	data, err := json.Marshal(m)
	if err != nil {
		log.Errorf("Can not send: %v", err)
		return
	}

	net.c.Publish(net.publishTopic, 2, false, data)
}

func (net *Network) evtDisconnect(client mqtt.Client, e error) {
	log.Infof("disconnect event %v", e)
	if net.RefreshUI != nil {
		net.RefreshUI()
	}
}

func (net *Network) evtConnect(client mqtt.Client) {
	log.Info("connect event")

	// Every time we connect we want to subscribe. THe assumption being made is that
	// it's okay to re-subscribe even if the server already knew about this. However
	// if the server is totally new (like it bounced without persistence) this is
	// new information to it potentially
	log.Infof("Starting subscription call id=%v", net.id)

	log.Infof("net.c.isConnected=%v", net.c.IsConnected())

	tkn := net.c.Subscribe("bchat/rooms/main/+/messages", 2, func(client mqtt.Client, message mqtt.Message) {
		net.handleChat(message)
	})
	// This should be fast, but don'tkn wait forever
	tkn.WaitTimeout(5 * time.Second)
	if tkn.Error() != nil {
		log.Errorf("Error subscribing: %v", tkn.Error())
		//net.c.Disconnect(250)
	} else {
		log.Info("Subscribed")
	}

	if net.RefreshUI != nil {
		net.RefreshUI()
	}
}

func (net *Network) tryToConnect() {

	var numErrors = 0

	for {
		select {
		case <-net.TimeToDie:
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
				return
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
