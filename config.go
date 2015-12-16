package main

import (
	"encoding/json"
	"fmt"
	flag "github.com/ogier/pflag"
	"io/ioutil"
	log "maunium.net/go/maulogger"
	"os"
	"strings"
)

// Configuration is a container struct for the configuration.
type Configuration struct {
	TrustHeaders bool      `json:"trust-headers"`
	Favicon      string    `json:"favicon"`
	URL          string    `json:"url"`
	IP           string    `json:"ip"`
	Port         int       `json:"port"`
	SQL          SQLConfig `json:"sql"`
}

// SQLConfig is the part of the config where details of the SQL database are stored.
type SQLConfig struct {
	Type           string      `json:"type"`
	Database       string      `json:"database"`
	Connection     SQLConnInfo `json:"connection"`
	Authentication SQLAuthInfo `json:"authentication"`
}

// SQLConnInfo contains the info about where to connect to.
type SQLConnInfo struct {
	Mode string `json:"mode"`
	IP   string `json:"ip"`
	Port int    `json:"port"`
}

// ToString turns a SQL connection info into a string for the DSN.
func (conn SQLConnInfo) ToString() string {
	mode := strings.ToLower(conn.Mode)
	if strings.HasPrefix(mode, "unix") {
		return fmt.Sprintf("%[1]s(%[2]s)", mode, conn.IP)
	}
	return fmt.Sprintf("%[1]s(%[2]s:%[3]d)", mode, conn.IP, conn.Port)
}

// SQLAuthInfo contains the username and password for the database.
type SQLAuthInfo struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// ToString turns a SQL authentication info into a string for the DSN.
func (auth SQLAuthInfo) ToString() string {
	if len(auth.Password) != 0 {
		return fmt.Sprintf("%[1]s:%[2]s", auth.Username, auth.Password)
	}
	return auth.Username
}

var confPath = flag.StringP("config", "c", "./config.json", "The path of the mau\\Lu configuration file.")

var config *Configuration

func loadConfig() {
	config = &Configuration{}
	// Read the file
	data, err := ioutil.ReadFile(*confPath)
	// Check if there was an error
	if err != nil {
		log.Fatalf("Failed to load config: %[1]s", err)
		os.Exit(1)
	}
	// No error, parse the data
	log.Infof("Reading config data...")
	err = json.Unmarshal(data, config)
	// Check if parsing failed
	if err != nil {
		log.Fatalf("Failed to load config: %[1]s", err)
		os.Exit(1)
	}
	log.Debugf("Successfully loaded config from disk")
}
