package main

import (
	"html/template"
	log "maunium.net/go/maulogger"
	"os"
)

var templIndex, templRedirect *template.Template

func loadTemplates() {
	var err error
	templRedirect, err = template.ParseFiles(config.RedirectTemplate)
	if err != nil {
		log.Fatalf("Failed to load HTML/JS redirect template: %s", err)
		os.Exit(3)
	}
	log.Debugln("Successfully loaded HTML/JS redirect template")
}
