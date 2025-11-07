package main

import (
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"log"

	agentInternal "github.com/JSchatten/go-practice-metrics/internal/agent"
	config "github.com/JSchatten/go-practice-metrics/internal/config"
	storage "github.com/JSchatten/go-practice-metrics/internal/service"

	"github.com/rs/zerolog"
	logZero "github.com/rs/zerolog/log"
)

func main() {

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logZero.Logger = logZero.Output(zerolog.ConsoleWriter{Out: log.Writer()})

	cfg, err := config.InitAgentFlags()

	if err != nil {
		logZero.Logger.Fatal().Err(err).Msg("Failed start agent agentFlags")
	}

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	storage, err := storage.NewMemStorage(os.DevNull, 0, false)
	if err != nil {
		logZero.Logger.Fatal().Err(err).Msg("Failed start agent memStorage")
	}
	done := make(chan struct{})

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Горутина для обработки сигналов
	go func() {
		<-sigChan
		close(done)
	}()

	// Запуск сбора и отправки метрик
	agentInternal.UpdateRuntimeMetrics(*cfg, storage, done)
	logZero.Logger.Info().Msg("Agent processed")

}
