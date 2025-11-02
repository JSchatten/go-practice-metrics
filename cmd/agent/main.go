package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	agentInternal "github.com/JSchatten/go-practice-metrics/internal/agent"
	config "github.com/JSchatten/go-practice-metrics/internal/config"
	storage "github.com/JSchatten/go-practice-metrics/internal/service"
)

func main() {

	cfg, err := config.InitAgentFlags()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	storage := storage.NewMemStorage()
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

	fmt.Println("Agent processed")
}
