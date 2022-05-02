package main

import (
	"strings"
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
