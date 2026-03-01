// Package main - точка входа агента сбора метрик.
//
// Агент:
//   - Регулярно собирает системные метрики (память, CPU и др.)
//   - Отправляет их на сервер по HTTP
//   - Поддерживает пакетную отправку и Gzip-сжатие
//   - Проверяет и добавляет хеши (если задан ключ)
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"log"

	agentInternal "github.com/JSchatten/go-practice-metrics/internal/agent"
	config "github.com/JSchatten/go-practice-metrics/internal/config"

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
	signal.Notify(done, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	// Запуск сбора и отправки метрик
	go func() {
		agentInst.Start()
	}()

	<-done
	agentInst.Stop()

	logZero.Logger.Info().Msg("Agent processed")
}
