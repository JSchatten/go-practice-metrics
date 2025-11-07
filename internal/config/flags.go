package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	// common
	constServerAddr = "localhost:8080"
	// agent
	constPollInterval   = 2
	constReportInterval = 10
	// server
	constFilePath        = "./metrics.json"
	constFileIntervalSec = 300
	constRestoreFromFile = true
)

type AgentFlags struct {
	ServerAddr     string
	PollInterval   time.Duration
	ReportInterval time.Duration
}

type ServerFlags struct {
	ServerAddr    string
	FilePath      string
	FileInterval  time.Duration
	FileIsRestore bool
}

// TODO Вынести флаги для работы с файлом
// type ServerFileFlags struct {
// 	FilePath      string
// 	FileInterval  time.Duration
// 	FileIsRestore bool
// }

func InitAgentFlags() (*AgentFlags, error) {
	var (
		pollInterval   = new(int)
		reportInterval = new(int)
		serverAddr     = new(string)
	)

	*pollInterval = constPollInterval
	*reportInterval = constReportInterval
	*serverAddr = constServerAddr

	// Читаем окружение
	if v, exists := os.LookupEnv("POLL_INTERVAL"); exists {
		if val, err := strconv.Atoi(v); err == nil {
			*pollInterval = val
		}
	}

	if v, exists := os.LookupEnv("REPORT_INTERVAL"); exists {
		if val, err := strconv.Atoi(v); err == nil {
			*reportInterval = val
		}
	}

	if v, exists := os.LookupEnv("ADDRESS"); exists {
		*serverAddr = v
	}

	// Регистрируем флаги — они перекроют env и default
	flag.IntVar(pollInterval, "p", *pollInterval, fmt.Sprintf("Poll interval in seconds (default: %d)", constPollInterval))
	flag.IntVar(reportInterval, "r", *reportInterval, fmt.Sprintf("Report interval in seconds (default: %d)", constReportInterval))
	flag.StringVar(serverAddr, "a", *serverAddr, fmt.Sprintf("Server address (default: %s)", constServerAddr))

	flag.Parse()
	if flag.NArg() > 0 {
		return nil, fmt.Errorf("error: unknown flags: %v", flag.Args())
	}

	// Финальная проверка
	if *pollInterval <= 0 {
		return nil, ErrInvalidPollInterval
	}
	if *reportInterval <= 0 {
		return nil, ErrInvalidReportInterval
	}

	return &AgentFlags{
		PollInterval:   time.Duration(*pollInterval) * time.Second,
		ReportInterval: time.Duration(*reportInterval) * time.Second,
		ServerAddr:     *serverAddr,
	}, nil
}

func InitServerFlags() (*ServerFlags, error) {
	var (
		serverAddr      = new(string)
		filePath        = new(string)
		fileIntervalSec = new(int)
		restoreFromFile = new(bool)
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

	// Флаги
	flag.StringVar(serverAddr, "a", *serverAddr, fmt.Sprintf("Server address (default: '%s')", constServerAddr))
	flag.StringVar(filePath, "f", *filePath, fmt.Sprintf("File path for writing metrics into file (default: '%s')", constFilePath))
	flag.IntVar(fileIntervalSec, "i", *fileIntervalSec, fmt.Sprintf("Store interval in seconds (default: '%d')", constFileIntervalSec))
	flag.BoolVar(restoreFromFile, "r", *restoreFromFile, fmt.Sprintf("Restore metrics from file (default: '%t')", constRestoreFromFile))

	flag.Parse()
	if flag.NArg() > 0 {
		return nil, fmt.Errorf("error: unknown flags: %v", flag.Args())
	}

	if *fileIntervalSec < 0 {
		return nil, ErrInvalidStoreInterval

	}

	return &ServerFlags{
		ServerAddr:    *serverAddr,
		FilePath:      *filePath,
		FileInterval:  time.Duration(*fileIntervalSec) * time.Second,
		FileIsRestore: *restoreFromFile,
	}, nil
}
