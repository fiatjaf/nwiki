package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/fiatjaf/go-nostr"
	"github.com/jroimartin/gocui"
	"github.com/mitchellh/go-homedir"
)

const (
	KIND_WIKI       = 17
	KIND_REPUTATION = 16

	VIEW_VERSIONS = "versions"
	VIEW_CONTENT  = "content"
)

var (
	article  string
	events   []*nostr.Event
	selected = 0
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
	if err := g.SetKeybinding(VIEW_VERSIONS, gocui.KeyEnter, gocui.ModNone, selectVersion); err != nil {
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
	if v, err := g.SetView(VIEW_VERSIONS, 0, 0, maxX/3, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		fmt.Fprint(v, "loading available articles...")
		v.Title = article
	}
	if v, err := g.SetView(VIEW_CONTENT, maxX/3+1, 0, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		v.Editor = gocui.DefaultEditor
		v.Editable = true
	}
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
