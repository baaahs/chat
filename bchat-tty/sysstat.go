package main

import (
    "github.com/eyethereal/go-archercl"
    "github.com/godbus/dbus"
    "time"
)

type SysStat struct {
    conn *dbus.Conn
}

func NewSysStat(cfg *archercl.AclNode) *SysStat {
    ss := &SysStat{
    }
    return ss
}

func (ss *SysStat) Run() {
    var err error
    ss.conn, err = dbus.SystemBus()

    if err != nil {
        log.Errorf("Failed to connect to system bus: %v", err)
        return
    }
    log.Info("Got system dbus connection")

    go ss.CheckBattery()
}

//type PowerInfo struct {
//    NativePath string
//    Vendor string
//    Model string
//    Serial string
//    UpdateTime uint64
//    Type uint32
//    PowerSupply bool
//    HasHistory bool
//    HasStatistics bool
//    Online bool
//    Energy float64
//    EnergyEmpty float64
//    EnergyFull float64
//    EnergyFullDesign float64
//    EnergyRate float64
//}
func (ss *SysStat) CheckBattery() {
    for {
        info := make(map[string]dbus.Variant)

        err := ss.conn.
            Object("org.freedesktop.UPower",
                "/org/freedesktop/UPower/devices/DisplayDevice").
            Call("org.freedesktop.DBus.Properties.GetAll", 0, "org.freedesktop.UPower.Device").
            Store(info)

        if err != nil {
            log.Errorf("Failed to get battery info: %v", err)
        } else {
            log.Noticef("DisplayDevice Info: %v", info)

            if val, ok := info["Percentage"]; ok {
                if p, isFloat := val.Value().(float64); isFloat {
                    log.Warningf("Power Percentage is %v", p)
                } else {
                    log.Errorf("Expected a float for Percentage from DisplayDevice property Percentage")
                }
            } else {
                log.Errorf("No 'Percentage' value in the returned map")
            }
        }

        time.Sleep(2 * time.Second)
    }
}
