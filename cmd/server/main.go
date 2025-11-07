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
	handlers "github.com/JSchatten/go-practice-metrics/internal/handler"

	gzipMiddleaware "github.com/JSchatten/go-practice-metrics/internal/gzip"
	loggingMiddleware "github.com/JSchatten/go-practice-metrics/internal/logging"
	storage "github.com/JSchatten/go-practice-metrics/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	logZero "github.com/rs/zerolog/log"
)

func main() {

	cfg, err := config.InitServerFlags()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logZero.Logger = logZero.Output(zerolog.ConsoleWriter{Out: log.Writer()})

	storageObj := storage.NewMemStorage(
		cfg.ServerFileFlags.FilePath,
		cfg.ServerFileFlags.FileInterval,
		cfg.ServerFileFlags.FileIsRestore,
	)

	gin.DefaultWriter = io.Discard
	router := gin.New()
	// middleware
	router.Use(loggingMiddleware.LoggingMiddleware(logZero.Logger))
	router.Use(gzipMiddleaware.GzipMiddleware())
	// routes
	router.POST("/update/:type/:name/:value", handlers.UpdateHandler(storageObj))
	router.GET("/value/:type/:name", handlers.ValueHandler(storageObj))
	router.POST("/update", handlers.UpdateHandlerJSON(storageObj))
	router.POST("/value", handlers.ValueHandlerJSON(storageObj))
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
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
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
