package main

import (
	"flag"
	"fmt"
	"github.com/eyethereal/go-archercl"
	"github.com/gdamore/tcell/v2"
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
	net         *Network
	ss          *SysStat
	messageView *tview.TextView
	inputField  *tview.InputField

	logText *tview.TextView

	batteryLevel *tview.TextView
	statusText   *tview.TextView
	statusFlex   *tview.Flex

	mainFlex *tview.Flex

	app *tview.Application
}

func NewUI(cfg *archercl.AclNode, net *Network, ss *SysStat) *UI {
	ui := &UI{
		net: net,
		ss:  ss,
	}

	// The big message area at top
	ui.messageView = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true)

	// Maybe have a log area in there
	showLog := cfg.ChildAsBool("log")
	if showLog {
		ui.logText = tview.NewTextView().
			SetDynamicColors(true).
			SetScrollable(true)
	}

	/////// Status things
	ui.batteryLevel = tview.NewTextView().
		SetScrollable(false)
	ui.batteryLevel.
		SetBackgroundColor(tcell.ColorBlue)

	ui.statusText = tview.NewTextView().
		SetScrollable(false).
		SetTextAlign(tview.AlignRight)
	ui.statusText.SetBackgroundColor(tcell.ColorRebeccaPurple)

	ui.statusFlex = tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(ui.batteryLevel, 4, 0, false).
		AddItem(ui.statusText, 0, 4, false)

	// And then a simple input field for the bottom of the screen
	ui.inputField = tview.NewInputField().
		SetLabel(fmt.Sprintf("%s> ", ui.net.name)).
		SetFieldBackgroundColor(tcell.ColorBlack)

	// What to do when the user hits enter or leaves the field
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
		AddItem(ui.statusFlex, 1, 0, false).
		AddItem(ui.inputField, 1, 0, true)

	// For certain keys we want them to ALWAYS go to the messageView
	ui.mainFlex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		key := event.Key()
		//log.Infof("Got key %v", key)

		switch key {
		case tcell.KeyPgUp:
			row, col := ui.messageView.GetScrollOffset()
			_, _, _, h := ui.messageView.GetInnerRect()
			row -= h
			if row < 0 {
				row = 0
			}

			//log.Infof("Scrolling up by %v to %v", h, row)
			ui.messageView.ScrollTo(row, col)
			event = nil

		case tcell.KeyPgDn:
			row, col := ui.messageView.GetScrollOffset()
			_, _, _, h := ui.messageView.GetInnerRect()
			row += h
			//if row < 0 {
			//	row = 0
			//}

			//log.Infof("Scrolling down by %v to %v", h, row)
			ui.messageView.ScrollTo(row, col)
			event = nil
		}

		return event
	})

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

	// Start some go routines to update parts of the UI
	go ui.batteryUpdater()

	if err := ui.app.Run(); err != nil {
		panic(err)
	}
}

func (ui *UI) batteryUpdater() {
	for {
		percent := <-ui.ss.PowerPercent

		log.Infof("batteryUpdater got %v", percent)
		s := fmt.Sprintf("%v%%", percent)
		log.Infof("Will set text to %v", s)
		go func() {
			ui.app.QueueUpdateDraw(func() {
				ui.batteryLevel.SetText(s)
			})
		}()
	}
}

func (ui *UI) Refresh() {
	//log.Info("Refresh")

	status := "Disconnected"
	if ui.net.err != nil {
		status = "ERROR"
	} else if ui.net.IsConnected() {
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
	color := "green"
	if msg.me {
		color = "yellow"
	}
	_, err := fmt.Fprintf(ui.messageView, "[%v]%v[lightgray]: %v\n[-:-:-]", color, msg.From, msg.Msg)
	if err != nil {
		log.Errorf("Fprintf Err: %v", err)
	}

	// TODO: Clean up that buffer sometimes!

	// Because RecvMessage isn't on the Main routine we have to explicitly
	// tell the UI that we want an update on the next chance. Note that per
	// the UI library documentation it is NOT safe to call Draw() from the
	// main routine, which we didn't know and had it sprinkled all over.
	// Thus - deadlocks. Oops! :)
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

		case "/nick":
			ui.cmdNick(words)

		default:
			ui.cmdUnknown(words)
		}

		ui.messageView.ScrollToEnd()
	} else {
		// Computers are fast. This is dumb but it gets us easy
		// access to the go syntax for double quoted strings like \u2800 etc.
		// This does mean that users can send \r \n or whatever. I think
		// we have things setup where it will all get interpreted as character
		// data though so the terminal should be ok

		// If we don't do the first replacement of double quotes they won't
		// make it through the unquoting
		dql := strings.Replace(line, "\"", "\\\"", -1)
		q := fmt.Sprintf("\"%s\"", dql)
		uq, e := strconv.Unquote(q)
		if e != nil {
			log.Error(e)
			return
		}

		// Do this on a co-routine because it might fail
		go ui.sendText(uq)
	}
}

// sendText will send a string as a message. It expects to be invoked on a co-routine and
// not directly running on the main routine. This is important because when things go south
// it is going to call ui.app.Draw() which MUST NOT happen on the main routine or
// nasty deadlocks occur. So don't do that. Always call this on a coroutine or as
// a coroutine!
func (ui *UI) sendText(unquoted string) {
	if ui.net == nil {
		ui.sysText("ERROR: No network object. Did not send message")
		ui.app.Draw()
		return
	}

	err := ui.net.SendText(nil, unquoted)
	if err != nil {
		errStr := fmt.Sprintf("ERROR: Failed to send msg (%v): %v", unquoted, err)
		ui.sysText(errStr)
		ui.app.Draw()
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
	repeat := lorem.Int("r", 1, "Number of times to repeat")

	err := lorem.Parse(words[1:])
	if err != nil {
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
		ui.sysText("invalid kind")
		return
	}

	if *repeat < 1 {
		*repeat = 1
	}
	for i := 0; i < *repeat; i++ {

		if *chars == 0 || *chars < 0 {
			// The whole thing
			ui.sendText(src)
			continue
		}

		if *chars <= len(src) {
			ui.sendText(src[:*chars])
		} else {
			ui.sysText(fmt.Sprintf("only have %v chars of that", len(src)))
		}
	}
}

func (ui *UI) cmdChars(words []string) {
	cmd := flag.NewFlagSet("/chars", flag.ContinueOnError)
	cmd.SetOutput(ui.messageView)
	start := cmd.Int("start", 32, "start code")
	end := cmd.Int("end", 127, "end code")

	err := cmd.Parse(words[1:])
	if err != nil {
		ui.sysText("Error")
		return
	}

	s := *start
	e := *end
	log.Debugf("s=%v e=%v", s, e)

	// TODO: Make this much sexier in some sort of table or perhaps a popover window thing?

	amt := e - s
	out := make([]rune, amt)

	for i := 0; i < amt; i++ {
		out[i] = rune(i + s)
	}

	ui.sysText(string(out))
}

const DEFAULT_HELP = `
Enter text and press return. If the text begins with a / it is interpreted as a command. Everything else is sent as a chat message. Your messages will be prepended by your current nickname in yellow. Messages from others will show their nickname in cyan.

The core commands are: [red]help[magenta], [red]lorem[magenta], [red]chars[magenta], [red]nick[magenta]

There may be other commands and features not documented in the help system. Additional help for each command is obtained by using the command [red]/help {name}[magenta] where {name} is the name of a command.

Some commands take additional parameters specified using attributes demarked with leading - characters followed by an attribute value. See each command help page for specifics.
`

var HelpText = map[string]string{
	"lorem": `Lorem Help`,

	"chars": `Displays a list of characters between two numerical values which are specified using the -start and -end parameters.

Example: /chars -start 45 -end 83

Why? Because the messages support unicode and can contain escape codes as follows:

  \x  followed by exactly 2 hex digits
  \   followed by exactly 3 octal digits
  \u  followed by exactly 4 hex digits
  \U  followed by exactly 8 hex digits
`,

	"colors": `Colors are specified in messages using square brackets such as [red[]. They can be specified using names or a # followed by a 24 bit (6 digit) hex string such as [#8080ff[]. 

Foreground, background, and flags can also be set. The full square bracket notation is [<foreground>:<background>:<flags>[]. Fields may be blank and trailing fields may be omitted. The character - in a field resets that to a default value.

The flags are: l (blink), b (bold), i (italic), d (dim), r (reverse), u (underline), s (strike-through)

To enter non-color bracketed text insert a left bracket immediately before the closing bracket as so [whatever[[]
`,
}

func (ui *UI) cmdHelp(words []string) {
	if len(words) < 2 {
		_, _ = fmt.Fprintf(ui.messageView, "[magenta]%v\n", DEFAULT_HELP)

		//ui.sysText(fmt.Sprintf("lorem chars help nick\nThat's all there is."))
	} else {
		text := HelpText[words[1]]

		if text == "" {
			ui.sysText(fmt.Sprintf("No help for [red]%v\n", words[1]))
		} else {
			_, _ = fmt.Fprintf(ui.messageView, "Help for [red]%v[magenta]\n", words[1])
			_, _ = fmt.Fprintf(ui.messageView, "\n[magenta]%v[-:-:-]", text)
		}
	}
}

//TODO: Add a cmdColorNames that shows all the color names and even how they look

func (ui *UI) cmdNick(words []string) {
	if len(words) > 1 {
		ui.net.name = words[1]
		ui.inputField.SetLabel(fmt.Sprintf("%s> ", ui.net.name))
	}
}

func (ui *UI) cmdUnknown(words []string) {
	ui.sysText(fmt.Sprintf("Unknown command %v", words[0]))
}

func (ui *UI) sysText(txt string) {
	_, err := fmt.Fprintf(ui.messageView, "[red]* [magenta]%v\n[-:-:-]", txt)
	if err != nil {
		log.Errorf("Failed to Fprintf sysText %v", err)
	}

	// The assumption is we are on the main routine and thus do not
	// call ui.app.Draw() explicitly. In the future we might want
	// to add a check for whether this is needed or not. It is not safe
	// to call from a callback (i.e. when handling a /command)
}

//
//func (ui *UI) addLocalMessage(msg string) {
//	fmt.Fprintf(ui.messageView, "\n[green]me [lightgray]:[blue]%v", msg)
//
//  Not safe from a callback!
//	ui.app.Draw()
//}

func (ui *UI) Log(level logging.Level, calldepth int, rec *logging.Record) error {

	_ = calldepth
	if ui.logText == nil {
		return nil
	}

	/*
		CRITICAL Level = iota
		ERROR
		WARNING
		NOTICE
		INFO
		DEBUG
	*/
	var prefix string
	switch level {
	case logging.DEBUG:
		prefix = "[cyan]"

	case logging.INFO:
		prefix = "[-:-:-]"

	case logging.NOTICE:
		prefix = "[green]"

	case logging.WARNING:
		prefix = "[yellow]"

	case logging.ERROR:
		prefix = "[red]"

	case logging.CRITICAL:
		prefix = "[magenta]"

	default:
		prefix = ""
	}

	_, _ = fmt.Fprintf(ui.logText, "%s%v\n", prefix, rec.Formatted(0))

	ui.logText.ScrollToEnd()

	return nil
}
