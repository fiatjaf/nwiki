package main

import (
	"sort"

	"github.com/fiatjaf/go-nostr"
	"github.com/jroimartin/gocui"
)

func listVersions(g *gocui.Gui, article string) {
	initNostr()

	sub := pool.Sub(nostr.Filters{
		{
			Kinds: []int{KIND_WIKI},
			Tags:  nostr.TagMap{"w": []string{article}},
		},
	})

	for {
		if len(events) == 0 {
			g.SetCurrentView(VIEW_VERSIONS)
		}

		evt := <-sub.UniqueEvents
		evt.Content = normalizeContent(evt.Content)
		events = append(events, &evt)

		sortVersions()
		removeOldFromSameAuthor(g)
		renderVersions(g)
		renderContent(g)
		gatherNames(g)
	}
}

func removeOldFromSameAuthor(g *gocui.Gui) {
	filteredEvents := make([]*nostr.Event, 0, len(events))
	for _, event := range events {
		f, found := getMatchingPubKey(event.PubKey, filteredEvents)

		if found == nil {
			// we didn't find any matching pubkey
			filteredEvents = append(filteredEvents, event)
		} else {
			// we got a matching pubkey, so either discard or replace
			if event.CreatedAt.After(found.CreatedAt) {
				filteredEvents[f] = event
			}
		}
	}

	events = filteredEvents
}

func sortVersions() {
	sort.Slice(events, func(i, j int) bool {
		return events[i].CreatedAt.After(events[j].CreatedAt)
	})
}

func moveSelection(incrSelected int) func(*gocui.Gui, *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		if len(events) == 0 {
			return nil
		}

		selected = (len(events) + selected + incrSelected) % len(events)

		renderVersions(g)
		renderContent(g)

		return nil
	}
}

func renderVersions(g *gocui.Gui) {
	g.Update(func(g *gocui.Gui) error {
		v, err := g.View(VIEW_VERSIONS)
		if err != nil {
			return err
		}

		v.Clear()

		viewX, _ := v.Size()

		for i, evt := range events {
			c := colorNormal
			if selected == i {
				c = colorSelected
			}

			_, err = c.Fprintf(v, "%s at %s: %s\n",
				authorName(evt.PubKey),
				evt.CreatedAt.Format("Jan 02 15:04"),
				shortenText(evt.Content, viewX-28),
			)
			if err != nil {
				return err
			}
		}

		return nil
	})
}
