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

// Package data contains database and configuration parsing related things
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
	TrustHeaders     bool      `json:"trust-headers"`
	RedirectTemplate string    `json:"redirect-template"`
	URL              string    `json:"url"`
	IP               string    `json:"ip"`
	Port             int       `json:"port"`
	SQL              SQLConfig `json:"sql"`
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

	return config, nil
}
