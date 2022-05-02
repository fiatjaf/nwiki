package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/fiatjaf/go-nostr"
	"github.com/jroimartin/gocui"
)

var colorNormal = color.New(color.FgWhite)
var colorSelected = color.New(color.FgBlack).Add(color.BgCyan).Add(color.Bold)

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
		events = append(events, &evt)

		sortVersions()
		// removeOldFromSameAuthor() TODO
		renderVersions(g)
		renderContent(g)
	}
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

func selectVersion(g *gocui.Gui, v *gocui.View) error {
	tmp, err := os.CreateTemp(os.TempDir(), "nwiki")
	if err != nil {
		log.Println("Failed to create temporary file: ", err.Error())
		return err
	}
	if _, err := tmp.WriteString(events[selected].Content); err != nil {
		log.Println("Failed to write temporary file: ", err.Error())
		return err
	}
	tmpName := tmp.Name()

	var editor string
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		if _, err := os.Open("/usr/bin/editor"); err == nil {
			editor = "/usr/bin/editor"
		}
	}
	if editor == "" {
		editor = "/usr/bin/vi"
	}
	tmp.Close()
	cmd := exec.Command(editor, tmpName)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		output, _ := cmd.CombinedOutput()
		log.Printf(string(output))
		log.Printf("Failed to wait editor (%s): %s\n", editor, err.Error())
	}

	tmp, err = os.Open(tmpName)
	if err != nil {
		log.Println("Failed to open file after editing: ", err.Error())
		return err
	}
	defer tmp.Close()
	data, err := ioutil.ReadAll(tmp)
	if err != nil {
		log.Println("Failed to read file contents after editing: ", err.Error())
		return err
	}
	content := string(data)

	unpauser := make(chan struct{})
	go func() {
		// do nothing if empty
		isEmpty := true
		for _, line := range strings.Split(content, "\n") {
			if strings.TrimSpace(line) != "" {
				isEmpty = false
				break
			}
		}
		if isEmpty {
			fmt.Println("Empty article. Doing nothing.")
			time.Sleep(2 * time.Second)
		} else {
			// publish article
			if evt, status, err := pool.PublishEvent(&nostr.Event{
				Content:   content,
				CreatedAt: time.Now(),
				Tags:      nostr.Tags{[]string{"w", article}},
				Kind:      KIND_WIKI,
			}); err != nil {
				fmt.Printf("Error publishing: %s.\n", err.Error())
				time.Sleep(2 * time.Second)
			} else {
				fmt.Printf("Event %s sent.\n", evt.ID)
				timeout := time.After(3 * time.Second)
				for {
					select {
					case s := <-status:
						fmt.Printf("  - %s: %s\n", s.Relay, s.Status)
					case <-timeout:
						goto unpause
					}
				}
			}
		}

	unpause:
		unpauser <- struct{}{}
	}()

	return PauseMainLoop{unpauser}
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
				shortenKey(evt.PubKey),
				evt.CreatedAt.Format("Jan 02 15:04"),
				strings.TrimSpace(shortenText(evt.Content, viewX-28)),
			)
			if err != nil {
				return err
			}
		}

		return nil
	})
}
