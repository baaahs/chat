package main

import (
	"flag"
	"fmt"
	"github.com/eyethereal/go-archercl"
	"github.com/gdamore/tcell"
	"github.com/op/go-logging"
	"github.com/rivo/tview"
	"strconv"
	"strings"
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
		SetLabel(fmt.Sprintf("%s> ", ui.net.name)).
		SetFieldBackgroundColor(tcell.ColorBlack)

	ui.inputField.SetDoneFunc(func(key tcell.Key) {
		txt := ui.inputField.GetText()

		switch key {
		case tcell.KeyEnter:
			ui.handleLine(txt)
			ui.inputField.SetText("")

		case tcell.KeyESC:
			ui.inputField.SetText("")

		default:
			log.Debugf("Unhandled done key %v", key)
		}
	})

	ui.mainFlex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(ui.messageView, 0, 3, false)

	if showLog {
		ui.mainFlex.AddItem(ui.logText, 0, 1, false)
	}

	ui.mainFlex.
		AddItem(ui.statusText, 1, 0, false).
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

	shortId := ui.net.id
	if len(shortId) > 9 {
		shortId = fmt.Sprintf("...%s", shortId[len(shortId)-6:])
	}
	statusLine := fmt.Sprintf("id:%v mqtt:%v", shortId, status)

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


func (ui *UI) handleLine(line string) {
	line = strings.TrimSpace(line)
	if len(line) == 0 {
		return
	}

	if line[0] == '/' {
		words := strings.Split(line, " ")
		switch words[0] {
		case "/lorem":
			ui.cmdLorem(words)

		case "/chars":
			ui.cmdChars(words)

		case "/help":
			ui.cmdHelp(words)

		default:
			ui.cmdUnknown(words)
		}
	} else {
		// Computers are fast. This is dumb but it gets us easy
		// access to the go syntax for double quoted strings like \u2800 etc.
		// This does mean that users can send \r \n or whatever. I think
		// we have things setup where it will all get interpreted as character
		// data though so the terminal should be ox
		q := fmt.Sprintf("\"%s\"", line)
		uq, e := strconv.Unquote(q)
		if e != nil {
			log.Error(e)
			return
		}
		ui.net.SendText(uq)
	}
}

const PIRATE = `Transom gangway Jolly Roger poop deck Cat o'nine tails sutler run a shot across the bow snow starboard mutiny. Interloper cable fathom smartly black jack transom draft chase weigh anchor splice the main brace. Scuppers topgallant rope's end landlubber or just lubber to go on account scuttle crack Jennys tea cup Privateer broadside booty.
Hands lee cog draft warp measured fer yer chains list cutlass jib quarterdeck. Plunder spyglass black spot scourge of the seven seas hardtack yawl Privateer marooned jury mast long boat. Swing the lead aye carouser heave down lugger transom rope's end strike colors rum provost.
Spike salmagundi scuttle Arr jack heave to lateen sail yo-ho-ho Letter of Marque code of conduct. Heave down scuppers lee dance the hempen jig hardtack scallywag hornswaggle killick hands walk the plank. Gally pirate yard gunwalls execution dock belay bilged on her anchor gangway Sink me Jack Ketch.`

const SAMUEL = `Now that there is the Tec-9, a crappy spray gun from South Miami. This gun is advertised as the most popular gun in American crime. Do you believe that shit? It actually says that in the little book that comes with it: the most popular gun in American crime. Like they're actually proud of that shit.
My money's in that office, right? If she start giving me some bullshit about it ain't there, and we got to go someplace else and get it, I'm gonna shoot you in the head then and there. Then I'm gonna shoot that bitch in the kneecaps, find out where my goddamn money is. She gonna tell me too. Hey, look at me when I'm talking to you, motherfucker. You listen: we go in there, and that nigga Winston or anybody else is in there, you the first motherfucker to get shot. You understand?
Your bones don't break, mine do. That's clear. Your cells react to bacteria and viruses differently than mine. You don't get sick, I do. That's also clear. But for some reason, you and I react the exact same way to water. We swallow it too fast, we choke. We get some in our lungs, we drown. However unreal it may seem, we are connected, you and I. We're on the same curve, just on opposite ends.`

const LIPSOM = `Lorem ipsum dolor sit amet, consectetur adipiscing elit. Aliquam id fringilla nisi. Fusce lacinia elit vel eros molestie consequat. Donec sollicitudin gravida quam, vel faucibus nibh molestie id. Suspendisse vel arcu eros. Duis non aliquet sapien. Nam et mi facilisis, aliquam risus vitae, dignissim arcu. Cras eu tempus ante. Nulla fermentum dui eu purus eleifend, ac aliquet erat lobortis.
Phasellus at feugiat justo. Vestibulum odio arcu, eleifend et lacinia id, condimentum eu tortor. Etiam nec risus nec ligula congue interdum. Aliquam iaculis egestas dignissim. Fusce eu mi vitae orci venenatis porttitor. Integer at fermentum nisl. Pellentesque id ligula eget quam consequat fringilla et a odio. Maecenas vitae dolor mattis, sollicitudin mauris vel, dictum turpis. Vestibulum eget nisi et sapien placerat bibendum sit amet tristique arcu. Vivamus porta velit laoreet dolor facilisis congue. Nam non libero mollis, mollis justo sit amet, varius ex.
Morbi eleifend nunc enim, vel molestie velit tempus aliquet. Maecenas viverra erat scelerisque, rutrum nibh et, sollicitudin tellus. In ac urna ut tortor accumsan blandit. Proin elit eros, tempor in purus nec, pharetra sagittis magna. Nam dictum ullamcorper iaculis. Pellentesque sit amet suscipit massa, at suscipit velit. Quisque consequat nibh ante, ut consequat sapien pulvinar ac. Etiam fringilla congue odio, et finibus lectus posuere quis. Vestibulum ante ipsum primis in faucibus orci luctus et ultrices posuere cubilia Curae; Etiam semper, enim a auctor maximus, mauris ligula fermentum risus, mattis vulputate ligula metus at massa.
`

func (ui *UI) cmdLorem(words []string) {

	lorem := flag.NewFlagSet("/lorem", flag.ContinueOnError)
	kind := lorem.String("k", "lipsom", "Kind of text (lipsom, pirate, samuel)")
	chars := lorem.Int("c", 0, "number of chars to print")
	lorem.SetOutput(ui.messageView)

	err := lorem.Parse(words[1:])
	if err != nil {
		ui.app.Draw()
		return
	}

	var src string
	switch *kind {
	case "lipsom":
		src = LIPSOM

	case "pirate":
		src = PIRATE

	case "samuel":
		src = SAMUEL

	default:
		ui.addRawText("invalid kind")
		return
	}

	if *chars == 0 || *chars < 0{
		// The whole thing
		ui.net.SendText(src)
		return
	}

	if *chars <= len(src) {
		ui.net.SendText(src[:*chars])
		return
	}

	ui.addRawText(fmt.Sprintf("only have %v chars of that", len(src)))
}

func (ui *UI) cmdChars(words []string) {
	cmd := flag.NewFlagSet("/chars", flag.ContinueOnError)
	cmd.SetOutput(ui.messageView)
	start := cmd.Int("start", 32, "start code")
	end := cmd.Int("end", 127, "end code")

	err := cmd.Parse(words[1:])
	if err != nil {
		ui.addRawText("Error")
		ui.app.Draw()
		return
	}

	s := *start
	e := *end
	log.Debugf("s=%v e=%v", s, e)

	amt := e - s
	out := make([]rune, amt)

	for i := 0; i < amt; i++ {
		out[i] = rune(i + s)
	}

	ui.addRawText(string(out))
}

func (ui *UI) cmdHelp(words []string) {
	ui.addRawText(fmt.Sprintf("There is no help"))
}

func (ui *UI) cmdUnknown(words []string) {
	ui.addRawText(fmt.Sprintf("Unknown command %v", words[0]))
}



func (ui *UI) addRawText(txt string) {
	fmt.Fprintf(ui.messageView, "\n[red]* [magenta]%v", txt)

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