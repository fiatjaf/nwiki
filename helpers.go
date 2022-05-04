package main

import (
	"strings"

	"github.com/fiatjaf/go-nostr"
)

func shortenKey(key string) string {
	return key[0:4] + "…" + key[len(key)-3:]
}

func shortenText(text string, maxChars int) string {
	if len(text) < maxChars {
		return text
	}

	idx := strings.LastIndexAny(text[0:maxChars-1], " .,?!")
	if idx != -1 {
		return text[0:idx] + "…"
	}

	return text[0:maxChars-1] + "…"
}

func getMatchingPubKey(pubkey string, events []*nostr.Event) (int, *nostr.Event) {
	for i, evt := range events {
		if evt.PubKey == pubkey {
			return i, evt
		}
	}

	return -1, nil
}

func stringExists(needle string, haystack []string) bool {
	for _, hay := range haystack {
		if hay == needle {
			return true
		}
	}

	return false
}
