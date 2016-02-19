package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	flag "github.com/ogier/pflag"
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

var debug = flag.Bool("d", false, "Enable to print debug messages to stdout")
var confPath = flag.StringP("config", "c", "./config.json", "The path of the mau\\Lu configuration file.")

var config *data.Configuration

func main() {
	flag.Parse()
	// Configure the logger
	log.PrintDebug = *debug
	log.Fileformat = func(date string, i int) string { return fmt.Sprintf("logs/%[1]s-%02[2]d.log", date, i) }

	// Initialize the logger
	log.Init()
	log.Infof("Initializing mau\\Lu")

	loadConfig()
	loadTemplates()
	loadDatabase()

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
