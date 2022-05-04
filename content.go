package main

import (
	"strings"

	"github.com/fatih/color"
	"github.com/jroimartin/gocui"
)

var titleColor = color.New(color.Bold).Add(color.FgCyan)
var metaColor = color.New(color.FgYellow)
var textColor = color.New(color.FgWhite)
var separatorColor = color.New(color.FgMagenta)

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

		evt := events[selected]
		v.Title = evt.ID

		titleSeparator := ""
		for i := 0; i < len(article); i++ {
			titleSeparator += "="
		}

		titleColor.Fprint(v, strings.ToUpper(article)+"\n")
		separatorColor.Fprint(v, titleSeparator)
		metaColor.Fprintf(v, "\n\nauthored by: %s\nat %s\n",
			evt.PubKey,
			evt.CreatedAt.Format("Jan 2 15:04"),
		)
		separatorColor.Fprint(v, "\n---\n")
		textColor.Fprint(v, "\n"+evt.Content)

		return nil
	})
}
