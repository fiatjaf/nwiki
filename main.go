package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/docopt/docopt-go"
	"github.com/mitchellh/go-homedir"
)

const USAGE = `nwiki

Usage:
  nwiki <article>
  nwiki edit <article>
`

const (
	KIND_WIKI       = 17
	KIND_REPUTATION = 16
)

func main() {
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

	// parse args
	opts, err := docopt.ParseDoc(USAGE)
	if err != nil {
		return
	}

	switch {
	case opts["edit"].(bool):
		edit(opts)
	default:
		read(opts)
	}
}
