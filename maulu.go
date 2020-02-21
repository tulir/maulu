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
	"fmt"
	"html/template"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	flag "github.com/ogier/pflag"
	log "maunium.net/go/maulogger/v2"

	"maunium.net/go/maulu/data"
)

func getIP(r *http.Request) string {
	if config.TrustHeaders && len(r.Header.Get("X-Forwarded-For")) > 0 {
		return r.Header.Get("X-Forwarded-For")
	}
	return r.RemoteAddr
}

var debug = flag.BoolP("debug", "d", false, "Enable to print debug messages to stdout")
var confPath = flag.StringP("config", "c", "/etc/maulu/config.json", "The path of the mau\\Lu configuration file.")
var logPath = flag.StringP("logs", "l", "/var/log/maulu", "The path to store log files in")

var config *data.Configuration

var templRedirect *template.Template

func init() {
	flag.Parse()
}

func main() {
	// Configure the logger
	log.DefaultLogger.PrintLevel = log.LevelInfo.Severity
	if *debug {
		log.DefaultLogger.PrintLevel = log.LevelDebug.Severity
	}
	log.DefaultLogger.FileFormat = func(date string, i int) string {
		return fmt.Sprintf("%[3]s/%[1]s-%02[2]d.log", date, i, *logPath)
	}

	// Initialize the logger
	if len(*logPath) > 0 {
		err := log.OpenFile()
		if err != nil {
			log.Errorln("Error opening log file:", err)
		}
	}
	log.Infofln("Initializing mau\\Lu")

	loadConfig()
	loadTemplates()
	loadDatabase()

	log.Infofln("Listening on %s:%d", config.IP, config.Port)
	r := mux.NewRouter()
	r.HandleFunc("/api/shorten", shorten).Methods(http.MethodPost, http.MethodOptions)
	r.HandleFunc("/{short:[a-zA-Z0-9.-_ ]+}", get).Methods(http.MethodGet, http.MethodHead)
	r.HandleFunc("/{short:[a-zA-Z0-9.-_ ]+}", put).Methods(http.MethodPut)
	r.HandleFunc("/{short:[a-zA-Z0-9.-_ ]+}", options).Methods(http.MethodOptions)
	err := http.ListenAndServe(fmt.Sprintf("%s:%d", config.IP, config.Port), r)
	if err != nil {
		log.Fatalln("Fatal error listening:", err)
	}
}

func loadConfig() {
	log.Infoln("Loading config...")
	var err error
	config, err = data.LoadConfig(*confPath)
	if err != nil {
		log.Fatalfln("Failed to load config: %[1]s", err)
		os.Exit(1)
	}
	log.Debugln("Successfully loaded config.")
}

func loadDatabase() {
	log.Infoln("Loading database...")

	var err error
	err = data.LoadDatabase(config.Database)
	if err != nil {
		log.Fatalfln("Failed to load database: %[1]s", err)
		os.Exit(2)
	}

	log.Debugln("Successfully loaded database.")
}

func loadTemplates() {
	log.Infoln("Loading HTML redirect template...")

	var err error
	templRedirect, err = template.ParseFiles(config.RedirectTemplate)
	if err != nil {
		log.Fatalfln("Failed to load HTML redirect template: %s", err)
		os.Exit(3)
	}
	log.Debugln("Successfully loaded HTML redirect template.")
}
