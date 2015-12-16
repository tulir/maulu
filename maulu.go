package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	flag "github.com/ogier/pflag"
	"html/template"
	log "maunium.net/go/maulogger"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

func getIP(r *http.Request) string {
	if config.TrustHeaders {
		return r.Header.Get("X-Forwarded-For")
	}
	return r.RemoteAddr
}

func writeError(w http.ResponseWriter, err string, args ...interface{}) {
	w.Header().Add("Location", config.URL+"?error="+url.QueryEscape(fmt.Sprintf(err, args...)))
	w.WriteHeader(http.StatusFound)
}

func query(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	reqURL := r.Form.Get("url")
	action := r.URL.Path[len("/query/"):]
	if action == "unshorten" {
		log.Debugf("%[1]s requested unshortening of %[2]s", getIP(r), reqURL)
		longURL, err := queryURL(reqURL)
		if err != nil {
			log.Errorf("Failed to find long url from short url %[2]s: %[1]s", err, reqURL)
			writeError(w, "The short url %[1]s doesn't exist!", reqURL)
			return
		}
		w.Header().Add("Location", config.URL+"?url="+url.QueryEscape(config.URL+longURL))
		w.WriteHeader(http.StatusFound)
	} else if action == "google" || action == "shorten" {
		if action == "google" {
			log.Debugf("%[1]s requested lmgtfy shortening of %[2]s", getIP(r), reqURL)
			reqURL = "http://lmgtfy.com/?q=" + url.QueryEscape(reqURL)
		} else if action == "shorten" {
			log.Debugf("%[1]s requested shortening of %[2]s", getIP(r), reqURL)
		}
		w.Header().Add("Location", config.URL+"?url="+url.QueryEscape(config.URL+insert(reqURL)))
		w.WriteHeader(http.StatusFound)
	} else {
		log.Errorf("%[1]s attempted to use an unidentified action: %[2]s", getIP(r), action)
		writeError(w, "Invalid action \"%[1]s\"", action)
	}
}

func get(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[1:]
	if path == "" {
		log.Debugf("%[1]s requested the index page", getIP(r))
		reqURL, _ := url.QueryUnescape(r.URL.Query().Get("url"))
		reqErr, _ := url.QueryUnescape(r.URL.Query().Get("error"))
		if len(reqErr) != 0 {
			reqErr = fmt.Sprintf("<h3>Error: %s</h3>", reqErr)
			index.Execute(w, struct{ URL, Error interface{} }{reqURL, template.HTML(reqErr)})
		} else {
			index.Execute(w, struct{ URL, Error interface{} }{reqURL, ""})
		}
	} else if path == "favicon.ico" {
		log.Debugf("%[1]s requested the favicon", getIP(r))
	} else {
		log.Debugf("%[1]s requested long url of of %[2]s", getIP(r), path)
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

var index *template.Template

func loadTemplates() {
	var err error
	index, err = template.ParseFiles("index.html")
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

	log.Debugln("Loading config...")
	loadConfig()
	log.Debugln("Loading templates...")
	loadTemplates()
	log.Debugln("Loading database...")
	loadDatabase()

	log.Infof("Listening on %s:%d", config.IP, config.Port)
	http.HandleFunc("/query/", query)
	http.HandleFunc("/", get)
	http.ListenAndServe(config.IP+":"+strconv.Itoa(config.Port), nil)
}
