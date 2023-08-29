package main

import (
	"github.com/eyethereal/go-archercl"
	"github.com/godbus/dbus"
	"time"
)

type SysStat struct {
	conn *dbus.Conn

	PowerPercent chan float64
	TimeString   chan string
}

func NewSysStat(cfg *archercl.AclNode) *SysStat {
	ss := &SysStat{
		PowerPercent: make(chan float64, 4),
		TimeString:   make(chan string, 4),
	}
	return ss
}

func (ss *SysStat) Run() {
	var err error

	// Do this before the battery thing that can fail
	go ss.CheckTime()

	ss.conn, err = dbus.SystemBus()

	if err != nil {
		log.Errorf("Failed to connect to system bus: %v", err)
		ss.PowerPercent <- 0.69
	} else {
		log.Info("Got system dbus connection")
	}

	go ss.CheckBattery()

	// Let's also send the time every 10 seconds or so
}

//	type PowerInfo struct {
//	   NativePath string
//	   Vendor string
//	   Model string
//	   Serial string
//	   UpdateTime uint64
//	   Type uint32
//	   PowerSupply bool
//	   HasHistory bool
//	   HasStatistics bool
//	   Online bool
//	   Energy float64
//	   EnergyEmpty float64
//	   EnergyFull float64
//	   EnergyFullDesign float64
//	   EnergyRate float64
//	}
func (ss *SysStat) CheckBattery() {
	for {
		info := make(map[string]dbus.Variant)

		if ss.conn != nil {
			err := ss.conn.
				Object("org.freedesktop.UPower",
					"/org/freedesktop/UPower/devices/DisplayDevice").
				Call("org.freedesktop.DBus.Properties.GetAll", 0, "org.freedesktop.UPower.Device").
				Store(info)

			if err != nil {
				log.Errorf("Failed to get battery info: %v", err)
				ss.PowerPercent <- float64(69.6)
			} else {
				// Not need to totally spam syslog
				//log.Noticef("DisplayDevice Info: %v", info)

				if val, ok := info["Percentage"]; ok {
					if p, isFloat := val.Value().(float64); isFloat {
						// log.Warningf("Power Percentage is %v", p)
						ss.PowerPercent <- p
					} else {
						log.Errorf("Expected a float for Percentage from DisplayDevice property Percentage")
					}
				} else {
					log.Errorf("No 'Percentage' value in the returned map")
				}
			}
		} else {
			ss.PowerPercent <- float64(12.3)
		}

		time.Sleep(2 * time.Second)
	}
}

func (ss *SysStat) CheckTime() {
	for {
		now := time.Now()

		formatted := now.Format(" 15:04 MST")

		//log.Warningf("Sending time %v", formatted)
		ss.TimeString <- formatted

		time.Sleep(10 * time.Second)
	}
}
