package main

import (
	"fmt"
	"github.com/eyethereal/go-archercl"
	"github.com/gdamore/tcell"
	"github.com/op/go-logging"
	"github.com/rivo/tview"
)


/*

	It seems like these are the 16 colors supported by the terminal

	// While we _could_ presumably make a nice palette, why would we do that???

   echo -en "\e]P0222222" #black    -> this is the background color as well.
   echo -en "\e]P1803232" #darkred
   echo -en "\e]P25b762f" #darkgreen
   echo -en "\e]P3aa9943" #brown
   echo -en "\e]P4324c80" #darkblue
   echo -en "\e]P5706c9a" #darkmagenta
   echo -en "\e]P692b19e" #darkcyan
   echo -en "\e]P7ffffff" #lightgray
   echo -en "\e]P8222222" #darkgray
   echo -en "\e]P9982b2b" #red
   echo -en "\e]PA89b83f" #green
   echo -en "\e]PBefef60" #yellow
   echo -en "\e]PC2b4f98" #blue
   echo -en "\e]PD826ab1" #magenta
   echo -en "\e]PEa1cdcd" #cyan
   echo -en "\e]PFdedede" #white


*/

type UI struct {
	net *Network
	messageView *tview.TextView
	inputField *tview.InputField

	logText *tview.TextView

	statusText *tview.TextView
	mainFlex *tview.Flex

	app *tview.Application
}

func NewUI(cfg *archercl.AclNode, net *Network) *UI {
	ui := &UI{
		net: net,
	}

	ui.messageView = tview.NewTextView().
		SetDynamicColors(true)

	showLog := cfg.ChildAsBool("log")
	if showLog {
		ui.logText = tview.NewTextView().
			SetDynamicColors(true)
	}

	ui.statusText = tview.NewTextView().
		SetScrollable(false).
		SetTextAlign(tview.AlignRight)
	ui.statusText.SetBackgroundColor(tcell.ColorRebeccaPurple)

	ui.inputField = tview.NewInputField().
		SetLabel(">").
		SetFieldBackgroundColor(tcell.ColorBlack)

	ui.inputField.SetDoneFunc(func(key tcell.Key) {
		txt := ui.inputField.GetText()
		if key == tcell.KeyEnter {
			ui.net.SendText(txt)
			// No local echo
			//ui.addLocalMessage(txt)
		}

		// Tabs also come in here so don't clear those!
		if key == tcell.KeyTAB {
			return
		}

		// Everything else we do clear though
		ui.inputField.SetText("")
	})

	ui.mainFlex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(ui.messageView, 0, 3, false)

	if showLog {
		ui.mainFlex.AddItem(ui.logText, 0, 1, false)
	}

	ui.mainFlex.
		AddItem(ui.statusText, 2, 0, false).
		AddItem(ui.inputField, 1, 0, true)


	// Bind to the net refresh
	net.RefreshUI = func() {
		ui.Refresh()
	}
	net.RecvMessage = func(m *Message) {
		ui.RecvMessage(m)
	}

	return ui
}

func (ui *UI) Run() {
	log.Info("Running now....")

	ui.app = tview.NewApplication().SetRoot(ui.mainFlex, true)

	// Keep it fresh from the start...
	ui.Refresh()

	if err := ui.app.Run(); err != nil {
		panic(err)
	}
}

func (ui *UI) Refresh() {
	//log.Info("Refresh")


	status := "Disconnected"
	if ui.net.err != nil {
		status = "ERROR"
	} else if ui.net.c.IsConnectionOpen() {
		status = "Connected"
	}

	statusLine := fmt.Sprintf("id:%v\nname:%v mqtt:%v", ui.net.id, ui.net.name, status)

	go func() {
		ui.app.QueueUpdateDraw(func() {
			ui.statusText.SetText(statusLine)
		})
	}()
}

func (ui *UI) RecvMessage(msg *Message) {
	if len(msg.From) == 0 {
		msg.From = "Anon"
	}

	// TODO: Implement some sort of disk cache to persist across restarts

	// For now just let the TextView buffer handle it
	color := "cyan"
	if msg.me {
		color = "red"
	}
	fmt.Fprintf(ui.messageView, "\n[%v]%v[lightgray]: [blue]%v", color, msg.From, msg.Msg)

	// TODO: Clean up that buffer sometimes!

	ui.app.Draw()
}


//
//func (ui *UI) addLocalMessage(msg string) {
//	fmt.Fprintf(ui.messageView, "\n[green]me [lightgray]:[blue]%v", msg)
//
//	ui.app.Draw()
//}

func (ui *UI) Log(level logging.Level, calldepth int, rec *logging.Record) error {

	if ui.logText != nil {
		fmt.Fprintf(ui.logText, "%v\n", rec.Formatted(0))
	}

	//msg := Message{
	//	"level":     level.String(),
	//	"id":        rec.ID,
	//	"timestamp": rec.Time.UTC().Format("2006-01-02T15:04:05.999999Z"),
	//	"module":    rec.Module,
	//	"msg":       rec.Formatted(calldepth + 1),
	//}
	//
	//return c.Send(msg)

	return nil
}