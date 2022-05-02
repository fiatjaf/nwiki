package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/docopt/docopt-go"
	"github.com/fiatjaf/go-nostr"
)

func edit(opts docopt.Opts) {
	if config.PrivateKey == "" {
		log.Printf("Can't publish. Private key not set.\n")
		return
	}
	article := strings.ToLower(opts["<article>"].(string))
	if article == "" {
		log.Println("Please provide an article name.")
		return
	}
	initNostr()

	// try to grab our latest article for this topic
	sub := pool.Sub(nostr.Filters{
		{
			Kinds:   []int{KIND_WIKI},
			Authors: []string{article},
			Tags:    nostr.TagMap{"w": []string{article}},
		},
	})
	newest := nostr.Event{}
	timeout := time.After(3 * time.Second)
	for {
		select {
		case evt := <-sub.UniqueEvents:
			if evt.CreatedAt.After(newest.CreatedAt) {
				newest = evt
			}
		case <-timeout:
			goto edit
		}
	}

	// actually edit by invoking an editor
edit:
	tmp, err := os.CreateTemp(os.TempDir(), "nwiki")
	if err != nil {
		log.Println("Failed to create temporary file: ", err.Error())
		return
	}
	if _, err := tmp.WriteString(newest.Content); err != nil {
		log.Println("Failed to write temporary file: ", err.Error())
		return
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
		return
	}
	defer tmp.Close()
	content, err := ioutil.ReadAll(tmp)
	if err != nil {
		log.Println("Failed to read file contents after editing: ", err.Error())
		return
	}

	// publish article
	if evt, status, err := pool.PublishEvent(&nostr.Event{
		Content:   string(content),
		CreatedAt: time.Now(),
		Tags:      nostr.Tags{[]string{"w", article}},
		Kind:      KIND_WIKI,
	}); err != nil {
		log.Printf("Error publishing: %s.\n", err.Error())
		return
	} else {
		fmt.Printf("Event %s sent.\n", evt.ID)
		for s := range status {
			fmt.Printf("  - %s: %s\n", s.Relay, s.Status)
		}
	}
}
