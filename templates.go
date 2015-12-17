package main

import (
	"html/template"
	log "maunium.net/go/maulogger"
	"os"
)

var index *template.Template

func loadTemplates() {
	var err error
	index, err = template.ParseFiles("index.html")
	if err != nil {
		log.Fatalf("Failed to load index page: %s", err)
		os.Exit(3)
	}
}
