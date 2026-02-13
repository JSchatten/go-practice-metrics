package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"
)

type AgentFlags struct {
	ServerAddr     string        // ServerAddr — адрес сервера для отправки метрик.
	HashKey        string        // HashKey — ключ для SHA256-хеширования тела запроса.
	RateLimit      int           // RateLimit — количество одновременных HTTP-соединений на отправку метрик.
	PollInterval   time.Duration // PollInterval — интервал опроса метрик из runtime.
	ReportInterval time.Duration // ReportInterval — интервал отправки метрик на сервер.
}

func InitAgentFlags() (*AgentFlags, error) {
	var (
		pollInterval   = new(int)
		reportInterval = new(int)
		serverAddr     = new(string)
		hashKey        = new(string)
		rateLimit      = new(int)
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
	if v, exists := os.LookupEnv("RATE_LIMIT"); exists {
		if val, err := strconv.Atoi(v); err == nil {
			*rateLimit = val
		}
	}

	// Регистрируем флаги — они перекроют env и default
	flag.IntVar(pollInterval, "p", *pollInterval, fmt.Sprintf("Poll interval in seconds (default: %d)", constPollInterval))
	flag.IntVar(reportInterval, "r", *reportInterval, fmt.Sprintf("Report interval in seconds (default: %d)", constReportInterval))
	flag.StringVar(serverAddr, "a", *serverAddr, fmt.Sprintf("Server address (default: %s)", constServerAddr))
	flag.StringVar(hashKey, "k", *hashKey, "Hash key for SHA256 (default is empty which is disable crypto)")
	flag.IntVar(rateLimit, "l", *rateLimit, fmt.Sprintf("Limit http-senders (default: %d)", constRateLimit))

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

	if *rateLimit <= 0 {
		if *rateLimit == 0 {
			*rateLimit = constRateLimit
		} else {
			return nil, ErrInvalidRateLimit
		}
	}

	return &AgentFlags{
		PollInterval:   time.Duration(*pollInterval) * time.Second,
		ReportInterval: time.Duration(*reportInterval) * time.Second,
		ServerAddr:     *serverAddr,
		HashKey:        *hashKey,
		RateLimit:      *rateLimit,
	}, nil
}
