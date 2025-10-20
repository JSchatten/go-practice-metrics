package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	agentInternal "github.com/JSchatten/go-practice-metrics/internal/agent"
	storage "github.com/JSchatten/go-practice-metrics/internal/service"
)

var (
	address        string
	reportInterval int
	pollInterval   int
)

func init() {
	flag.StringVar(&address, "a", "localhost:8080", "Server address (default: localhost:8080)")
	flag.IntVar(&reportInterval, "r", 10, "Report interval in seconds (default: 10)")
	flag.IntVar(&pollInterval, "p", 2, "Poll interval in seconds (default: 2)")
}

func main() {
	flag.Parse()

	if flag.NArg() > 0 {
		fmt.Fprintf(os.Stderr, "Error: unknown flags: %v\n", flag.Args())
		os.Exit(1)
	}

	// Проверка корректности интервалов
	if reportInterval <= 0 {
		fmt.Fprintf(os.Stderr, "Error: reportInterval must be positive\n")
		os.Exit(1)
	}
	if pollInterval <= 0 {
		fmt.Fprintf(os.Stderr, "Error: pollInterval must be positive\n")
		os.Exit(1)
	}

	cfg := agentInternal.InitFlags(
		pollInterval,
		reportInterval,
		address,
	)

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
