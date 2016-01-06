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
	result, err := database.Query("CREATE TABLE IF NOT EXISTS links (url VARCHAR(255), short VARCHAR(20) NOT NULL, redirect VARCHAR(4), PRIMARY KEY(url, redirect));")
	if err != nil {
		log.Errorf("Failed to create database: %s", err)
	}
	if result.Err() != nil {
		log.Errorf("Failed to create database: %s", result.Err())
	}
}

func insert(url, ishort, redirect string) string {
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
	insertURL(ishort, url, redirect)
	return ishort
}

func insertURL(short, url, redirect string) error {
	_, err := database.Query("INSERT INTO links VALUES(?, ?, ?);", url, short, redirect)
	if err != nil {
		return err
	}
	return nil
}

func queryURL(short string) (string, string, error) {
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
	return "", "", errors.New("ID not found")
}
