package main

import (
	"bufio"
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
		if args[0] == "short" {
			data.DeleteShort(args[1])
		} else if args[0] == "url" {
			data.DeleteURL(args[1])
		}
	} else if command == "set" && len(args) > 2 {
		data.InsertDirect(args[0], args[1], args[2])
	}
}
