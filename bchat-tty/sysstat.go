package main

import (
    "github.com/eyethereal/go-archercl"
    "github.com/godbus/dbus"
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

}
