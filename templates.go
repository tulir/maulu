package main

import (
	"html/template"
	log "maunium.net/go/maulogger"
	"os"
)

var templIndex, templRedirect *template.Template

func loadTemplates() {
	log.Infoln("Loading HTML templates...")
	var err error
	templIndex, err = template.ParseFiles(config.Files.HTMLDirectory + "index.html")
	if err != nil {
		log.Fatalf("Failed to load index page: %s", err)
		os.Exit(3)
	}
	templRedirect, err = template.ParseFiles(config.Files.RedirectTemplate)
	if err != nil {
		log.Fatalf("Failed to load HTML/JS redirect page: %s", err)
		os.Exit(3)
	}
	log.Debugln("Successfully loaded HTML templates")
}
