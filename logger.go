package main

import (
	"os"
	"errors"
)

type LOG_TYPE int

const (
	ERROR LOG_TYPE = iota
	BLOCKED
)

//write data to a logger file
func writeLog(logType LOG_TYPE, data string) error {

	var path string
	switch logType {
	case ERROR:
		path = "error.txt"
	case BLOCKED:
		path = "blocked.txt"
	default:
		return errors.New("invalid log type")
	}

	//create file if not exists
	if _, err := os.Stat(path); os.IsNotExist(err) { // see we have two statements here separated by a semicolon. This is a common pattern in Go
		file, err := os.Create(path)
		if err != nil {
			return err
		}
		defer file.Close()
	}

	//open file
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	//write data to file
	if _, err = file.WriteString(data); err != nil {
		return err
	}

	return nil
}

func formatBlockText(url string) string {
	return "Blocked request to: " + url + "\n"
}