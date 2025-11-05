package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
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

func getEnvStr(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

func getEnvBool(key string, defaultVal bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		value = strings.TrimSpace(value)
		switch strings.ToLower(value) {
		case "1", "t", "true":
			return true
		// Пусть пустое значение = false, т.к. не указано явно
		case "0", "f", "false", "":
			return false
		}
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if value, exists := os.LookupEnv(key); exists {
		if v, err := strconv.Atoi(value); err == nil {
			return v
		}
	}
	return defaultVal
}

func InitServerFlags() (*ServerFlags, error) {
	var serverAddr = getEnvStr("ADDRESS", "")
	var fileStoragePath = getEnvStr("FILE_STORAGE_PATH", "")
	var storeIntervalSec = getEnvInt("STORE_INTERVAL", 0)
	var restoreFromFile = getEnvBool("RESTORE", false)

	if serverAddr == "" {
		flag.StringVar(&serverAddr, "a", "localhost:8080", "Server address (default: localhost:8080)")
	}
	if storeIntervalSec == 0 {
		flag.IntVar(&storeIntervalSec, "i", 300, "File path for writeing metrics into file")
	}
	if fileStoragePath == "" {
		flag.StringVar(&fileStoragePath, "f", "metrics.json", "File path for writeing metrics into file")
	}
	if !restoreFromFile {
		// Таки включим по умолчанию попытку чтения из файла
		flag.BoolVar(&restoreFromFile, "r", true, "File path for writeing metrics into file")
	}
	flag.Parse()
	if flag.NArg() > 0 {
		return nil, fmt.Errorf("error: unknown flags: %v", flag.Args())
	}
	return &ServerFlags{
		ServerAddr:    serverAddr,
		FilePath:      fileStoragePath,
		FileIsRestore: restoreFromFile,
		FileInterval:  time.Duration(storeIntervalSec) * time.Second,
	}, nil
}

func InitAgentFlags() (*AgentFlags, error) {
	var reportInterval = getEnvInt("REPORT_INTERVAL", 0)
	var pollInterval = getEnvInt("POLL_INTERVAL", 0)
	var serverAddr = getEnvStr("ADDRESS", "")

	if pollInterval == 0 {
		flag.IntVar(&pollInterval, "p", 2, "Poll interval in seconds (default: 2)")
	}

	if reportInterval == 0 {
		flag.IntVar(&reportInterval, "r", 10, "Report interval in seconds (default: 10)")
	}

	if serverAddr == "" {
		flag.StringVar(&serverAddr, "a", "localhost:8080", "Server address (default: localhost:8080)")
	}
	flag.Parse()
	if flag.NArg() > 0 {
		return nil, fmt.Errorf("error: unknown flags: %v", flag.Args())
	}

	// Проверка корректности интервалов
	if reportInterval <= 0 {
		return nil, ErrInvalidReportInterval

	}
	if pollInterval <= 0 {
		return nil, ErrInvalidPollInterval
	}

	if pollInterval == 0 {
		pollInterval = 2
	}
	if reportInterval == 0 {
		reportInterval = 10
	}

	return &AgentFlags{
		PollInterval:   time.Duration(pollInterval) * time.Second,
		ReportInterval: time.Duration(reportInterval) * time.Second,
		ServerAddr:     serverAddr,
	}, nil
}
