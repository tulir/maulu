package main

import (
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

// GetIP gets the IP that sent the given request.
func GetIP(r *http.Request, trustHeaders bool) string {
	if trustHeaders {
		return r.Header.Get("X-Forwarded-For")
	}
	return r.RemoteAddr
}

func query(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	reqURL := r.Form.Get("url")
	action := r.URL.Path[len("/query/"):]
	if action == "unshorten" {
		log.Debugf("%[1]s requested unshortening of %[2]s", GetIP(r, config.TrustHeaders), reqURL)
		longURL, err := queryURL(reqURL[len(config.URL):])
		if err != nil {
			log.Warnf("Failed to find long url from short url id %[2]s: %[1]s", err, reqURL)
			writeError(w, "The short url id %[1]s doesn't exist!", reqURL)
			return
		}
		w.Header().Add("Location", config.URL+"?url="+url.QueryEscape(longURL))
		w.WriteHeader(http.StatusFound)
	} else if action == "google" || action == "shorten" {
		if strings.HasPrefix(reqURL, "https://mau.lu") {
			log.Warnf("%[1]s attempted to shorten the mau\\Lu url %[2]s", GetIP(r, config.TrustHeaders), reqURL)
			w.Header().Add("Location", config.URL+"?url="+url.QueryEscape(reqURL))
			w.WriteHeader(http.StatusFound)
			return
		} else if !strings.HasPrefix(reqURL, "https://") && !strings.HasPrefix(reqURL, "http://") {
			log.Warnf("%[1]s attempted to shorten an URL with an unidentified protocol", GetIP(r, config.TrustHeaders))
			writeError(w, "Protocol couldn't be identified.")
			return
		}
		if action == "google" {
			reqURL = "http://lmgtfy.com/?q=" + url.QueryEscape(reqURL)
		}
		if len(reqURL) > 255 {
			log.Warnf("%[1]s attempted to shorten a very long URL", GetIP(r, config.TrustHeaders))
			writeError(w, "The URL you entered is too long.")
			return
		}
		log.Debugf("%[1]s requested shortening of %[2]s", GetIP(r, config.TrustHeaders), reqURL)
		w.Header().Add("Location", config.URL+"?url="+url.QueryEscape(config.URL+insert(reqURL)))
		w.WriteHeader(http.StatusFound)
	} else {
		log.Warnf("%[1]s attempted to use an unidentified action: %[2]s", GetIP(r, config.TrustHeaders), action)
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
		log.Debugf("%[1]s requested the favicon", GetIP(r, config.TrustHeaders))
		w.Write(favicon)
	} else {
		// Path not recognized. Check if it's a redirect key
		log.Debugf("%[1]s requested long url of of %[2]s", GetIP(r, config.TrustHeaders), path)
		url, err := queryURL(path)
		if err != nil {
			// No short url. Redirect to the index
			log.Warnf("Failed to find redirect from short url %[2]s: %[1]s", err, path)
			writeError(w, "Error 404: https://mau.lu/%[1]s not found", path)
			return
		}
		// Short url identified. Redirect to long url
		w.Header().Add("Location", url)
		w.WriteHeader(http.StatusFound)
	}
}

var favicon []byte

func main() {
	// Configure the logger
	log.PrintDebug = true
	log.Fileformat = func(s string, i int) string { return "maulu.log" }

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
