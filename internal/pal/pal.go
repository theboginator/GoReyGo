package pal

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

const (
	Info  = "INFO: "
	Warn  = "WARNING: "
	Error = "ERROR: "
)

var (
	warningLogger *log.Logger
	infoLogger    *log.Logger
	errorLogger   *log.Logger
	logDir        = "logs"
)

func init() {
	logPath, err := LogPath()
	// If the log file can't be set up, just quit
	if err != nil {
		log.Fatal(err)
	}
	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	infoLogger = log.New(file, "INFO: ", log.Ldate|log.Ltime)
	warningLogger = log.New(file, "WARNING: ", log.Ldate|log.Ltime)
	errorLogger = log.New(file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func PrintLog(level string, msg string) {
	switch level {
	case Info:
		infoLogger.Println(msg)
	case Warn:
		warningLogger.Println(msg)
	case Error:
		errorLogger.Println(msg)
	}
	fmt.Println(msg)
}

func BaseDir() (string, error) {
	path, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return path, nil
}

func LogPath() (string, error) {
	baseDir, err := BaseDir()
	if err != nil {
		return "", err
	}
	now := time.Now()
	logDate := now.Format("03-01-2006_15.04.05")
	logName := logDate + "_GoReyGo.log"
	return filepath.Join(baseDir, logDir, logName), nil
}
