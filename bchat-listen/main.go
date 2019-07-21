package main

import (
	"fmt"
	"os"
	"time"

	"github.com/eclipse/paho.mqtt.golang"
)

var f mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("%v : %v\n", msg.Topic(), msg.Payload())
}

var TOPIC = "bchat/#"

func main() {
	var opts = mqtt.NewClientOptions().AddBroker("tcp://localhost:18830").SetClientID("listener")
	opts.SetKeepAlive(2 * time.Second)
	opts.SetDefaultPublishHandler(f)
	opts.SetPingTimeout(1 * time.Second)

	c := mqtt.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	if token := c.Subscribe(TOPIC, 0, nil); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}

	fmt.Printf("Listening for messages on %v\n", TOPIC)
	time.Sleep(30 * time.Second)

	fmt.Printf("Done. Unsubscribing and exiting\n")
	if token := c.Unsubscribe(TOPIC); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}

	c.Disconnect(250)

	time.Sleep(1 * time.Second)
}
