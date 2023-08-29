package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/denisbrodbeck/machineid"
	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
	"github.com/eyethereal/go-archercl"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
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

//////////////////////////////////////////////////////////////////////

// PahoLoggingShim is a shim between the Paho MQTT library and our
// logging infrastructure we already have in place.
//
// The Paho MQTT library has a couple of points where a logging object can
// be injected. This is objectively a great logging design based on
// dependency injection. This is the shim object that lets us bridge
// from their interface into the logging infrastructure we are already
// using.
type PahoLoggingShim struct {
	isDebug bool
	prefix  string
}

func (shim *PahoLoggingShim) Println(v ...interface{}) {
	if len(shim.prefix) == 0 {
		if shim.isDebug {
			log.Debug(v...)
		} else {
			log.Warning(v...)
		}
	} else {
		sl := []interface{}{shim.prefix}
		sl = append(sl, ": ")
		sl = append(sl, v...)
		if shim.isDebug {
			log.Debug(sl)
		} else {
			log.Warning(sl)
		}
	}
}

func (shim *PahoLoggingShim) Printf(format string, v ...interface{}) {
	if len(shim.prefix) == 0 {
		if shim.isDebug {
			log.Debugf(format, v...)
		} else {
			log.Warningf(format, v...)
		}
	} else {
		format = fmt.Sprintf("%v: %v", shim.prefix, format)
		if shim.isDebug {
			log.Debugf(format, v...)
		} else {
			log.Warningf(format, v...)
		}
	}
}

//////////////////////////////////////////////////////////////////////

// Network is the thing which, you know, does network stuff. Specifically
// it handles the MQTT connection as configured from the regular configuration
// infrastructure.
type Network struct {
	cliCfg autopaho.ClientConfig

	ctx    context.Context
	cancel context.CancelFunc
	conmgr *autopaho.ConnectionManager

	url  string
	id   string
	name string

	err error

	// External interface methods to be set by the UI
	RefreshUI   UIRefresher
	RecvMessage MessageRecver

	isConnected bool

	TimeToDie chan bool

	publishTopic string

	// This should NOT be here, but I guess it makes sense as
	// maintaining what is known on the network (vs. what might have been cached
	// locally)
	sheepLat float64
	sheepLon float64

	haveSheepPos     bool
	sheepPosAt       time.Time
	sheepPosErrorStr string
}

func NewNetwork(cfg *archercl.AclNode) *Network {
	cfgurl := cfg.DefChildAsString("tcp://localhost:1883", "url")
	id := cfg.ChildAsString("id")

	if len(id) == 0 {
		var err error
		id, err = machineid.ProtectedID("bchat")
		if err != nil {
			log.Errorf("Failed to read machine id. Will use hostname: %v", err)
			id, _ = os.Hostname()
		}
	}
	log.Infof("Using id of '%v'", id)

	defName := id
	if len(defName) > 6 {
		defName = defName[len(defName)-6:]
	}
	name := cfg.DefChildAsString(defName, "name")

	// Make a Regex to say we only want letters and numbers
	reg, _ := regexp.Compile("[^a-zA-Z0-9]+")
	name = reg.ReplaceAllString(name, "-")

	serverUrl, err := url.Parse(cfgurl)
	if serverUrl == nil {
		log.Errorf("Failed to parse url %v : %v", cfgurl, err)
		return nil
	}

	bnet := &Network{
		cliCfg: autopaho.ClientConfig{
			BrokerUrls:        []*url.URL{serverUrl},
			KeepAlive:         60, // In Seconds. Must be configured or code fails
			ConnectRetryDelay: time.Duration(10) * time.Second,
			ConnectTimeout:    time.Duration(10) * time.Second,

			Debug:      &PahoLoggingShim{isDebug: true, prefix: "CM Debug"},
			PahoDebug:  &PahoLoggingShim{isDebug: true, prefix: "Paho Debug"},
			PahoErrors: &PahoLoggingShim{isDebug: false, prefix: "Paho"},
		},
		url:          cfgurl,
		id:           id,
		name:         name,
		TimeToDie:    make(chan bool),
		publishTopic: fmt.Sprintf("bchat/rooms/main/%v/messages", name),
	}

	bnet.cliCfg.OnConnectionUp = bnet.onConnectionUp
	bnet.cliCfg.OnServerDisconnect = bnet.onServerDisconnect

	// Some shit we gots to pass in to the actual single instance config
	bnet.cliCfg.ClientConfig.ClientID = id

	// I'm pretty sure it would go badly if we tried to let the infrastructure
	// provide a default router, so let's initialize our own, although we are
	// using the default implementation without a lot of fancy
	router := paho.NewStandardRouter()
	router.SetDebugLogger(&PahoLoggingShim{isDebug: true, prefix: "Router"})
	router.RegisterHandler("bchat/rooms/main/+/messages", func(msg *paho.Publish) {
		bnet.gotMessage(msg)
	})
	router.RegisterHandler("bchat/rooms/main/sheep_loc", func(msg *paho.Publish) {
		bnet.gotSheepLoc(msg)
	})

	bnet.cliCfg.ClientConfig.Router = router

	return bnet
}

func (bnet *Network) Start() {

	if bnet.conmgr != nil {
		log.Warning("Starting a Network while an existing Connection Manager is around. Probably bad. Telling that other one to disconnect")
		bnet.conmgr.Disconnect(bnet.ctx)

		// Of course this will probably prevent that from happening
		bnet.cancel()
	}

	log.Infof("Starting network %v %v %v", bnet.url, bnet.id, bnet.name)

	bnet.ctx, bnet.cancel = context.WithCancel(context.Background())

	var err error = nil
	bnet.conmgr, err = autopaho.NewConnection(bnet.ctx, bnet.cliCfg)
	if bnet.conmgr == nil {
		log.Errorf("Unable to create a new connection manager: %v", err)
		return
	}

	return
	bnet.conmgr.AwaitConnection(bnet.ctx)
	//
	//	var opts = mqtt.NewClientOptions().AddBroker(net.url).SetClientID(net.id)
	//	opts.SetCleanSession(false)
	//
	//	opts.SetConnectionLostHandler(net.evtDisconnect)
	//	opts.SetOnConnectHandler(net.evtConnect)
	//
	//	net.c = mqtt.NewClient(opts)
	//
	//	go net.tryToConnect()
}

func (bnet *Network) onConnectionUp(conMgr *autopaho.ConnectionManager, connAck *paho.Connack) {
	log.Info("On Connection Up for new network class")

	bnet.isConnected = true

	// Every time we connect we want to subscribe. THe assumption being made is that
	// it's okay to re-subscribe even if the server already knew about this. However
	// if the server is totally new (like it bounced without persistence) this is
	// new information to it potentially

	_, err := conMgr.Subscribe(context.Background(), &paho.Subscribe{
		Subscriptions: map[string]paho.SubscribeOptions{
			"bchat/rooms/main/+/messages": {QoS: 2},
			"bchat/rooms/main/sheep_loc":  {QoS: 2},
		},
	})
	if err != nil {
		log.Errorf("Subscription failure: %v", err)
	}

	if bnet.RefreshUI != nil {
		bnet.RefreshUI()
	}
}

func (bnet *Network) onServerDisconnect(disconnect *paho.Disconnect) {
	log.Warning("Server disconnect %v : %v",
		disconnect.Properties.ServerReference,
		disconnect.Properties.ReasonString)

	bnet.isConnected = false

	if bnet.RefreshUI != nil {
		bnet.RefreshUI()
	}
}

func (bnet *Network) gotMessage(recv *paho.Publish) {
	str := recv.String()

	log.Debugf("Message: %v", str)

	// With QOS 2 we don't have to worry about dupes
	msg := Message{}
	err := json.Unmarshal(recv.Payload, &msg)
	if err != nil {
		log.Warningf("Error unmarshalling a message: %v", err)
		return
	}

	if msg.From == bnet.name {
		msg.me = true
	}

	if bnet.RecvMessage != nil {
		bnet.RecvMessage(&msg)
	} else {
		log.Warning("Dropping message because there is no RecvMessage")
	}

}

func (bnet *Network) gotSheepLoc(msg *paho.Publish) {
	str := string(msg.Payload)

	log.Debugf("SheepLoc: %v", str)

	// Let's try and figure out if it's a simple or more complex format
	ix := strings.Index(str, ":") /// This will be there in JSON
	if ix == -1 {
		// Treat it as a simple string
		list := strings.Split(str, ",")
		if len(list) < 2 {
			bnet.sheepPosErrorStr = fmt.Sprintf("ERR: Bad loc #{str}")
			bnet.haveSheepPos = false
		} else {
			// Okay cool, turn it into two float64's!
			lat, err_lat := strconv.ParseFloat(list[0], 64)
			lon, err_lon := strconv.ParseFloat(list[1], 64)
			if err_lat != nil || err_lon != nil {
				bnet.sheepPosErrorStr = fmt.Sprintf("ERR: Unparsable #{str}")
				bnet.haveSheepPos = false
			} else {
				bnet.sheepLat = lat
				bnet.sheepLon = lon
				bnet.haveSheepPos = true
				bnet.sheepPosAt = time.Now()
				bnet.sheepPosErrorStr = ""
				log.Infof("Saved sheep location as %v, %v at ", bnet.sheepLat, bnet.sheepLon, bnet.sheepPosAt)
			}
		}
	} else {
		log.Errorf("Position is probably JSON but we aren't ready ix=%v str='%v'", ix, str)
		bnet.sheepPosErrorStr = "ERROR: JSON Position"
		bnet.haveSheepPos = false
	}

	if bnet.RefreshUI != nil {
		bnet.RefreshUI()
	}

}

// SendText will send the given string in the given context, which should
// itself be a sub-context of bnet.context if specified. If not specified then
// the underlying context will be used directly. That's probably what should
// be the normal mode of operation but just in case the context is there for
// people who would want it.
//
// This is a blocking call and thus would generally be called on a coroutine
// which will handle a non-nil error response appropriately. The most expected
// error condition will be a message send timeout.
func (bnet *Network) SendText(ctx context.Context, txt string) error {
	m := Message{
		From: bnet.name,
		Msg:  txt,
		Sent: time.Now().Unix(),
	}

	data, err := json.Marshal(m)
	if err != nil {
		log.Errorf("Can not send: %v", err)
		return err
	}

	// Create our Publish packet
	me := uint32(3600)
	publish := &paho.Publish{
		QoS:    2,
		Retain: true,
		Topic:  bnet.publishTopic,
		Properties: &paho.PublishProperties{
			MessageExpiry: &me,                // In seconds. 3600 is one hour
			ContentType:   "application/json", // be nice and declare this
		},
		Payload: data,
	}

	if ctx == nil {
		ctx = bnet.ctx
	}

	// Hey maybe we care about success but really we don't?
	_, err = bnet.conmgr.Publish(ctx, publish)

	return err
}

func (bnet *Network) IsConnected() bool {
	return bnet.conmgr != nil && bnet.isConnected
}

//func (net *Network) SendText(txt string) {
//	m := Message{
//		From: net.name,
//		Msg:  txt,
//		Sent: time.Now().Unix(),
//	}
//
//	data, err := json.Marshal(m)
//	if err != nil {
//		log.Errorf("Can not send: %v", err)
//		return
//	}
//
//	net.c.Publish(net.publishTopic, 2, false, data)
//}
//
//func (net *Network) evtDisconnect(client mqtt.Client, e error) {
//	log.Infof("disconnect event %v", e)
//	if net.RefreshUI != nil {
//		net.RefreshUI()
//	}
//}
//
//func (net *Network) evtConnect(client mqtt.Client) {
//	log.Info("connect event")
//
//	// Every time we connect we want to subscribe. THe assumption being made is that
//	// it's okay to re-subscribe even if the server already knew about this. However
//	// if the server is totally new (like it bounced without persistence) this is
//	// new information to it potentially
//	log.Infof("Starting subscription call id=%v", net.id)
//
//	log.Infof("net.c.isConnected=%v", net.c.IsConnected())
//
//	tkn := net.c.Subscribe("bchat/rooms/main/+/messages", 2, func(client mqtt.Client, message mqtt.Message) {
//		net.handleChat(message)
//	})
//	// This should be fast, but don'tkn wait forever
//	tkn.WaitTimeout(5 * time.Second)
//	if tkn.Error() != nil {
//		log.Errorf("Error subscribing: %v", tkn.Error())
//		//net.c.Disconnect(250)
//	} else {
//		log.Info("Subscribed")
//	}
//
//	if net.RefreshUI != nil {
//		net.RefreshUI()
//	}
//}
//
//func (net *Network) tryToConnect() {
//
//	var numErrors = 0
//
//	for {
//		select {
//		case <-net.TimeToDie:
//			log.Info("Giving up on starting a network connection")
//			return
//
//		default:
//			token := net.c.Connect()
//			token.WaitTimeout(30 * time.Second)
//			if token.Error() != nil {
//				numErrors++
//				log.Errorf("Connection Failure #%v: %v", numErrors, token.Error())
//
//				if numErrors > 10 {
//					log.Errorf("Giving up")
//					panic("Too many connection errors")
//				}
//
//				// Always delay a few seconds before trying again
//				time.Sleep(5 * time.Second)
//			} else {
//				// Hey we are connected now!
//				return
//			}
//		}
//	}
//}
//
//func (net *Network) handleChat(message mqtt.Message) {
//	// With QOS 2 we don't have to worry about dupes
//	msg := Message{}
//	err := json.Unmarshal(message.Payload(), &msg)
//	if err != nil {
//		log.Warningf("Error unmarshalling a message: %v", err)
//		return
//	}
//
//	if msg.From == net.name {
//		msg.me = true
//	}
//
//	if net.RecvMessage != nil {
//		net.RecvMessage(&msg)
//	}
//}
