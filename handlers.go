// mau\Lu - A simple URL shortening backend.
// Copyright (C) 2020 Tulir Asokan
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/gorilla/mux"
	log "maunium.net/go/maulogger/v2"

	"maunium.net/go/maulu/data"
)

type URLInfo struct {
	LongURL        string `json:"url"`
	ShortURL       string `json:"short_url"`
	ShortCode      string `json:"short_code"`
	RedirectMethod string `json:"redirect"`
}

type Error struct {
	ErrCode string `json:"errcode,omitempty"`
	Error   string `json:"error,omitempty"`
}

// ShortenRequest is a general shortening/unshortening request
type ShortenRequest struct {
	Type         string `json:"type"`
	URL          string `json:"url"`
	RedirectType string `json:"redirect,omitempty"`
	RequestShort string `json:"short_code,omitempty"`
}

type UnshortenRequest struct {
	URL string `json:"url"`
}

type RedirectTemplate struct {
	URL string
}

func addGetCORS(w http.ResponseWriter) {
	w.Header().Add("Access-Control-Allow-Method", "POST, GET, HEAD, OPTIONS")
	w.Header().Add("Access-Control-Allow-Headers", "Accept, Content-Type")
	w.Header().Add("Access-Control-Allow-Origin", "*")
}

func get(w http.ResponseWriter, r *http.Request) {
	addGetCORS(w)
	vars := mux.Vars(r)
	short := strings.ToLower(vars["short"])

	log.Debugfln("%[1]s requested long url of %[2]s", getIP(r), short)
	target, redirect, err := data.Query(short)

	if err != nil {
		log.Debugfln("Failed to find redirect from short url %s: %v", short, err)
		writeError(w, http.StatusNotFound, "NOT_FOUND", "%s is not an existing short url", short)
		return
	}

	accept := r.Header.Get("Accept")
	if accept == "application/json" {
		writeSuccess(w, http.StatusOK, target, redirect, short)
	} else if redirect == "http" {
		// Short URL with HTTP redirect found.
		w.Header().Add("Location", target)
		w.WriteHeader(http.StatusFound)
	} else if redirect == "html" {
		// Short URL with HTML redirect found.
		w.Header().Add("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		err := templRedirect.Execute(w, RedirectTemplate{target})
		if err != nil {
			log.Warnln("Error executing template:", err)
		}
	}
}

func options(w http.ResponseWriter, r *http.Request) {
	addGetCORS(w)
	vars := mux.Vars(r)
	short := strings.ToLower(vars["short"])

	_, _, err := data.Query(short)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func put(w http.ResponseWriter, r *http.Request) {
	addGetCORS(w)
	vars := mux.Vars(r)
	short := vars["short"]

	var req ShortenRequest
	contentType := r.Header.Get("Content-Type")
	if contentType == "application/json" {
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			log.Debugfln("%s sent an invalid shorten request.", getIP(r))
			writeError(w, http.StatusBadRequest, "NOT_JSON", "Request body is not JSON")
			return
		}
	} else if contentType == "text/plain" {
		urlBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Debugfln("%s sent an invalid shorten request.", getIP(r))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		req.URL = string(urlBytes)
	} else {
		w.Header().Add("Accept", "application/json,text/plain")
		writeError(w, http.StatusUnsupportedMediaType, "INVALID_MEDIA_TYPE", "PUTting an URL requires either JSON or plain text URL.")
		return
	}
	req.RequestShort = short
	actuallyShorten(w, getIP(r), req)
}

func shorten(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Method", "POST, OPTIONS")
	w.Header().Add("Access-Control-Allow-Headers", "Accept, Content-Type")
	w.Header().Add("Access-Control-Allow-Origin", "*")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var req ShortenRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Debugfln("%s sent an invalid shorten request.", getIP(r))
		writeError(w, http.StatusBadRequest, "NOT_JSON", "Request body is not JSON")
		return
	}
	actuallyShorten(w, getIP(r), req)
}

func unshorten(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Method", "POST, OPTIONS")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Add("Access-Control-Allow-Origin", "*")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var req UnshortenRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Debugfln("%s sent an invalid unshorten request.", getIP(r))
		writeError(w, http.StatusBadRequest, "NOT_JSON", "Request body is not JSON")
		return
	}

	parsed, err := url.Parse(req.URL)
	if err != nil {
		log.Debugfln("%s sent an invalid unshorten request.", getIP(r))
		writeError(w, http.StatusBadRequest, "INVALID_URL", "The given URL is not valid")
		return
	}
	if parsed.Host != baseURL.Host || !strings.HasPrefix(parsed.Path, baseURL.Path) {
		log.Debugfln("%s sent an unshorten request with a non-mau\\Lu URL.", getIP(r))
		writeError(w, http.StatusBadRequest, "NOT_SHORTENED", "The given URL is not a valid short URL")
		return
	}
	short := path.Base(parsed.Path)

	log.Debugfln("%[1]s requested long url of %[2]s", getIP(r), short)
	target, redirect, err := data.Query(short)

	if err != nil {
		log.Debugfln("Failed to find redirect from short url %s: %v", short, err)
		writeError(w, http.StatusNotFound, "NOT_FOUND", "%s is not an existing short url", short)
		return
	}

	writeSuccess(w, http.StatusOK, target, redirect, short)
}

func actuallyShorten(w http.ResponseWriter, ip string, req ShortenRequest) {
	if len(req.URL) == 0 {
		log.Debugfln("%s sent a shorten request with no URL", ip)
		writeError(w, http.StatusBadRequest, "MISSING_URL", "Request body does not contain URL")
		return
	}

	if req.RequestShort == "" {
		req.RequestShort = randomShortURL()
	} else {
		req.RequestShort = strings.ToLower(req.RequestShort)
	}

	if req.RedirectType == "" {
		req.RedirectType = "http"
	} else if req.RedirectType != "http" && req.RedirectType != "html" {
		log.Debugfln("%s attempted to use invalid redirect type", ip)
		writeError(w, http.StatusBadRequest, "UNKNOWN_REDIRECT_TYPE", "Redirect type %s is not allowed", req.RedirectType)
		return
	}

	if !validShortURL(req.RequestShort) {
		log.Debugfln("%s attempted to use invalid characters in a short URL", ip)
		writeError(w, http.StatusBadRequest, "ILLEGAL_CHARACTERS", "The short URL contains illegal characters")
		return
	}

	if strings.HasPrefix(req.URL, config.URL) {
		log.Debugfln("%[1]s attempted to shorten the mau\\Lu url %[2]s", ip, req.URL)
		writeError(w, http.StatusBadRequest, "ALREADY_SHORTENED", "The given URL is already a mau\\Lu URL")
		return
	}

	if req.Type == "google" {
		req.URL = "http://lmgtfy.com/?q=" + url.QueryEscape(req.URL)
	} else if req.Type == "duckduckgo" {
		req.URL = "http://lmddgtfy.net/?q=" + strings.ReplaceAll(url.QueryEscape(req.URL), "+", " ")
	} else {
		parsed, err := url.Parse(req.URL)
		if err != nil {
			log.Debugfln("%s attempted to shorten an invalid URL: %v", ip, err)
			writeError(w, http.StatusBadRequest, "INVALID_URL", "The given URL is not valid")
			return
		} else if parsed.Scheme != "http" && parsed.Scheme != "https" {
			log.Debugfln("%s attempted to shorten a URL with an unknown scheme %s: %v", ip, parsed.Scheme, err)
			writeError(w, http.StatusBadRequest, "ILLEGAL_SCHEME", "URL scheme %s is not allowed", parsed.Scheme)
			return
		}
		req.URL = parsed.String()
	}

	if len(req.URL) > 65535 {
		log.Debugfln("%[1]s attempted to shorten a very long URL", ip)
		writeError(w, http.StatusRequestEntityTooLarge, "TOO_LONG", "The URL you entered is too long")
		return
	}

	str, _, err := data.Query(req.RequestShort)
	if (err == nil || len(str) != 0) && str != req.URL {
		log.Debugfln("%s attempted to replace short URL %s, but it is already in use.", ip, req.RequestShort)
		writeError(w, http.StatusConflict, "ALREADY_IN_USE", "The short url %s is already in use.", req.RequestShort)
		return
	}

	insertResult, err := data.Insert(req.URL, req.RequestShort, req.RedirectType)
	if err != nil {
		log.Errorfln("Error inserting %v: %v", req, err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "The server encountered an error")
		return
	}
	status := http.StatusCreated
	if str != "" {
		status = http.StatusOK
	}
	log.Debugfln("%s shortened %s into %s", ip, req.URL, insertResult)
	writeSuccess(w, status, req.URL, req.RedirectType, insertResult)
}

func writeError(w http.ResponseWriter, errcode int, simple, errmsg string, args ...interface{}) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(errcode)
	err := json.NewEncoder(w).Encode(Error{ErrCode: simple, Error: fmt.Sprintf(errmsg, args...)})
	if err != nil {
		log.Warnln("Error encoding error response:", err)
	}
}

func writeSuccess(w http.ResponseWriter, status int, longURL, redirectMethod, shortCode string) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)

	resultURL, _ := url.Parse(config.URL)
	resultURL.Path = path.Join(resultURL.Path, shortCode)
	err := json.NewEncoder(w).Encode(URLInfo{
		LongURL:        longURL,
		RedirectMethod: redirectMethod,
		ShortCode:      shortCode,
		ShortURL:       resultURL.String(),
	})
	if err != nil {
		log.Warnln("Error encoding success response:", err)
	}
}

func validShortURL(short string) bool {
	if short == "api" {
		return false
	}
	for _, char := range short {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '_' || char == ' ' || char == '-' || char == '.' {
			continue
		}
		return false
	}
	return true
}
