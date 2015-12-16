package main

import (
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"html/template"
	"io/ioutil"
	log "maunium.net/go/maulogger"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func getIP(r *http.Request) string {
	if config.TrustHeaders {
		return r.Header.Get("X-Forwarded-For")
	}
	return r.RemoteAddr
}

func query(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	reqURL := r.Form.Get("url")
	action := r.URL.Path[len("/query/"):]
	if action == "unshorten" {
		if !strings.HasPrefix(reqURL, "https://mau.lu/") {
			log.Warnf("%[1]s attempted to unshorten an invalid URL.", getIP(r))
			writeError(w, "The URL you entered is not a mau\\Lu short URL.")
			return
		}
		shortID := reqURL[len(config.URL):]
		if len(shortID) > 20 {
			log.Warnf("%[1]s attempted to unshorten an impossibly long short URL", getIP(r))
			writeError(w, "The URL you entered is too long.")
			return
		}
		log.Debugf("%[1]s requested unshortening of %[2]s", getIP(r), reqURL)
		longURL, err := queryURL(shortID)
		if err != nil {
			log.Warnf("Failed to find long url from short url id %[2]s: %[1]s", err, reqURL)
			writeError(w, "The short url id %[1]s doesn't exist!", reqURL)
			return
		}
		w.Header().Add("Location", config.URL+"?url="+url.QueryEscape(longURL))
		w.WriteHeader(http.StatusFound)
	} else if action == "google" || action == "shorten" {
		if strings.HasPrefix(reqURL, "https://mau.lu") {
			log.Warnf("%[1]s attempted to shorten the mau\\Lu url %[2]s", getIP(r), reqURL)
			w.Header().Add("Location", config.URL+"?url="+url.QueryEscape(reqURL))
			w.WriteHeader(http.StatusFound)
			return
		} else if !strings.HasPrefix(reqURL, "https://") && !strings.HasPrefix(reqURL, "http://") {
			log.Warnf("%[1]s attempted to shorten an URL with an unidentified protocol", getIP(r))
			writeError(w, "Protocol couldn't be identified.")
			return
		}
		if action == "google" {
			reqURL = "http://lmgtfy.com/?q=" + url.QueryEscape(reqURL)
		}
		if len(reqURL) > 255 {
			log.Warnf("%[1]s attempted to shorten a very long URL", getIP(r))
			writeError(w, "The URL you entered is too long.")
			return
		}
		log.Debugf("%[1]s requested shortening of %[2]s", getIP(r), reqURL)
		w.Header().Add("Location", config.URL+"?url="+url.QueryEscape(config.URL+insert(reqURL)))
		w.WriteHeader(http.StatusFound)
	} else {
		log.Warnf("%[1]s attempted to use an unidentified action: %[2]s", getIP(r), action)
		writeError(w, "Invalid action \"%[1]s\"", action)
	}
}

func get(w http.ResponseWriter, r *http.Request) {
	// Cut out the prefix slash
	path := r.URL.Path[1:]
	if path == "" {
		// Path is empty, serve the index page
		queryURL, _ := url.QueryUnescape(r.URL.Query().Get("url"))
		queryErr, _ := url.QueryUnescape(r.URL.Query().Get("error"))
		if len(queryErr) != 0 {
			queryErr = fmt.Sprintf("<h3>Error: %s</h3>", queryErr)
			index.Execute(w, struct{ URL, Error interface{} }{queryURL, template.HTML(queryErr)})
		} else {
			index.Execute(w, struct{ URL, Error interface{} }{queryURL, ""})
		}
	} else if path == "favicon.ico" {
		w.Write(favicon)
	} else {
		// Path not recognized. Check if it's a redirect key
		log.Debugf("%[1]s requested long url of of %[2]s", getIP(r), path)
		url, err := queryURL(path)
		if err != nil {
			// No short url. Redirect to the index
			log.Warnf("Failed to find redirect from short url %[2]s: %[1]s", err, path)
			writeError(w, "404: https://mau.lu/%[1]s not found", path)
			return
		}
		// Short url identified. Redirect to long url
		w.Header().Add("Location", url)
		w.WriteHeader(http.StatusFound)
	}
}

var favicon []byte

var debug = flag.Bool("d", false, "Enable to print debug messages to stdout")

func main() {
	flag.Parse()
	// Configure the logger
	log.PrintDebug = *debug
	log.Fileformat = func(date string, i int) string { return fmt.Sprintf("logs/%[1]s-%02[2]d.log", date, i) }

	// Initialize the logger
	log.Init()
	log.Infof("Initializing mau\\Lu")

	log.Debugln("Loading config...")
	loadConfig()
	log.Debugln("Loading templates...")
	loadTemplates()
	log.Debugln("Loading database...")
	loadDatabase()

	log.Debugln("Loading favicon...")
	favicon, _ = ioutil.ReadFile(config.Favicon)

	log.Infof("Listening on %s:%d", config.IP, config.Port)
	http.HandleFunc("/query/", query)
	http.HandleFunc("/", get)
	http.ListenAndServe(config.IP+":"+strconv.Itoa(config.Port), nil)
}
