package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/jroimartin/gocui"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip05"
)

var (
	pubkeys    []string
	nip05Cache = sync.Map{} // { [identifier]: boolean }
)

func gatherNames(g *gocui.Gui) {
	for _, evt := range events {
		if stringExists(evt.PubKey, pubkeys) {
			// do not update the subscription with this pubkey
			// as we're already listening for it
			continue
		}

		ctx, cancel := context.WithTimeout(ctx, time.Second*3)
		evt := pool.QuerySingle(ctx, config.Relays, nostr.Filter{Kinds: []int{0}, Authors: []string{evt.PubKey}})
		cancel()

		iexisting, exists := metadata.Load(evt.PubKey)
		if !exists || evt.CreatedAt > iexisting.CreatedAt {
			metadata.Store(evt.PubKey, evt)
			renderVersions(g)
			renderContent(g)
		}
	}
}

func nameFromMetadataEvent(ctx context.Context, event *nostr.Event) string {
	var data struct {
		Name  string `json:"name"`
		Nip05 string `json:"nip05"`
	}

	json.Unmarshal([]byte(event.Content), &data)

	if data.Nip05 != "" {
		if valid, cached := nip05Cache.Load(data.Nip05); !cached {
			// if not cached query HTTPS
			result, err := nip05.QueryIdentifier(ctx, data.Nip05)
			var valid bool
			if err == nil {
				// will be true if this pubkey matches the one from the event
				valid = result.PublicKey == event.PubKey
			} else {
				valid = false
			}

			// cache it
			nip05Cache.Store(data.Nip05, valid)

			if valid {
				return nip05.NormalizeIdentifier(data.Nip05)
			}
		} else {
			// otherwise reuse the result from the previous call
			if valid.(bool) {
				return nip05.NormalizeIdentifier(data.Nip05)
			} else {
				// nip05 invalid, let's use the normal name
			}
		}
	}

	if data.Name != "" {
		return fmt.Sprintf("\"%s\"", data.Name)
	}

	return ""
}

func authorName(ctx context.Context, pubkey string) string {
	if metadataEvent, ok := metadata.Load(pubkey); ok {
		name := nameFromMetadataEvent(ctx, metadataEvent)
		if name != "" {
			return name
		}
	}

	return shortenKey(pubkey)
}
