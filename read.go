package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/docopt/docopt-go"
	"github.com/fiatjaf/go-nostr"
	terminaldimensions "github.com/wayneashleyberry/terminal-dimensions"
)

func read(opts docopt.Opts) {
	article := strings.ToLower(opts["<article>"].(string))
	if article == "" {
		log.Println("Please provide an article name.")
		return
	}
	initNostr()

	sub := pool.Sub(nostr.Filters{
		{
			Kinds: []int{KIND_WIKI},
			Tags:  nostr.TagMap{"w": []string{article}},
		},
	})

	fmt.Printf("Select version to read: \n")

	w, _ := terminaldimensions.Width()
	if w == 0 {
		w = 75
	}

	events := make([]*nostr.Event, 0, 5)
	go func() {
		i := 0
		for {
			evt := <-sub.UniqueEvents
			events = append(events, &evt)
			fmt.Printf("(%d) %s at %s: %s\n",
				i+1,
				shortenKey(evt.PubKey),
				evt.CreatedAt.Format("Jan 2 15:04"),
				strings.TrimSpace(shortenText(evt.Content, int(w-30))),
			)
			i++
		}
	}()

	for {
		scanner := bufio.NewScanner(os.Stdin)
		ok := scanner.Scan()
		if ok {
			input := strings.TrimRight(scanner.Text(), "\r\n")
			choice, _ := strconv.Atoi(input)
			if choice > 0 && choice <= len(events) {
				evt := events[choice-1]

				pager := os.Getenv("PAGER")
				cmd := exec.Command(pager)
				cmd.Stdin = bytes.NewBufferString(fmt.Sprintf(`
%s
===

authored by: %s
at %s

---

%s

`,
					strings.ToUpper(article),
					evt.PubKey,
					evt.CreatedAt.Format("Jan 2 15:04"),
					evt.Content,
				))
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				err := cmd.Run()
				if err != nil {
					output, _ := cmd.CombinedOutput()
					log.Printf(string(output))
					log.Printf("Failed to open pager (%s): %s\n", pager, err.Error())
				}
			}
		}

		fmt.Printf("Select version to read: ")
	}
}
