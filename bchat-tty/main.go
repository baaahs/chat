package main

import (
	"github.com/eyethereal/go-archercl"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("bchat")

func main() {

	// Phase 0 of the boot sequence: Configuration

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

	if cfg.ChildAsBool("configDebug") {
		log.Error("Stopping because configDebug was set")
		return
	}

	// Moving on to boot phase 1: Instantiation of system components

	bnet := NewNetwork(cfg)
	//oldnet := NewOldNetwork(cfg)

	ss := NewSysStat(cfg)

	// Don't really love breaking dependency injection here at the UI layer, but
	// what the hell. Go for it! Who cares!!!
	ui := NewUI(cfg, bnet, ss)

	if be, ok := archercl.GetBackend("ui").(*archercl.DelayedBackend); ok {
		be.SetRealBackend(ui)
	} else {
		log.Debug(cfg.String())
		panic("What? no ui backend...")
	}

	// And now Phase 2: Actually starting the components we instantiated now that they
	// are all wired together

	log.Infof("Starting bchat-tty on %v", cfg.ChildAsString("tty"))

	bnet.Start()
	//oldnet.Start()
	ss.Run()
	ui.Run()
}
