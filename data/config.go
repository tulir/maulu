package data

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	log "maunium.net/go/maulogger"
	"strings"
)

// Configuration is a container struct for the configuration.
type Configuration struct {
	TrustHeaders bool        `json:"trust-headers"`
	AllowAPI     bool        `json:"allow-api"`
	URL          string      `json:"url"`
	IP           string      `json:"ip"`
	Port         int         `json:"port"`
	Files        FilesConfig `json:"files"`
	SQL          SQLConfig   `json:"sql"`
}

// FilesConfig contains configuration data about HTML files.
type FilesConfig struct {
	HTMLDirectory    string            `json:"html-directory"`
	RedirectTemplate string            `json:"template-redirect"`
	ErrorPages       map[string]string `json:"error-pages"`
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

// LoadConfig loads a Configuration from the specified path.
func LoadConfig(path string) (*Configuration, error) {
	var config *Configuration
	// Read the file
	data, err := ioutil.ReadFile(path)
	// Check if there was an error
	if err != nil {
		return nil, err
	}
	// No error, parse the data
	log.Infof("Reading config data...")
	err = json.Unmarshal(data, config)
	// Check if parsing failed
	if err != nil {
		return nil, err
	}

	config.Files.RedirectTemplate = strings.Replace(config.Files.RedirectTemplate, "$htmldir", config.Files.HTMLDirectory, 1)

	for key, val := range config.Files.ErrorPages {
		config.Files.ErrorPages[key] = strings.Replace(val, "$htmldir", config.Files.HTMLDirectory, 1)
	}

	return config, nil
}
