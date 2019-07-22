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

	backends := make([]logging.Backend, 1)

	// Attach the UI. It might not always display but it can cache I guess
	backends[0] = ui

	// Always log to syslog
	sl, e := logging.NewSyslogBackend("bchat")
	if e != nil {
		log.Warning("Unable to create syslog backend")
	} else {
		backends = append(backends, sl)
	}

	//fName := fmt.Sprintf("bchat-%v.log", os.Getpid())
	//file, err := os.OpenFile(fName, os.O_WRONLY|os.O_APPEND|os.O_CREATE, os.FileMode(0666))
	//if err != nil {
	//	log.Panicf("Unable to open file '%s' : %s", fName, err)
	//	return
	//}
	//
	//fBE := logging.NewLogBackend(file, "", 0)

	//logging.SetBackend(ui, fBE)
	logging.SetBackend(backends...)

	log.Info("Starting the app")

	net.start()
	ui.Run()
}