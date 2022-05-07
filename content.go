package main

import (
	"fmt"
	"strings"

	"github.com/jroimartin/gocui"
)

func renderContent(g *gocui.Gui) {
	g.Update(func(g *gocui.Gui) error {
		v, err := g.View(VIEW_CONTENT)
		if err != nil {
			return err
		}

		if selected >= len(events) {
			return nil
		}

		v.Clear()
		v.SetOrigin(0, 0)

		evt := events[selected]
		v.Title = evt.ID

		titleSeparator := ""
		for i := 0; i < len(article); i++ {
			titleSeparator += "="
		}

		author := evt.PubKey
		name := authorName(evt.PubKey)
		if name != shortenKey(evt.PubKey) {
			author += fmt.Sprintf(" (%s)", name)
		}

		titleColor.Fprint(v, strings.ToUpper(article)+"\n")
		separatorColor.Fprint(v, titleSeparator)
		metaColor.Fprintf(v, "\n\nauthored by: %s\nat %s\n",
			author,
			evt.CreatedAt.Format("Jan 2 15:04"),
		)
		separatorColor.Fprint(v, "\n---\n")
		textColor.Fprint(v, "\n"+evt.Content)

		return nil
	})
}

func scroll(direction int) func(*gocui.Gui, *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		v, err := g.View(VIEW_CONTENT)
		if err != nil {
			return err
		}

		_, y := v.Size()
		l := len(v.BufferLines())
		_, o := v.Origin()
		newOrigin := o + direction*5
		if newOrigin < 0 {
			newOrigin = 0
		} else if newOrigin > l-y {
			newOrigin = l - y
		}

		return v.SetOrigin(0, newOrigin)
	}
}
