package main

import (
	"fmt"
	"net/http"

	handlers "github.com/JSchatten/go-practice-metrics/internal/handler"
	MetricsModel "github.com/JSchatten/go-practice-metrics/internal/model"
	storage "github.com/JSchatten/go-practice-metrics/internal/service"
)

type Storage interface {
	UpdateMetric(metric *MetricsModel.Metrics) error
}

func main() {
	storageObj := storage.NewMemStorage()
	http.HandleFunc("/update/", handlers.UpdateHandler(storageObj))
	// http.HandleFunc("/live/", handlers.LiveHandler())
	// http.HandleFunc("/", handlers.LiveHandler())
	fmt.Println("Server started at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
