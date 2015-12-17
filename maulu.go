package main

import (
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"io/ioutil"
	log "maunium.net/go/maulogger"
	"net/http"
	"strconv"
)

func getIP(r *http.Request) string {
	if config.TrustHeaders {
		return r.Header.Get("X-Forwarded-For")
	}
	return r.RemoteAddr
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
