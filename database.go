package main

import (
	"database/sql"
	"errors"
	"fmt"
	log "maunium.net/go/maulogger"
	"os"
	"strings"
)

var database *sql.DB

func loadDatabase() {
	var err error
	sqlType := strings.ToLower(config.SQL.Type)
	if sqlType == "mysql" {
		database, err = sql.Open(sqlType, fmt.Sprintf("%[1]s@%[2]s/%[3]s", config.SQL.Authentication.ToString(), config.SQL.Connection.ToString(), config.SQL.Database))
	} else {
		log.Fatalf("%[1]s is not yet supported.", config.SQL.Type)
		os.Exit(2)
	}

	if err != nil {
		log.Fatalf("Error while opening SQL connection: %s", err)
		os.Exit(2)
	} else if database == nil {
		log.Fatalf("Failed to open SQL connection!")
		os.Exit(2)
	}
	result, err := database.Query("CREATE TABLE IF NOT EXISTS links (url VARCHAR(255) PRIMARY KEY, short VARCHAR(20) NOT NULL);")
	if err != nil {
		log.Errorf("Failed to create database: %s", err)
	}
	if result.Err() != nil {
		log.Errorf("Failed to create database: %s", result.Err())
	}
}

func insert(url string) string {
	result, err := database.Query("SELECT short FROM links WHERE url=?;", url)
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
	rand := randomShortURL()
	insertURL(rand, url)
	return rand
}

func insertURL(short string, url string) error {
	_, err := database.Query("INSERT INTO links VALUES(?, ?);", url, short)
	if err != nil {
		return err
	}
	return nil
}

func queryURL(short string) (string, error) {
	result, err := database.Query("SELECT url FROM links WHERE short=?;", short)
	if err != nil {
		return "", err
	}
	defer result.Close()
	for result.Next() {
		if result.Err() != nil {
			return "", result.Err()
		}
		var long string
		result.Scan(&long)
		if len(long) == 0 {
			continue
		}
		return long, nil
	}
	result.Close()
	return "", errors.New("ID not found")
}
