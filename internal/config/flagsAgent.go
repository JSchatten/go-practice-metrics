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
	HashKey        string
}

func InitAgentFlags() (*AgentFlags, error) {
	var (
		pollInterval   = new(int)
		reportInterval = new(int)
		serverAddr     = new(string)
		hashKey        = new(string)
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

	if v, exists := os.LookupEnv("KEY"); exists {
		*hashKey = v
	}

	// Регистрируем флаги — они перекроют env и default
	flag.IntVar(pollInterval, "p", *pollInterval, fmt.Sprintf("Poll interval in seconds (default: %d)", constPollInterval))
	flag.IntVar(reportInterval, "r", *reportInterval, fmt.Sprintf("Report interval in seconds (default: %d)", constReportInterval))
	flag.StringVar(serverAddr, "a", *serverAddr, fmt.Sprintf("Server address (default: %s)", constServerAddr))
	flag.StringVar(hashKey, "k", *hashKey, "Hash key for SHA256 (default is empty which is disable crypto)")

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
		HashKey:        *hashKey,
	}, nil
}
