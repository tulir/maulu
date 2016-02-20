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
	"database/sql"
	"fmt"
	"strings"
)

var database *sql.DB

// LoadDatabase loads the database based on the given configuration.
func LoadDatabase(conf SQLConfig) error {
	var err error
	sqlType := strings.ToLower(conf.Type)
	if sqlType == "mysql" {
		database, err = sql.Open(sqlType, fmt.Sprintf("%[1]s@%[2]s/%[3]s", conf.Authentication.ToString(), conf.Connection.ToString(), conf.Database))
	} else {
		return fmt.Errorf("%[1]s is not yet supported", conf.Type)
	}

	if err != nil {
		return err
	} else if database == nil {
		return fmt.Errorf("Failed to open SQL connection!")
	}
	result, err := database.Query("CREATE TABLE IF NOT EXISTS links (url VARCHAR(255), short VARCHAR(20) NOT NULL, redirect VARCHAR(4), PRIMARY KEY(url, redirect));")
	if err != nil {
		return err
	} else if result.Err() != nil {
		return result.Err()
	}
	return nil
}

// DeleteShort deletes all the entries with the given short URL.
func DeleteShort(short string) error {
	_, err := database.Query("DELETE FROM links WHERE short=?", short)
	return err
}

// DeleteURL deletes all the entries pointing to the given URL.
func DeleteURL(url string) error {
	_, err := database.Query("DELETE FROM links WHERE url=?", url)
	return err
}

// Insert inserts the given URL, short url and redirect type into the database.
// If the URL has already been shortened with the same redirect type, the already existing short URL will be returned.
// In any other case, the requested short URL will be returned.
// Warning: This will NOT check if the short URL is in use.
func Insert(url, ishort, redirect string) string {
	redirect = strings.ToLower(redirect)
	if redirect != "http" && redirect != "html" && redirect != "js" {
		redirect = "http"
	}
	result, err := database.Query("SELECT short FROM links WHERE url=? AND redirect=?;", url, redirect)
	if err == nil {
		for result.Next() {
			if result.Err() != nil {
				break
			}
			var short string
			result.Scan(&short)
			if len(short) != 0 {
				return short
			}
		}
	}
	InsertDirect(ishort, url, redirect)
	return ishort
}

// InsertDirect inserts the given values into the database, no questions asked (except by the database itself)
func InsertDirect(short, url, redirect string) error {
	_, err := database.Query("INSERT INTO links VALUES(?, ?, ?);", url, short, redirect)
	if err != nil {
		return err
	}
	return nil
}

// Query queries for the given short URL and returns the long URL and redirect type.
func Query(short string) (string, string, error) {
	result, err := database.Query("SELECT url, redirect FROM links WHERE short=?;", short)
	if err != nil {
		return "", "", err
	}
	defer result.Close()
	for result.Next() {
		if result.Err() != nil {
			return "", "", result.Err()
		}
		var long, redirect string
		result.Scan(&long, &redirect)
		if len(long) == 0 {
			continue
		} else if len(redirect) == 0 {
			redirect = "http"
		}
		return long, redirect, nil
	}
	result.Close()
	return "", "", fmt.Errorf("ID not found")
}
