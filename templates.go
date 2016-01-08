package main

import (
	"html/template"
	log "maunium.net/go/maulogger"
	"os"
)

var index, redirjs, redirhtml *template.Template

func loadTemplates() {
	log.Infoln("Loading HTML templates...")
	var err error
	index, err = template.ParseFiles("index.html")
	if err != nil {
		log.Fatalf("Failed to load index page: %s", err)
		os.Exit(3)
	}
	redirhtml, err = template.ParseFiles("redirect-html.html")
	if err != nil {
		log.Fatalf("Failed to load HTML redirect page: %s", err)
		os.Exit(3)
	}
	redirjs, err = template.ParseFiles("redirect-js.html")
	if err != nil {
		log.Fatalf("Failed to load JavaScrip redirect page: %s", err)
		os.Exit(3)
	}
	log.Debugln("Successfully loaded HTML templates")
}
