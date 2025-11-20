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
	HashKey         string
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
		hashKey         = new(string)
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
	if v, exists := os.LookupEnv("KEY"); exists {
		*hashKey = v
	}

	// Флаги
	flag.StringVar(serverAddr, "a", *serverAddr, fmt.Sprintf("Server address (default: '%s')", constServerAddr))
	flag.StringVar(filePath, "f", *filePath, fmt.Sprintf("File path for writing metrics into file (default: '%s')", constFilePath))
	flag.IntVar(fileIntervalSec, "i", *fileIntervalSec, fmt.Sprintf("Store interval in seconds (default: '%d')", constFileIntervalSec))
	flag.BoolVar(restoreFromFile, "r", *restoreFromFile, fmt.Sprintf("Restore metrics from file (default: '%t')", constRestoreFromFile))
	flag.StringVar(postgresDSN, "d", *postgresDSN, "DSN string for connectnion to Postgresql")
	flag.StringVar(hashKey, "k", *hashKey, "Hash key for SHA256 (default is empty which is disable crypto)")

	flag.Parse()
	if flag.NArg() > 0 {
		return nil, fmt.Errorf("error: unknown flags: %v", flag.Args())
	}

	// Проверим: был ли флаг -d передан явно
	wasDSNFlagSet := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "d" {
			wasDSNFlagSet = true
		}
	})

	// Если флаг -d был передан, но значение пустое — ошибка
	if wasDSNFlagSet && *postgresDSN == "" {
		return nil, ErrInvalidDSN
	}

	// Если DSN задан через env, но пуст — тоже ошибка
	if os.Getenv("DATABASE_DSN") != "" && *postgresDSN == "" {
		// Это может быть, только если env был "", но это редкий случай
		// На практике: если env="DATABASE_DSN=", то os.LookupEnv вернёт exists=true, v=""
		// Мы уже присвоили *postgresDSN = v, т.е. ""
		// Значит, если env существует и пуст — это тоже ошибка
		return nil, ErrInvalidDSN
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
		HashKey: *hashKey,
	}

	return result, nil
}
