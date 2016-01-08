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
