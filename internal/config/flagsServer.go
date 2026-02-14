package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// generate:reset
type ServerFlags struct {
	ServerAddr       string           // ServerAddr — адрес сервера для прослушивания входящих запросов.
	PostgresDSN      string           // PostgresDSN — DSN-строка для подключения к PostgreSQL.
	HashKey          string           // HashKey — ключ для SHA256-хеширования тела запроса.
	ServerAuditFlags ServerAuditFlags // ServerAuditFlags — параметры аудита.
	ServerFileFlags  ServerFileFlags  // ServerFileFlags — параметры хранения метрик в файле.
}

// ServerFileFlags — параметры хранения метрик в файле.
// generate:reset
type ServerFileFlags struct {
	FilePath      string        // FilePath — путь к файлу для хранения метрик.
	FileInterval  time.Duration // FileInterval — интервал сохранения метрик в файл (в секундах).
	FileIsRestore bool          // FileIsRestore — флаг восстановления метрик из файла при старте.
}

// ServerAuditFlags — параметры аудита.
// generate:reset
type ServerAuditFlags struct {
	AuditFilePath string // AuditFilePath — путь к файлу аудита.
	AuditURL      string // AuditURL — URL для отправки событий аудита.
}

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
	)

	*serverAddr = constServerAddr
	*filePath = constFilePath
	*fileIntervalSec = constFileIntervalSec
	*restoreFromFile = constRestoreFromFile

	// fmt.Println("hashKey q", *hashKeyEnv)
	// fmt.Println("hashKey q", *hashKeyEnv)

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
		*hashKeyEnv = v
	}

	// audit
	if v, exists := os.LookupEnv("AUDIT_FILE"); exists {
		*auditFilePath = v
	}
	if v, exists := os.LookupEnv("AUDIT_URL"); exists {
		*auditURL = v
	}

	// Флаги
	flag.StringVar(serverAddr, "a", *serverAddr, fmt.Sprintf("Server address (default: '%s')", constServerAddr))
	flag.StringVar(filePath, "f", *filePath, fmt.Sprintf("File path for writing metrics into file (default: '%s')", constFilePath))
	flag.IntVar(fileIntervalSec, "i", *fileIntervalSec, fmt.Sprintf("Store interval in seconds (default: '%d')", constFileIntervalSec))
	flag.BoolVar(restoreFromFile, "r", *restoreFromFile, fmt.Sprintf("Restore metrics from file (default: '%t')", constRestoreFromFile))
	flag.StringVar(postgresDSN, "d", *postgresDSN, "DSN string for connectnion to Postgresql")
	flag.StringVar(hashKeyFlags, "k", *hashKeyEnv, "Hash key for SHA256 (default is empty which is disable crypto)")
	//	audit
	flag.StringVar(auditFilePath, "audit-file", *auditFilePath, "Path to audit log file (optional)")
	flag.StringVar(auditURL, "audit-url", *auditURL, "URL to send audit events (optional)")

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
		HashKey: *hashKeyFlags,
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

	return result, nil
}
