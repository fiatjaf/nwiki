package main

import (
	"fmt"

	"github.com/docopt/docopt-go"
)

func addRelay(opts docopt.Opts) {
	addr := opts["<url>"].(string)
	config.Relays[addr] = Policy{
		Read:  true,
		Write: true,
	}
	fmt.Printf("Added relay %s.\n", addr)
}

func removeRelay(opts docopt.Opts) {
	addr := opts["<url>"].(string)
	delete(config.Relays, addr)
	fmt.Printf("Removed relay %s.\n", addr)
}

func recommendRelay(opts docopt.Opts) {
	addr := opts["<url>"].(string)

	// TODO

	fmt.Printf("Published a relay recommendation for %s.", addr)
}

func listRelays(opts docopt.Opts) {
	for relay, policy := range config.Relays {
		fmt.Printf("%s: %s\n", relay, policy)
	}
}
