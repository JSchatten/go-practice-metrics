package config

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog/log"
)

// AgentFlags содержит параметры командной строки агента.
//
// Поля:
//   - ServerAddr: адрес сервера для отправки метрик.
//   - HashKey: ключ для SHA256-хеширования тела запроса.
//   - RateLimit: количество одновременных HTTP-соединений.
//   - PollInterval: интервал опроса метрик из runtime (в секундах).
//   - ReportInterval: интервал отправки метрик на сервер (в секундах).
//
// generate:reset
type AgentFlags struct {
	ServerAddr     string        // ServerAddr — адрес сервера для отправки метрик.
	HashKey        string        // HashKey — ключ для SHA256-хеширования тела запроса.
	RateLimit      int           // RateLimit — количество одновременных HTTP-соединений на отправку метрик.
	PollInterval   time.Duration // PollInterval — интервал опроса метрик из runtime.
	ReportInterval time.Duration // ReportInterval — интервал отправки метрик на сервер.
	CryptoKey      string        // CryptoKey — путь к файлу с публичным ключом для шифрования тела запроса.
	GRPCServerAddr string        // GRPCServerAddr — адрес gRPC-сервера для отправки метрик.
}

// InitAgentFlags инициализирует и возвращает структуру AgentFlags, устанавливая значения флагов.
//
// Значения устанавливаются в порядке приоритета:
// 1. Флаги командной строки (имеют наивысший приоритет).
// 2. Переменные окружения.
// 3. Значения по умолчанию.
//
// Возвращает указатель на AgentFlags и ошибку, если значения параметров некорректны.
func InitAgentFlags() (*AgentFlags, error) {
	var (
		pollInterval   = new(int)
		reportInterval = new(int)
		serverAddr     = new(string)
		hashKey        = new(string)
		cryptoKey      = new(string)
		rateLimit      = new(int)
		// config
		configPath = new(string)
		// grpc
		gRPCServerAddr = new(string)
	)

	*pollInterval = constPollInterval
	*reportInterval = constReportInterval
	*serverAddr = constServerAddr

	// Получаем путь к конфигурационному файлу из окружения
	if v, exists := os.LookupEnv("CONFIG"); exists {
		*configPath = v
	}

	// Получаем адрес gRPC-сервера из окружения
	grpcAddr := constGRPCServerAddr
	if v, exists := os.LookupEnv("GRPC_SERVER_ADDRESS"); exists {
		grpcAddr = v
	}

	// Флаги
	flag.IntVar(pollInterval, "p", *pollInterval, fmt.Sprintf("Poll interval in seconds (default: %d)", constPollInterval))
	flag.IntVar(reportInterval, "r", *reportInterval, fmt.Sprintf("Report interval in seconds (default: %d)", constReportInterval))
	flag.StringVar(serverAddr, "a", *serverAddr, fmt.Sprintf("Server address (default: %s)", constServerAddr))
	flag.StringVar(gRPCServerAddr, "g", grpcAddr, fmt.Sprintf("gRPC server address (default: %s)", constGRPCServerAddr))
	flag.StringVar(hashKey, "k", *hashKey, "Hash key for SHA256 (default is empty which is disable crypto)")
	flag.StringVar(cryptoKey, "crypto-key", *cryptoKey, "Path to public key file for encrypting request body (optional)")
	flag.IntVar(rateLimit, "l", *rateLimit, fmt.Sprintf("Limit http-senders (default: %d)", constRateLimit))
	// config
	flag.StringVar(configPath, "c", *configPath, "Path to config file")
	flag.StringVar(configPath, "config", *configPath, "Path to config file")

	flag.Parse()
	if flag.NArg() > 0 {
		return nil, fmt.Errorf("error: unknown flags: %v", flag.Args())
	}

	// Загружаем конфигурацию из файла, если указан
	if *configPath != "" {
		config, err := LoadAgentConfig(*configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load agent config: %w", err)
		}
		// Применяем значения из файла, если они не были заданы через флаги
		if *serverAddr == constServerAddr {
			*serverAddr = config.Address
		}
		if *reportInterval == constReportInterval {
			interval, err := time.ParseDuration(config.ReportInterval)
			if err == nil {
				*reportInterval = int(interval.Seconds())
			} else {
				log.Warn().Err(err).Msg("invalid report_interval in config")
			}
		}
		if *pollInterval == constPollInterval {
			interval, err := time.ParseDuration(config.PollInterval)
			if err == nil {
				*pollInterval = int(interval.Seconds())
			} else {
				log.Warn().Err(err).Msg("invalid poll_interval in config")
			}
		}
		if *cryptoKey == "" {
			*cryptoKey = config.CryptoKey
		}
		if *gRPCServerAddr == "" {
			*gRPCServerAddr = config.GRPCServerAddr
		}
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
		CryptoKey:      *cryptoKey,
		GRPCServerAddr: grpcAddr,
	}, nil
}
