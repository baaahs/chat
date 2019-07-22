package main

import (
	"github.com/eyethereal/go-archercl"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("bchat")

func main() {
	opts := &archercl.Opts{
		Name: "bchat",
		DefaultText: `
		logging level: debug
		logging backends {
			ui {
				type: delayed
				format: "%{time:15:04} %{level} %{message}"
			}

			syslog {
				type: syslog
				facility: user
				prefix: bchat
				level: debug
			}
		}

		dumpConfig: true
		dumpColor: false
		`,
		DumpConfig: true,

		// Set this so we can see early log messages, but not desirable
		// in most deployments
		AddColorConsoleLogging: false,
	}

	cfg, err := archercl.Load(opts)
	if err != nil {
		panic(err)
	}

	net := NewNetwork(cfg)
	ui := NewUI(cfg, net)

	if be, ok := archercl.GetBackend("ui").(*archercl.DelayedBackend); ok {
		be.SetRealBackend(ui)
	} else {
		log.Debug(cfg.String())
		panic("What? no ui backend...")
	}
	//backends := make([]logging.Backend, 1)
	//
	//// Attach the UI. It might not always display but it can cache I guess
	//backends[0] = ui
	//
	//// Always log to syslog
	//sl, e := logging.NewSyslogBackendPriority("bchat", syslog.LOG_DEBUG)
	//if e != nil {
	//	log.Warning("Unable to create syslog backend")
	//} else {
	//	backends = append(backends, sl)
	//}
	//
	////fName := fmt.Sprintf("bchat-%v.log", os.Getpid())
	////file, err := os.OpenFile(fName, os.O_WRONLY|os.O_APPEND|os.O_CREATE, os.FileMode(0666))
	////if err != nil {
	////	log.Panicf("Unable to open file '%s' : %s", fName, err)
	////	return
	////}
	////
	////fBE := logging.NewLogBackend(file, "", 0)
	//
	////logging.SetBackend(ui, fBE)
	//logging.SetBackend(backends...)

	log.Error("Starting bchat-tty")

	net.start()
	ui.Run()
}
