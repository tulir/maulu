// mau\Lu - A simple URL shortening backend.
// Copyright (C) 2016 Tulir Asokan
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
package main

import (
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	log "maunium.net/go/maulogger"
	"maunium.net/go/maulu/data"
	"net/http"
	"net/url"
	"strings"
)

// Response is a general response
type Response struct {
	Result    string `json:"result,omitempty"`
	Error     string `json:"error,omitempty"`
	ErrorLong string `json:"error-long,omitempty"`
}

// Request is a general shortening/unshortening request
type Request struct {
	Action       string `json:"action"`
	URL          string `json:"url"`
	RedirectType string `json:"redirect-type,omitempty"`
	RequestShort string `json:"short-request,omitempty"`
}

func get(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.Header().Add("Allow", "GET")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// Cut out the prefix slash
	path := r.URL.Path[1:]

	log.Debugf("%[1]s requested long url of %[2]s", getIP(r), path)
	url, redirect, err := data.Query(path)

	if err != nil {
		// No short url found
		log.Warnf("Failed to find redirect from short url %[2]s: %[1]s", err, path)
		writeError(w, http.StatusNotFound, "notfound", "404: %[1]s is not a valid short url", path)
		return
	}

	if redirect == "http" {
		// Short URL with HTTP redirect found.
		w.Header().Add("Location", url)
		w.WriteHeader(http.StatusFound)
	} else if redirect == "html" {
		// Short URL with HTML redirect found.
		templRedirect.Execute(w, struct{ URL interface{} }{url})
		w.WriteHeader(http.StatusOK)
	}
}

func query(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.Header().Add("Allow", "POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	ip := getIP(r)

	decoder := json.NewDecoder(r.Body)
	var req Request
	// Decode the payload.
	err := decoder.Decode(&req)

	if err != nil || len(req.Action) == 0 || len(req.URL) == 0 {
		log.Debugf("%[1]s sent an invalid insert request.", ip)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if req.Action == "unshorten" {
		if !strings.HasPrefix(req.URL, config.URL) {
			log.Warnf("%[1]s attempted to unshorten an invalid URL.", ip)
			writeError(w, http.StatusBadRequest, "notshortened", "The URL you entered is not a mau\\Lu short URL.")
			return
		}
		shortID := req.URL[len(config.URL):]
		if len(shortID) > 20 {
			log.Warnf("%[1]s attempted to unshorten an impossibly long short URL", ip)
			writeError(w, http.StatusBadRequest, "toolong", "The URL you entered is too long.")
			return
		}
		longURL, _, err := data.Query(shortID)
		if err != nil {
			log.Warnf("%[1]s queried the target of the non-existent short URL %[2]s", ip, req.URL)
			writeError(w, http.StatusNotFound, "notfound", "The short url id %[1]s doesn't exist!", req.URL)
			return
		}
		log.Debugf("%[1]s queried the target of %[2]s.", ip, req.URL)
		writeSuccess(w, longURL)
	} else if req.Action == "shorten" || req.Action == "google" || req.Action == "duckduckgo" {
		if len(req.RequestShort) == 0 {
			req.RequestShort = randomShortURL()
		}

		if req.RedirectType == "js" {
			req.RedirectType = "html"
		} else if len(req.RedirectType) == 0 || (req.RedirectType != "http" && req.RedirectType != "html") {
			req.RedirectType = "http"
		}

		if !validShortURL(req.RequestShort) {
			log.Warnf("%[1]s attempted to use invalid characters in a short URL", ip)
			writeError(w, http.StatusBadRequest, "illegalchars", "The short URL contains illegal characters.")
			return
		}

		if req.Action == "google" {
			req.URL = "http://lmgtfy.com/?q=" + url.QueryEscape(req.URL)
		} else if req.Action == "duckduckgo" {
			req.URL = "http://lmddgtfy.net/?q=" + strings.Replace(url.QueryEscape(req.URL), "+", " ", -1)
		} else {
			if strings.HasPrefix(req.URL, config.URL) {
				log.Warnf("%[1]s attempted to shorten the mau\\Lu url %[2]s", ip, req.URL)
				writeError(w, http.StatusBadRequest, "alreadyshortened", "The given URL is already a mau\\Lu URL")
				return
			} else if !strings.HasPrefix(req.URL, "https://") && !strings.HasPrefix(req.URL, "http://") {
				log.Warnf("%[1]s attempted to shorten an URL with an unidentified protocol", ip)
				writeError(w, http.StatusBadRequest, "invalidprotocol", "Protocol couldn't be identified.")
				return
			}
		}

		if len(req.URL) > 255 {
			log.Warnf("%[1]s attempted to shorten a very long URL", ip)
			writeError(w, http.StatusBadRequest, "toolong", "The URL you entered is too long.")
			return
		}

		str, _, err := data.Query(req.RequestShort)
		if (err == nil || len(str) != 0) && str != req.URL {
			log.Warnf("%[1]s attempted to insert %[3]s into the short url %[2]s, but it is already in use.", ip, req.RequestShort, req.URL)
			writeError(w, http.StatusConflict, "alreadyinuse", "The short url %[1]s is already in use.", req.RequestShort)
			return
		}

		resultURL := config.URL + data.Insert(req.URL, req.RequestShort, req.RedirectType)
		log.Debugf("%[1]s shortened %[3]s into %[2]s", ip, req.URL, resultURL)
		writeSuccess(w, resultURL)
	} else {
		log.Warnf("%[1]s attempted to use an unidentified action: %[2]s", ip, req.Action)
		writeError(w, http.StatusNotFound, "action", "Invalid action \"%[1]s\"", req.Action)
	}
}

func writeError(w http.ResponseWriter, errcode int, simple, errmsg string, args ...interface{}) {
	json, err := json.Marshal(Response{Error: simple, ErrorLong: fmt.Sprintf(errmsg, args...)})
	if err != nil {
		log.Errorf("Failed to marshal output json: %s", err)
		return
	}
	w.Write(json)
	w.WriteHeader(errcode)
}

func writeSuccess(w http.ResponseWriter, result string) {
	json, err := json.Marshal(Response{Result: result})
	if err != nil {
		log.Errorf("Failed to marshal output json: %s", err)
		return
	}
	w.Write(json)
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
