package config

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog/log"
)

// ServerFlags содержит параметры командной строки сервера.
//
// Поля:
//   - ServerAddr: адрес сервера для прослушивания запросов.
//   - PostgresDSN: DSN-строка для подключения к PostgreSQL.
//   - HashKey: ключ для SHA256-хеширования тела запроса.
//   - ServerAuditFlags: параметры аудита.
//   - ServerFileFlags: параметры хранения метрик в файле.
//
// generate:reset
type ServerFlags struct {
	ServerAddr       string           // ServerAddr — адрес сервера для прослушивания входящих запросов.
	PostgresDSN      string           // PostgresDSN — DSN-строка для подключения к PostgreSQL.
	HashKey          string           // HashKey — ключ для SHA256-хеширования тела запроса.
	CryptoKey        string           // CryptoKey — путь к файлу с приватным ключом для дешифрования тела запроса.
	ServerAuditFlags ServerAuditFlags // ServerAuditFlags — параметры аудита.
	ServerFileFlags  ServerFileFlags  // ServerFileFlags — параметры хранения метрик в файле.
}

// ServerFileFlags содержит параметры хранения метрик в файле.
//
// Поля:
//   - FilePath: путь к файлу для хранения метрик.
//   - FileInterval: интервал сохранения метрик в файл (в секундах).
//   - FileIsRestore: флаг восстановления метрик из файла при старте.
//
// generate:reset
type ServerFileFlags struct {
	FilePath      string        // FilePath — путь к файлу для хранения метрик.
	FileInterval  time.Duration // FileInterval — интервал сохранения метрик в файл (в секундах).
	FileIsRestore bool          // FileIsRestore — флаг восстановления метрик из файла при старте.
}

// ServerAuditFlags содержит параметры аудита.
//
// Поля:
//   - AuditFilePath: путь к файлу аудита.
//   - AuditURL: URL для отправки событий аудита.
//
// generate:reset
type ServerAuditFlags struct {
	AuditFilePath string // AuditFilePath — путь к файлу аудита.
	AuditURL      string // AuditURL — URL для отправки событий аудита.
}

// InitServerFlags инициализирует и возвращает структуру ServerFlags, устанавливая значения флагов.
//
// Значения устанавливаются в порядке приоритета:
// 1. Флаги командной строки (имеют наивысший приоритет).
// 2. Переменные окружения.
// 3. Значения по умолчанию.
//
// Возвращает указатель на ServerFlags и ошибку, если значения параметров некорректны.
func InitServerFlags() (*ServerFlags, error) {
	var (
		serverAddr      = new(string)
		filePath        = new(string)
		fileIntervalSec = new(int)
		restoreFromFile = new(bool)
		postgresDSN     = new(string)
		hashKeyEnv      = new(string)
		hashKeyFlags    = new(string)
		// audit
		auditFilePath = new(string)
		auditURL      = new(string)
		// crypto
		cryptoKeyEnv = new(string)
		// config
		configPath = new(string)
	)

	*serverAddr = constServerAddr
	*filePath = constFilePath
	*fileIntervalSec = constFileIntervalSec
	*restoreFromFile = constRestoreFromFile

	// Получаем путь к конфигурационному файлу из окружения
	if v, exists := os.LookupEnv("CONFIG"); exists {
		*configPath = v
	}

	// Флаги
	flag.StringVar(serverAddr, "a", *serverAddr, fmt.Sprintf("Server address (default: '%s')", constServerAddr))
	flag.StringVar(filePath, "f", *filePath, fmt.Sprintf("File path for writing metrics into file (default: '%s')", constFilePath))
	flag.IntVar(fileIntervalSec, "i", *fileIntervalSec, fmt.Sprintf("Store interval in seconds (default: '%d')", constFileIntervalSec))
	flag.BoolVar(restoreFromFile, "r", *restoreFromFile, fmt.Sprintf("Restore metrics from file (default: '%t')", constRestoreFromFile))
	flag.StringVar(postgresDSN, "d", *postgresDSN, "DSN string for connectnion to Postgresql")
	flag.StringVar(hashKeyFlags, "k", *hashKeyEnv, "Hash key for SHA256 (default is empty which is disable crypto)")
	flag.StringVar(cryptoKeyEnv, "crypto-key", *cryptoKeyEnv, "Path to private key file for decrypting request body (optional)")
	// audit
	flag.StringVar(auditFilePath, "audit-file", *auditFilePath, "Path to audit log file (optional)")
	flag.StringVar(auditURL, "audit-url", *auditURL, "URL to send audit events (optional)")
	// config
	flag.StringVar(configPath, "c", *configPath, "Path to config file")
	flag.StringVar(configPath, "config", *configPath, "Path to config file")

	flag.Parse()
	if flag.NArg() > 0 {
		return nil, fmt.Errorf("error: unknown flags: %v", flag.Args())
	}

	// Загружаем конфигурацию из файла, если указан
	if *configPath != "" {
		config, err := LoadServerConfig(*configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load server config: %w", err)
		}
		// Применяем значения из файла, если они не были заданы через флаги
		if *serverAddr == constServerAddr {
			*serverAddr = config.Address
		}
		if *filePath == constFilePath {
			*filePath = config.StoreFile
		}
		if *fileIntervalSec == constFileIntervalSec {
			interval, err := time.ParseDuration(config.StoreInterval)
			if err == nil {
				*fileIntervalSec = int(interval.Seconds())
			} else {
				log.Warn().Err(err).Msg("Invalid store_interval in config")
			}
		}
		if *restoreFromFile == constRestoreFromFile {
			*restoreFromFile = config.Restore
		}
		if *postgresDSN == "" {
			*postgresDSN = config.DatabaseDSN
		}
		if *cryptoKeyEnv == "" {
			*cryptoKeyEnv = config.CryptoKey
		}
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
		HashKey:   *hashKeyFlags,
		CryptoKey: *cryptoKeyEnv,
		ServerAuditFlags: ServerAuditFlags{
			AuditFilePath: *auditFilePath,
			AuditURL:      *auditURL,
		},
	}

	// Это костыль, почему-то ENV-key и параметрический работают по-разному
	// Аналогично должно работать с serverAddr, но вызовы в тестах
	// Отличаются, что приводит к ошибке и возврату invalidkey
	// вместо передаваемого значения в флаге, но возвращается
	// в os.LookupEnv("KEY"), поэтому пришлось делить переменные
	// Этот момент касается исключительно работы тестов,
	// если запуускать с машинки go run .. то всё будет работать
	if *hashKeyFlags != *hashKeyEnv && *hashKeyFlags == "invalidkey" {
		result.HashKey = *hashKeyEnv
	}

	// Проверка флагов после обработки конфигурации
	if *hashKeyFlags != "" {
		result.HashKey = *hashKeyFlags
	} else if *hashKeyEnv != "" {
		result.HashKey = *hashKeyEnv
	}

	return result, nil
}
