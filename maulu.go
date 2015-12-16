package main

import (
	_ "github.com/go-sql-driver/mysql"
	flag "github.com/ogier/pflag"
	"html/template"
	log "maunium.net/go/maulogger"
	"net/http"
	"os"
	"strconv"
)

func shorten(w http.ResponseWriter, r *http.Request) {

}

func get(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[1:]
	if path == "" {
		log.Debugf("%[1]s requested the index page", r.RemoteAddr)
		index.Execute(w, r.URL.Query().Get("url"))
	} else if path == "unshorten" {
		log.Debugf("%[1]s requested the unshortening page", r.RemoteAddr)
	} else if path == "favicon.ico" {
		log.Debugf("%[1]s requested the favicon", r.RemoteAddr)
	} else {
		log.Debugf("%[1]s requested long url of of %[2]s", r.RemoteAddr, path)
		url, err := queryURL(path)
		if err != nil {
			log.Errorf("Failed to find redirect from short url %[2]s: %[1]s", err, path)
			return
		}
		w.Header().Add("Location", url)
		w.WriteHeader(http.StatusFound)
	}
	/*data, err := ioutil.ReadFile(file)
	if err != nil {
		log.Errorf("Failed to read file %[1]s", file)
	}
	w.Write(data)*/
}

var index, unshorten *template.Template

func loadTemplates() {
	var err error
	index, err = template.ParseFiles(config.Pages.Index)
	if err != nil {
		log.Fatalf("Failed to load index page: %s", err)
		os.Exit(3)
	}

}

func main() {
	flag.Parse()
	log.PrintDebug = true
	log.Fileformat = "logs/%[1]s-%02[2]d.log"
	log.Init()
	log.Infoln("Initializing mau\\Lu")

	log.Debugln("Loading templates...")
	loadTemplates()
	log.Debugln("Loading config...")
	loadConfig()
	log.Debugln("Loading database...")
	loadDatabase()

	log.Infof("Listening on %s:%d", config.IP, config.Port)
	http.HandleFunc("/", get)
	http.HandleFunc("/shorten", shorten)
	http.ListenAndServe(config.IP+":"+strconv.Itoa(config.Port), nil)
}
