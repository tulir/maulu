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
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

var database *sql.DB

// LoadDatabase loads the database based on the given configuration.
func LoadDatabase(path string) error {
	var err error
	database, err = sql.Open("sqlite3", path)
	if err != nil {
		return err
	} else if database == nil {
		return fmt.Errorf("failed to open SQL connection")
	}
	_, err = database.Exec(`CREATE TABLE IF NOT EXISTS links (
		short    VARCHAR(255) PRIMARY KEY,
		url      TEXT,
		redirect VARCHAR(4),
		UNIQUE(url, redirect)
	)`)
	if err != nil {
		return err
	}
	return nil
}

// DeleteShort deletes all the entries with the given short URL.
func DeleteShort(short string) error {
	_, err := database.Exec("DELETE FROM links WHERE short=$1", short)
	return err
}

// DeleteURL deletes all the entries pointing to the given URL.
func DeleteURL(url string) error {
	_, err := database.Exec("DELETE FROM links WHERE url=$1", url)
	return err
}

// Insert inserts the given URL, short url and redirect type into the database.
// If the URL has already been shortened with the same redirect type, the already existing short URL will be returned.
// In any other case, the requested short URL will be returned.
// Warning: This will NOT check if the short URL is in use.
func Insert(url, ishort, redirect string) (string, bool, error) {
	redirect = strings.ToLower(redirect)
	if redirect != "http" && redirect != "html" {
		redirect = "http"
	}
	result, err := database.Query("SELECT short FROM links WHERE url=$1 AND redirect=$2", url, redirect)
	if err == nil {
		defer result.Close()
		for result.Next() {
			if result.Err() != nil {
				break
			}
			var short string
			_ = result.Scan(&short)
			if len(short) != 0 {
				return short, true, nil
			}
		}
	}
	err = InsertDirect(ishort, url, redirect)
	return ishort, false, err
}

// InsertDirect inserts the given values into the database, no questions asked (except by the database itself)
func InsertDirect(short, url, redirect string) error {
	_, err := database.Exec("INSERT INTO links (short, url, redirect) VALUES($1, $2, $3)", short, url, redirect)
	if err != nil {
		return err
	}
	return nil
}

// Query queries for the given short URL and returns the long URL and redirect type.
func Query(short string) (long, redirect string, err error) {
	result, err := database.Query("SELECT url, redirect FROM links WHERE short=$1", short)
	if err != nil {
		return "", "", err
	}
	defer result.Close()
	for result.Next() {
		if result.Err() != nil {
			return "", "", result.Err()
		}
		_ = result.Scan(&long, &redirect)
		if len(long) == 0 {
			continue
		} else if len(redirect) == 0 {
			redirect = "http"
		}
		return long, redirect, nil
	}
	return "", "", fmt.Errorf("ID not found")
}
