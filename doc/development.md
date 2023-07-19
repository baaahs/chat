# Developing BChat

This document describes how to get up and running making modifications to the bchat go code. This was originally developed on Mac OS so the directions here are for Mac. They should be very similar for Linux, and are probably completely different for Windows.

Expect to have a bazillion shell windows open while working on this code. You're going to want at least one, and likely two mosquitto servers (on different ports) and then likely multiple instances of the `bchat-tty` or other clients running at the same time sending messages around. Big monitors are nice.

## go

As mentioned in the README, if you're working on the reference code you're going to need a go develoment environment. Install go from [the golang website](https://go.dev/doc/install).

There are several options, but the straightforward common one is download the package for you world and install it.

This should give you the `go` command in your path. Verify by running `go -version` (You may need to open a new shell or run `rehash`)

This code was originally developed with golang version something earlier. It has now been updated to run correctly with version 1.20.6 which is current as of Jul 2023. The only real thing this has meant so far is greating a `go.mod` file in the root of the project tree using the command:

    go mod init github.com/baaahs/chat

If things are good on your system, you should be able to run

    go mod tidy

from the root of this repo and have it download all required modules used by the various packages.

There are at least four different command line programs that all define a main package. 

   * **bchat-tty** is the main application. It is meant to be run by `/sbin/agetty` in place of the `/bin/login` command on tty1 - the first linux virtual console. See the `chromebooks` directory which contains ansible stuff for updating a "chromebook" instance via ssh for all of that configuration.
   * **bchat-say** a test program that sends a simple hello world test message to the MQTT server.
   * **bchat-listen** a companion test program to say which receives a message and shows it. The combination of these two show the most basic usage of MQTT.
   * **bchat-echo** a very simple test of the readline library

To run any of these programs while developing, cd into the relevant directory and use `go run .`

## mosquitto

Directions are in the readme. Should be a `brew install mosquitto` on Mac.

Run one or more local instances using the config files in the mosquitto directory. The `local.conf` explains all config options, although there might be some updates since it was created from an older version of mosquitto. The `local2.conf` defines a second server which connects to the first.

These two files are roughly correct for what should be deployed in the playa, but beware that additional tweaks may be necessary as the exact network details come together. For instance specific IP addresses, ports, etc.

## goland IDE

The JetBrains IDE for go is `GoLand`. If you have access to a license you probably want to use it. Otherwise, something like Sublime or vim or whatever is just fine as well.

To enable running from the IDE, I created a new run configuration using the "Go Build" template setting the following values:

  * **Name** bchat-tty
  * **Run kind** Package
  * **Package path** github.com/baaahs/chat/bchat-tty
  * **Run after build** checked
  * **Working Directory** /code/github/baaahs/chat/bchat-tty (as appropriate for your machine)
  * **Program Arguments** id=ide (So as not to conflict with other instances, see the readme)
