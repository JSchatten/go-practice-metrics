// Package main - точка входа сервера сбора метрик.
//
// Запускает HTTP-сервер, обрабатывающий:
//   - Приём метрик (POST /update, /updates)
//   - Получение значений (POST /value, GET /value/...)
//   - Проверку состояния (GET /ping, /)
//
// Поддерживает:
//   - Хранение в памяти или PostgreSQL
//   - Автосохранение на диск
//   - Восстановление из файла
//   - Аудит изменений (в файл или по HTTP)
//   - Проверку хешей
//   - Gzip-сжатие
//   - Логирование запросов
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JSchatten/go-practice-metrics/internal/config"
	// handlers "github.com/JSchatten/go-practice-metrics/internal/handler"
	"github.com/JSchatten/go-practice-metrics/internal/hashprocess"

	gzipMiddleaware "github.com/JSchatten/go-practice-metrics/internal/gzip"
	loggingMiddleware "github.com/JSchatten/go-practice-metrics/internal/logging"
	storage "github.com/JSchatten/go-practice-metrics/internal/service"

	handlers "github.com/JSchatten/go-practice-metrics/internal/handler"

	audit "github.com/JSchatten/go-practice-metrics/internal/audit"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	logZero "github.com/rs/zerolog/log"
)

// Build version of the application
var buildVersion string

// Build date of the application
var buildDate string

// Build commit of the application
var buildCommit string

func main() {

	// Функция для получения значения или "N/A"
	getValueOrNA := func(value string) string {
		if value == "" {
			return "N/A"
		}
		return value
	}

	// Вывод информации о сборке
	fmt.Printf("Build version: %s\n", getValueOrNA(buildVersion))
	fmt.Printf("Build date: %s\n", getValueOrNA(buildDate))
	fmt.Printf("Build commit: %s\n", getValueOrNA(buildCommit))

	// Вывод в консоль выполнения статиклинтера при анкомменте
	// $ make go_staticlint
	// go build -o staticlint cmd/staticlint/main.go
	// go vet -vettool=./staticlint ./cmd/... ./internal/... ./pkg/...
	// go: warning: "./pkg/..." matched no packages
	// # github.com/JSchatten/go-practice-metrics/cmd/server
	// # [github.com/JSchatten/go-practice-metrics/cmd/server]
	// cmd/server/main.go:47:6: the argument is already a string, there's no need to use fmt.Sprintf
	// cmd/server/main.go:55:3: запрещён прямой вызов os.Exit в функции main пакета main
	// make: *** [Makefile:115: go_staticlint] Error 1

	// s := "hello"
	// _ = fmt.Sprintf("%s", s) // ← должен поймать S1025
	// Поймали the argument is already a string, there's no need to use fmt.Sprintf

	cfg, err := config.InitServerFlags()

	if err != nil {
		log.Fatal(err)
		// fmt.Println(err)
		// os.Exit(1)
	}

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logZero.Logger = logZero.Output(zerolog.ConsoleWriter{Out: log.Writer()})

	storageObj, err := storage.NewMemStorage(
		cfg.ServerFileFlags.FilePath,
		cfg.ServerFileFlags.FileInterval,
		cfg.ServerFileFlags.FileIsRestore,
		cfg.PostgresDSN,
	)

	if err != nil {
		logZero.Logger.Fatal().Err(err).Msg(
			"Failed to create storage",
		)
	}

	gin.DefaultWriter = io.Discard
	router := gin.New()

	// ВНЕЗАПНО начали отправлять запросы
	// в 14й итерации на окончание слеша something/
	// отключаем редирект
	router.RedirectFixedPath = false

	// Проверим, включён ли аудит
	var auditManager *audit.AuditManager
	if cfg.ServerAuditFlags.AuditFilePath != "" || cfg.ServerAuditFlags.AuditURL != "" {
		auditManager = audit.NewAuditManager(logZero.Logger)
		if cfg.ServerAuditFlags.AuditFilePath != "" {
			auditManager.Register(audit.NewFileAuditObserver(cfg.ServerAuditFlags.AuditFilePath))
		}
		if cfg.ServerAuditFlags.AuditURL != "" {
			auditManager.Register(audit.NewHTTPAuditObserver(cfg.ServerAuditFlags.AuditURL))
		}
		// Подключаем middleware
		router.Use(audit.AuditMiddleware(auditManager))
	} else {
		logZero.Info().Msg("Audit is disabled: no AUDIT_FILE or AUDIT_URL provided")
	}

	// middleware
	router.Use(loggingMiddleware.LoggingMiddleware(logZero.Logger))
	router.Use(hashprocess.HashCheckMiddleware(cfg.HashKey))
	router.Use(gzipMiddleaware.GzipMiddleware())
	// Устанавливаем ключ шифрования в контекст
	if cfg.CryptoKey != "" {
		router.Use(func(c *gin.Context) {
			c.Set("cryptoKey", cfg.CryptoKey)
			c.Next()
		})
	}
	// Добавляем middleware для проверки IP-адреса
	if cfg.TrustedSubnet != "" {
		logZero.Info().Msg("Trusted Subnet is enabled")
		router.Use(handlers.IPCheckMiddleware(cfg.TrustedSubnet))
	}
	// routes
	router.POST("/update/", handlers.UpdateHandler(storageObj))
	router.POST("/updates", handlers.UpdateHandlerBatchJSON(storageObj))
	router.POST("/value/", handlers.ValueHandler(storageObj))
	router.GET("/ping", handlers.PingDatabaseHandler(storageObj))
	router.GET("/", handlers.RootHandler(storageObj))

	// Запуск сервера в отдельной горутине
	srv := &http.Server{
		Addr:    cfg.ServerAddr,
		Handler: router,
	}
	go func() {
		logZero.Logger.Info().Msgf("Server starting at %s", cfg.ServerAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logZero.Logger.Fatal().Err(err).Msg("Server failed to start")
		}
	}()

	// Перехват сигналов завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-quit

	logZero.Logger.Info().Msg("Shutting down server...")

	// Контекст для graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Останавливаем сервер
	if err := srv.Shutdown(ctx); err != nil {
		logZero.Logger.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	// Сохраняем данные ПОСЛЕ остановки сервера
	logZero.Logger.Info().Msg("Saving metrics to file before exit...")
	if err := storageObj.SaveToFile(); err != nil {
		logZero.Logger.Error().Err(err).Msg("Failed to save metrics on exit")
	} else {
		logZero.Logger.Info().Msg("Metrics saved successfully")
	}

	logZero.Logger.Info().Msg("Server exited gracefully")

}
