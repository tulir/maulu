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

package data

import (
	"encoding/json"
	"io/ioutil"

	log "maunium.net/go/maulogger/v2"
)

// Configuration is a container struct for the configuration.
type Configuration struct {
	TrustHeaders     bool   `json:"trust-headers"`
	RedirectTemplate string `json:"redirect-template"`
	URL              string `json:"url"`
	IP               string `json:"ip"`
	Port             int    `json:"port"`
	Database         string `json:"database"`
}

// LoadConfig loads a Configuration from the specified path.
func LoadConfig(path string) (*Configuration, error) {
	var config = &Configuration{}
	// Read the file
	data, err := ioutil.ReadFile(path)
	// Check if there was an error
	if err != nil {
		return nil, err
	}
	// No error, parse the data
	log.Infofln("Reading config data...")
	err = json.Unmarshal(data, config)
	// Check if parsing failed
	if err != nil {
		return nil, err
	}

	return config, nil
}
