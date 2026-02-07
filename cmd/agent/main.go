// Package main - точка входа агента сбора метрик.
//
// Агент:
//   - Регулярно собирает системные метрики (память, CPU и др.)
//   - Отправляет их на сервер по HTTP
//   - Поддерживает пакетную отправку и Gzip-сжатие
//   - Проверяет и добавляет хеши (если задан ключ)
package main

import (
	"os"
	"os/signal"
	"syscall"

	"log"

	agentInternal "github.com/JSchatten/go-practice-metrics/internal/agent"
	config "github.com/JSchatten/go-practice-metrics/internal/config"

	"github.com/rs/zerolog"
	logZero "github.com/rs/zerolog/log"
)

func main() {

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logZero.Logger = logZero.Output(zerolog.ConsoleWriter{Out: log.Writer()})

	cfg, err := config.InitAgentFlags()
	if err != nil {
		logZero.Logger.Fatal().Err(err).Msg("Failed start agent InitAgentFlags")
	}

	agentInst, err := agentInternal.NewAgent(cfg)

	if err != nil {
		logZero.Logger.Fatal().Err(err).Msg("Failed start agent NewAgent")
	}
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	// Запуск сбора и отправки метрик
	go func() {
		agentInst.Start()
	}()

	<-done
	agentInst.Stop()

	logZero.Logger.Info().Msg("Agent processed")
}
