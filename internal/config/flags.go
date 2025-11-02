package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"
)

type AgentFlags struct {
	ServerAddr     string
	PollInterval   time.Duration
	ReportInterval time.Duration
}

type ServerFlags struct {
	ServerAddr string
}

func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
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
	var serverAddr = getEnv("ADDRESS", "")

	flag.Parse()
	if flag.NArg() > 0 {
		return nil, fmt.Errorf("error: unknown flags: %v", flag.Args())
	}

	if serverAddr == "" {
		flag.StringVar(&serverAddr, "a", "localhost:8080", "Server address (default: localhost:8080)")
	}

	return &ServerFlags{
		ServerAddr: serverAddr,
	}, nil
}

func InitAgentFlags() (*AgentFlags, error) {
	var reportInterval = getEnvInt("REPORT_INTERVAL", 0)
	var pollInterval = getEnvInt("POLL_INTERVAL", 0)
	var serverAddr = getEnv("ADDRESS", "")

	flag.Parse()
	if flag.NArg() > 0 {
		return nil, fmt.Errorf("error: unknown flags: %v", flag.Args())
	}

	if pollInterval == 0 {
		flag.IntVar(&pollInterval, "p", 2, "Poll interval in seconds (default: 2)")
	}

	if reportInterval == 0 {
		flag.IntVar(&reportInterval, "r", 10, "Report interval in seconds (default: 10)")
	}

	if serverAddr == "" {
		flag.StringVar(&serverAddr, "a", "localhost:8080", "Server address (default: localhost:8080)")
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
