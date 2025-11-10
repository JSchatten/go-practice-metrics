package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type ServerFlags struct {
	ServerAddr      string
	ServerFileFlags ServerFileFlags
	PostgresDSN     string
}

type ServerFileFlags struct {
	FilePath      string
	FileInterval  time.Duration
	FileIsRestore bool
}

func InitServerFlags() (*ServerFlags, error) {
	var (
		serverAddr      = new(string)
		filePath        = new(string)
		fileIntervalSec = new(int)
		restoreFromFile = new(bool)
		postgresDSN     = new(string)
	)

	*serverAddr = constServerAddr
	*filePath = constFilePath
	*fileIntervalSec = constFileIntervalSec
	*restoreFromFile = constRestoreFromFile

	if v, exists := os.LookupEnv("ADDRESS"); exists {
		*serverAddr = v
	}
	if v, exists := os.LookupEnv("FILE_STORAGE_PATH"); exists {
		*filePath = v
	}
	if v, exists := os.LookupEnv("STORE_INTERVAL"); exists {
		if val, err := strconv.Atoi(v); err == nil && val >= 0 {
			*fileIntervalSec = val
		}
	}
	if v, exists := os.LookupEnv("RESTORE"); exists {
		v = strings.TrimSpace(v)
		switch strings.ToLower(v) {
		case "1", "t", "true":
			*restoreFromFile = true
		case "0", "f", "false", "":
			*restoreFromFile = false
		default:
			// Пытались передать странное, вернём обратно в дефолт
			*restoreFromFile = constRestoreFromFile
		}
	}
	if v, exists := os.LookupEnv("DATABASE_DSN"); exists {
		*postgresDSN = v
	}

	// Флаги
	flag.StringVar(serverAddr, "a", *serverAddr, fmt.Sprintf("Server address (default: '%s')", constServerAddr))
	flag.StringVar(filePath, "f", *filePath, fmt.Sprintf("File path for writing metrics into file (default: '%s')", constFilePath))
	flag.IntVar(fileIntervalSec, "i", *fileIntervalSec, fmt.Sprintf("Store interval in seconds (default: '%d')", constFileIntervalSec))
	flag.BoolVar(restoreFromFile, "r", *restoreFromFile, fmt.Sprintf("Restore metrics from file (default: '%t')", constRestoreFromFile))
	flag.StringVar(postgresDSN, "d", *postgresDSN, "DSN string for connectnion to Postgresql")

	flag.Parse()
	if flag.NArg() > 0 {
		return nil, fmt.Errorf("error: unknown flags: %v", flag.Args())
	}

	if *fileIntervalSec < 0 {
		return nil, ErrInvalidStoreInterval

	}

	var result = &ServerFlags{
		ServerAddr:  *serverAddr,
		PostgresDSN: *postgresDSN,
		ServerFileFlags: ServerFileFlags{
			FilePath:      *filePath,
			FileInterval:  time.Duration(*fileIntervalSec) * time.Second,
			FileIsRestore: *restoreFromFile,
		},
	}

	return result, nil
}
