package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/JSchatten/go-practice-metrics/internal/config"
	handlers "github.com/JSchatten/go-practice-metrics/internal/handler"
	loggingMiddleware "github.com/JSchatten/go-practice-metrics/internal/logging"

	// gzipMiddleware "github.com/JSchatten/go-practice-metrics/internal/gzip"
	storage "github.com/JSchatten/go-practice-metrics/internal/service"
	"github.com/gin-contrib/gzip"
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

	storageObj := storage.NewMemStorage()

	gin.DefaultWriter = io.Discard
	router := gin.New()
	router.Use(loggingMiddleware.LoggingMiddleware(logZero.Logger))
	// router.Use(gzipMiddleware.GzipMiddleware())
	// router.Use(gzipMiddleware.Gzip(1))
	router.Use(gzip.Gzip(gzip.DefaultCompression))

	router.POST("/update/:type/:name/:value", handlers.UpdateHandler(storageObj))
	router.GET("/value/:type/:name", handlers.ValueHandler(storageObj))
	router.POST("/update", handlers.UpdateHandlerJSON(storageObj))
	router.POST("/value", handlers.ValueHandlerJSON(storageObj))
	router.GET("/", handlers.RootHandler(storageObj))

	logZero.Logger.Info().Msgf("Server started at %s\n", cfg.ServerAddr)
	router.Run(cfg.ServerAddr)
}
