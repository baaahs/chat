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

	log.Infof("Starting bchat-tty on %v", cfg.ChildAsString("tty"))

	net.start()
	ui.Run()
}
