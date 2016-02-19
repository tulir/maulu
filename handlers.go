package main

import (
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"html/template"
	"io/ioutil"
	log "maunium.net/go/maulogger"
	"maunium.net/go/maulu/data"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// Output wraps mau\Lu API output messages
type Output struct {
	URL       string `json:"url,omitempty"`
	Error     string `json:"error,omitempty"`
	ErrorLong string `json:"error-long,omitempty"`
}

func get(w http.ResponseWriter, r *http.Request) {
	// Cut out the prefix slash
	path := r.URL.Path[1:]

	api := false
	if config.AllowAPI && r.URL.Query().Get("api") == "true" {
		api = true
	}

	if path == "" || path == "index.html" || path == "index.php" || path == "index" || path == "index.htm" {
		// Path is empty, serve the index page
		queryURL, _ := url.QueryUnescape(r.URL.Query().Get("url"))
		queryErr, _ := url.QueryUnescape(r.URL.Query().Get("error"))
		if len(queryErr) != 0 {
			queryErr = fmt.Sprintf("<h3>Error: %s</h3>", queryErr)
			templIndex.Execute(w, struct{ URL, Error interface{} }{queryURL, template.HTML(queryErr)})
		} else {
			templIndex.Execute(w, struct{ URL, Error interface{} }{queryURL, ""})
		}
	} else if _, err := os.Stat(config.Files.HTMLDirectory + path); err == nil {
		data, err := ioutil.ReadFile(config.Files.HTMLDirectory + path)
		if err != nil {
			w.Write(data)
			w.WriteHeader(http.StatusOK)
		} else {
			log.Errorf("Couldn't read file %s even though it seems to exist.", path)
			w.WriteHeader(http.StatusForbidden)
		}
	} else {
		// Path not recognized. Check if it's a redirect key
		log.Debugf("%[1]s requested long url of %[2]s", getIP(r), path)
		url, redirect, err := data.Query(path)
		if err != nil {
			// No short url. Redirect to the index
			log.Warnf("Failed to find redirect from short url %[2]s: %[1]s", err, path)
			writeError(w, api, "notfound", "404: https://mau.lu/%[1]s not found", path)
			return
		}
		// Short url identified. Redirect to long url
		if api {
			json, err := json.Marshal(struct {
				URL      string `json:"url"`
				Redirect string `json:"redirect"`
			}{url, redirect})
			if err != nil {
				log.Errorf("Failed to marshal output json: %s", err)
				return
			}
			w.Write(json)
		} else if redirect == "http" {
			w.Header().Add("Location", url)
			w.WriteHeader(http.StatusFound)
		} else if redirect == "html" || redirect == "js" {
			templRedirect.Execute(w, struct{ URL interface{} }{url})
		}
	}
}

func query(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	ip := getIP(r)
	reqURL := r.Form.Get("url")

	api := false
	if config.AllowAPI && r.URL.Query().Get("api") == "true" {
		api = true
	}

	action := r.URL.Path[len("/query/"):]
	if action == "unshorten" {
		if !strings.HasPrefix(reqURL, "https://mau.lu/") {
			log.Warnf("%[1]s attempted to unshorten an invalid URL.", ip)
			writeError(w, api, "notshortened", "The URL you entered is not a mau\\Lu short URL.")
			return
		}
		shortID := reqURL[len(config.URL):]
		if len(shortID) > 20 {
			log.Warnf("%[1]s attempted to unshorten an impossibly long short URL", ip)
			writeError(w, api, "length", "The URL you entered is too long.")
			return
		}
		longURL, _, err := data.Query(shortID)
		if err != nil {
			log.Warnf("%[1]s queried the target of the non-existent short URL %[2]s", ip, reqURL)
			writeError(w, api, "notfound", "The short url id %[1]s doesn't exist!", reqURL)
			return
		}
		log.Debugf("%[1]s queried the target of %[2]s.", ip, reqURL)
		writeSuccess(w, api, config.URL+"?url="+url.QueryEscape(longURL))
	} else if action == "shorten" || action == "google" || action == "duckduckgo" {
		reqShort := r.Form.Get("short")
		if len(reqShort) == 0 {
			reqShort = randomShortURL()
		}

		if !validShortURL(reqShort) {
			log.Warnf("%[1]s attempted to use invalid characters in a short URL", ip)
			writeError(w, api, "illegalchars", "The short URL contains illegal characters.")
			return
		}

		if action == "google" {
			reqURL = "http://lmgtfy.com/?q=" + url.QueryEscape(reqURL)
		} else if action == "duckduckgo" {
			reqURL = "http://lmddgtfy.net/?q=" + strings.Replace(url.QueryEscape(reqURL), "+", " ", -1)
		} else {
			if strings.HasPrefix(reqURL, "https://mau.lu") {
				log.Warnf("%[1]s attempted to shorten the mau\\Lu url %[2]s", ip, reqURL)
				writeSuccess(w, api, config.URL+"?url="+url.QueryEscape(reqURL))
				return
			} else if !strings.HasPrefix(reqURL, "https://") && !strings.HasPrefix(reqURL, "http://") {
				log.Warnf("%[1]s attempted to shorten an URL with an unidentified protocol", ip)
				writeError(w, api, "protocol", "Protocol couldn't be identified.")
				return
			}
		}

		if len(reqURL) > 255 {
			log.Warnf("%[1]s attempted to shorten a very long URL", ip)
			writeError(w, api, "length", "The URL you entered is too long.")
			return
		}

		str, _, err := data.Query(reqShort)
		if (err == nil || len(str) != 0) && str != reqURL {
			log.Warnf("%[1]s attempted to insert %[3]s into the short url %[2]s, but it is already in use.", ip, reqShort, reqURL)
			writeError(w, api, "used", "The short url %[1]s is already in use.", reqShort)
			return
		}

		resultURL := config.URL + data.Insert(reqURL, reqShort, r.Form.Get("redirect"))
		log.Debugf("%[1]s shortened %[3]s into %[2]s", ip, reqURL, resultURL)
		writeSuccess(w, api, config.URL+"?url="+url.QueryEscape(resultURL))
	} else {
		log.Warnf("%[1]s attempted to use an unidentified action: %[2]s", ip, action)
		writeError(w, api, "action", "Invalid action \"%[1]s\"", action)
	}
}

func writeError(w http.ResponseWriter, api bool, simple, err string, args ...interface{}) {
	if api {
		json, err := json.Marshal(Output{Error: simple, ErrorLong: fmt.Sprintf(err, args...)})
		if err != nil {
			log.Errorf("Failed to marshal output json: %s", err)
			return
		}
		w.Write(json)
	} else {
		w.Header().Add("Location", config.URL+"?error="+url.QueryEscape(fmt.Sprintf(err, args...)))
		w.WriteHeader(http.StatusFound)
	}
}

func writeSuccess(w http.ResponseWriter, api bool, url string) {
	if api {
		json, err := json.Marshal(Output{URL: url})
		if err != nil {
			log.Errorf("Failed to marshal output json: %s", err)
			return
		}
		w.Write(json)
	} else {
		w.Header().Add("Location", url)
		w.WriteHeader(http.StatusFound)
	}
}

func validShortURL(short string) bool {
	for _, char := range short {
		if (char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '_' || char == ' ' {
			continue
		}
		return false
	}
	return true
}
