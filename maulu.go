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
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	flag "github.com/ogier/pflag"
	"html/template"
	log "maunium.net/go/maulogger"
	"maunium.net/go/maulu/data"
	"net/http"
	"os"
	"strconv"
)

func getIP(r *http.Request) string {
	if config.TrustHeaders {
		return r.Header.Get("X-Forwarded-For")
	}
	return r.RemoteAddr
}

var favicon []byte

var debug = flag.BoolP("debug", "d", false, "Enable to print debug messages to stdout")
var confPath = flag.StringP("config", "c", "/etc/maulu/config.json", "The path of the mau\\Lu configuration file.")

var config *data.Configuration

var templRedirect *template.Template

func init() {
	flag.Parse()
}

func main() {
	// Configure the logger
	log.PrintDebug = *debug
	log.Fileformat = func(date string, i int) string { return fmt.Sprintf("logs/%[1]s-%02[2]d.log", date, i) }

	// Initialize the logger
	log.Init()
	log.Infof("Initializing mau\\Lu")

	loadConfig()
	loadTemplates()
	loadDatabase()

	go stdinListen()

	log.Infof("Listening on %s:%d", config.IP, config.Port)
	http.HandleFunc("/query/", query)
	http.HandleFunc("/", get)
	http.ListenAndServe(config.IP+":"+strconv.Itoa(config.Port), nil)
}

func loadConfig() {
	log.Infoln("Loading config...")
	var err error
	config, err = data.LoadConfig(*confPath)
	if err != nil {
		log.Fatalf("Failed to load config: %[1]s", err)
		os.Exit(1)
	}
	log.Debugln("Successfully loaded config.")
}

func loadDatabase() {
	log.Infoln("Loading database...")

	var err error
	err = data.LoadDatabase(config.SQL)
	if err != nil {
		log.Fatalf("Failed to load database: %[1]s", err)
		os.Exit(2)
	}

	log.Debugln("Successfully loaded database.")
}

func loadTemplates() {
	log.Infoln("Loading HTML redirect template...")

	var err error
	templRedirect, err = template.ParseFiles(config.RedirectTemplate)
	if err != nil {
		log.Fatalf("Failed to load HTML redirect template: %s", err)
		os.Exit(3)
	}
	log.Debugln("Successfully loaded HTML redirect template.")
}
