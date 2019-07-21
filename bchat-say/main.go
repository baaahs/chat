package main

import (
	"fmt"
	"time"

	"github.com/eclipse/paho.mqtt.golang"
)

var f mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("%v : %v\n", msg.Topic(), msg.Payload())
}

var TOPIC = "bchat/rooms/main/messages"

func main() {
	var opts = mqtt.NewClientOptions().AddBroker("tcp://localhost:1883").SetClientID("say")
	opts.SetKeepAlive(2 * time.Second)
	opts.SetPingTimeout(1 * time.Second)

	c := mqtt.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	fmt.Printf("Sending a message to %v\n", TOPIC)

	token := c.Publish(TOPIC, 0, false, "Hello World!")
	token.Wait()

	c.Disconnect(250)

	time.Sleep(1 * time.Second)
}
