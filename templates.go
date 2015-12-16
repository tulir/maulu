package main

import (
	"fmt"
	"html/template"
	log "maunium.net/go/maulogger"
	"net/http"
	"net/url"
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
func writeError(w http.ResponseWriter, err string, args ...interface{}) {
	w.Header().Add("Location", config.URL+"?error="+url.QueryEscape(fmt.Sprintf(err, args...)))
	w.WriteHeader(http.StatusFound)
}
