package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	agentInternal "github.com/JSchatten/go-practice-metrics/internal/agent"
	storage "github.com/JSchatten/go-practice-metrics/internal/service"
)

func main() {

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
	agentInternal.UpdateRuntimeMetrics(storage, done)

	fmt.Println("Agent processed")
}
