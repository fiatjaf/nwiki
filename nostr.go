package main

import (
	"github.com/nbd-wtf/go-nostr"
)

var pool *nostr.SimplePool

func initNostr() {
	pool = nostr.NewSimplePool(ctx)
}
