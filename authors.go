package main

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/fiatjaf/go-nostr"
	"github.com/fiatjaf/go-nostr/nip05"
	"github.com/jroimartin/gocui"
)

var pubkeys []string
var namesSub *nostr.Subscription
var nip05Cache = sync.Map{} // { [identifier]: boolean }

func gatherNames(g *gocui.Gui) {
	for _, evt := range events {
		if stringExists(evt.PubKey, pubkeys) {
			// do not update the subscription with this pubkey
			// as we're already listening for it
			continue
		}

		pubkeys = append(pubkeys, evt.PubKey)
		filters := nostr.Filters{
			{
				Kinds:   []int{nostr.KindSetMetadata},
				Authors: pubkeys,
			},
		}

		if namesSub != nil {
			namesSub.Sub(filters)
		} else {
			namesSub = pool.Sub(filters)
		}
	}

	go listenForNames(g)
}

func listenForNames(g *gocui.Gui) {
	for evt := range namesSub.UniqueEvents {
		iexisting, exists := metadata.Load(evt.PubKey)
		if !exists || evt.CreatedAt.After(iexisting.(*nostr.Event).CreatedAt) {
			metadata.Store(evt.PubKey, &evt)
			renderVersions(g)
			renderContent(g)
		}
	}
}

func nameFromMetadataEvent(event *nostr.Event) string {
	var data struct {
		Name  string `json:"name"`
		Nip05 string `json:"nip05"`
	}

	json.Unmarshal([]byte(event.Content), &data)

	if data.Nip05 != "" {
		if valid, cached := nip05Cache.Load(data.Nip05); !cached {
			// if not cached query HTTPS
			result := nip05.QueryIdentifier(data.Nip05)

			// will be true if this pubkey matches the one from the event
			valid := result == event.PubKey

			// cache it
			nip05Cache.Store(data.Nip05, valid)

			return nip05.NormalizeIdentifier(data.Nip05)
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

func authorName(pubkey string) string {
	if metadataEvent, ok := metadata.Load(pubkey); ok {
		name := nameFromMetadataEvent(metadataEvent.(*nostr.Event))
		if name != "" {
			return name
		}
	}

	return shortenKey(pubkey)
}
