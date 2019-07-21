package main

import (
	"github.com/eyethereal/go-config"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("bchat")


func main() {
	cfg := config.LoadACLConfig("bchat", "")

	net := NewNetwork(cfg)
	ui := NewUI(cfg, net)

	//fName := fmt.Sprintf("bchat-%v.log", os.Getpid())
	//file, err := os.OpenFile(fName, os.O_WRONLY|os.O_APPEND|os.O_CREATE, os.FileMode(0666))
	//if err != nil {
	//	log.Panicf("Unable to open file '%s' : %s", fName, err)
	//	return
	//}
	//
	//fBE := logging.NewLogBackend(file, "", 0)

	//logging.SetBackend(ui, fBE)
	logging.SetBackend(ui)

	log.Info("Starting the app")

	net.start()
	ui.Run()
}