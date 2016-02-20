package main

import (
	"bufio"
	log "maunium.net/go/maulogger"
	"maunium.net/go/maulu/data"
	"os"
	"strings"
)

func stdinListen() {
	reader := bufio.NewReader(os.Stdin)
	for {
		text, _ := reader.ReadString('\n')
		text = text[:len(text)-1]
		args := strings.Split(text, " ")

		onCommand(strings.ToLower(args[0]), args[1:])
	}
}

func onCommand(command string, args []string) {
	if command == "remove" && len(args) > 1 {
		var err error
		if args[0] == "short" {
			err = data.DeleteShort(args[1])
		} else if args[0] == "url" {
			err = data.DeleteURL(args[1])
		}
		if err != nil {
			log.Errorf("Failed to delete: %s", err)
		} else {
			log.Infof("Successfully deleted all entries where %s=%s", args[0], args[1])
		}
	} else if command == "set" && len(args) > 2 {
		err := data.InsertDirect(args[0], args[1], args[2])
		if err != nil {
			log.Errorf("Failed to insert: %s", err)
		} else {
			log.Infof("Successfully inserted entry: (%s, %s, %s)", args[0], args[1], args[2])
		}
	}
}
