package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/fiatjaf/go-nostr"
	"github.com/jroimartin/gocui"
	"github.com/mitchellh/go-homedir"
)

const (
	KIND_WIKI       = 17
	KIND_REPUTATION = 16

	VIEW_VERSIONS = "versions"
	VIEW_CONTENT  = "content"
	VIEW_CONTROL  = "control"
)

var (
	article        string
	events         []*nostr.Event
	selected       = 0
	metadata       = sync.Map{} // { [pubkey]: *nostr.Event }
	queuedMessages = make([]string, 0, 1)
)

var (
	instructionsColor = color.New(color.FgYellow)
	infoColor         = color.New(color.Bold).Add(color.FgBlue)
	titleColor        = color.New(color.Bold).Add(color.FgCyan)
	metaColor         = color.New(color.FgYellow)
	textColor         = color.New(color.FgWhite)
	separatorColor    = color.New(color.FgMagenta)
	colorNormal       = color.New(color.FgWhite)
	colorSelected     = color.New(color.FgBlack).Add(color.BgCyan).Add(color.Bold)
)

func main() {
	// args
	article = strings.ToLower(strings.Join(os.Args[1:], " "))
	if article == "" {
		fmt.Println("Please provide an article name.")
		return
	}

	// find datadir
	flag.StringVar(&config.DataDir, "datadir", "~/.config/nostr",
		"Base directory for configurations and data from Nostr.")
	flag.Parse()
	config.DataDir, _ = homedir.Expand(config.DataDir)
	os.Mkdir(config.DataDir, 0700)

	// logger config
	log.SetPrefix("<> ")

	// parse config
	path := filepath.Join(config.DataDir, "config.json")
	f, _ := os.Open(path)
	err := json.NewDecoder(f).Decode(&config)
	if err != nil {
		log.Fatal("can't parse config file " + path + ": " + err.Error())
		return
	}
	config.Init()

	// run main loop
	startMainLoop()
}

func startMainLoop() {
	// setup gocui
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Fatal(err)
	}
	defer g.Close()
	g.SetManagerFunc(layout)

	// reset in-memory events
	events = make([]*nostr.Event, 0, len(events))

	// query articles
	go listVersions(g, article)

	// set key bindings
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Fatal(err)
	}
	if err := g.SetKeybinding(VIEW_VERSIONS, gocui.KeyArrowUp, gocui.ModNone, moveSelection(-1)); err != nil {
		log.Fatal(err)
	}
	if err := g.SetKeybinding(VIEW_VERSIONS, gocui.KeyArrowDown, gocui.ModNone, moveSelection(1)); err != nil {
		log.Fatal(err)
	}
	if err := g.SetKeybinding(VIEW_VERSIONS, gocui.KeyEnter, gocui.ModNone, selectVersionToEdit); err != nil {
		log.Fatal(err)
	}

	if err := g.MainLoop(); err != nil {
		if pause, ok := err.(PauseMainLoop); ok {
			g.Close()
			<-pause.unpause
			startMainLoop()
		} else if err == gocui.ErrQuit {
			return
		} else {
			log.Fatal(err)
		}
	}
}

type PauseMainLoop struct {
	unpause chan struct{}
}

func (p PauseMainLoop) Error() string { return "pause-main-loop" }

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView(VIEW_VERSIONS, 0, 0, maxX/3, maxY*3/5); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		fmt.Fprint(v, "loading available articles...\n")
		v.Title = fmt.Sprintf("\"%s\" results", article)
	}
	if v, err := g.SetView(VIEW_CONTENT, maxX/3+1, 0, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		v.Autoscroll = false
		v.Wrap = true
	}
	if v, err := g.SetView(VIEW_CONTROL, 0, maxY*3/5, maxX/3, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		for _, message := range queuedMessages {
			fmt.Fprintln(v, message)
		}
		queuedMessages = make([]string, 0, 1)
		fmt.Fprintln(v, "")
		writeInitialMessages(v)

		v.Wrap = true
	}
	return nil
}

func logToView(g *gocui.Gui, fmessage string, args ...interface{}) {
	g.Update(func(g *gocui.Gui) error {
		v, err := g.View(VIEW_CONTROL)
		if err != nil {
			return err
		}

		contents, err := ioutil.ReadAll(v)
		if err != nil {
			return err
		}

		v.Clear()

		message := fmt.Sprintf(fmessage, args...)
		fmt.Fprintf(v, "- %s\n", message)

		idx := bytes.Index(contents, []byte("---"))
		if idx >= 0 {
			contents = contents[0:idx]
		}
		v.Write(contents)
		writeInitialMessages(v)

		return nil
	})
}

func queueLogToView(fmessage string, args ...interface{}) {
	queuedMessages = append(queuedMessages, fmt.Sprintf(fmessage, args...))
}

func writeInitialMessages(v *gocui.View) {
	separatorColor.Fprintln(v, "---")
	pubkey, _ := nostr.GetPublicKey(config.PrivateKey)
	infoColor.Fprintf(v, "pubkey: %v\n", pubkey)
	infoColor.Fprintf(v, "relays: %v\n", config.RelaysList())
	separatorColor.Fprintln(v, "---")
	instructionsColor.Fprintln(v, "> Use the arrow keys to select, Enter to edit on your local editor.")
	instructionsColor.Fprintln(v, "> If no articles are found, Enter will give you the chance to create a new one.")
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
