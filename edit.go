package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/jroimartin/gocui"
	"github.com/nbd-wtf/go-nostr"
)

func selectVersionToEdit(g *gocui.Gui, v *gocui.View) error {
	tmp, err := os.CreateTemp(os.TempDir(), "nwiki")
	if err != nil {
		logToView(g, "Failed to create temporary file: %s", err.Error())
		return nil
	}

	var content string
	if selected < len(events) {
		content = events[selected].Content
	}

	if _, err := tmp.WriteString(content); err != nil {
		logToView(g, "Failed to write temporary file: %s", err.Error())
		return nil
	}

	unpauser := make(chan struct{})
	go callExternalEditorAndPublish(tmp, content, unpauser)

	return PauseMainLoop{unpauser}
}

func callExternalEditorAndPublish(tmp *os.File, content string, unpauser chan struct{}) {
	// open local editor to edit
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

	if err := cmd.Run(); err != nil {
		output, _ := cmd.CombinedOutput()
		log.Printf(string(output))
		log.Printf("Failed to run editor (%s): %s\n", editor, err.Error())
		return
	}

	tmp, err := os.Open(tmpName)
	if err != nil {
		log.Println("Failed to open file after editing: ", err.Error())
		return
	}
	defer tmp.Close()
	data, err := ioutil.ReadAll(tmp)
	if err != nil {
		log.Println("Failed to read file contents after editing: ", err.Error())
		return
	}
	newContent := normalizeContent(string(data))

	// do nothing if empty or unchanged
	isEmpty := true
	for _, line := range strings.Split(newContent, "\n") {
		if strings.TrimSpace(line) != "" {
			isEmpty = false
			break
		}
	}

	if isEmpty {
		queueLogToView("Empty content. Won't publish.")
		return
	}

	if newContent == content {
		queueLogToView("Unchanged content. Won't publish.")
		return
	}

	evt := nostr.Event{
		Content:   newContent,
		CreatedAt: nostr.Now(),
		Tags: nostr.Tags{
			nostr.Tag{"d", strings.ToLower(article)},
		},
		Kind: KIND_WIKI,
	}
	if err := evt.Sign(config.PrivateKey); err != nil {
		queueLogToView("failed to sign event: %s", err)
		return
	}

	for _, url := range config.Relays {
		r, err := pool.EnsureRelay(url)
		if err != nil {
			queueLogToView("failed to connect to '%s': %s", url, err)
			continue
		}

		if _, err := r.Publish(ctx, evt); err != nil {
			queueLogToView("error publishing: %s", err.Error())
			time.Sleep(2 * time.Second)
		} else {
			queueLogToView("event %s sent.", evt.ID)
		}
	}
}
