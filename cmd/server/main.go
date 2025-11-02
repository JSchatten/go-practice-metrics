package main

import (
	"fmt"
	"os"

	"github.com/JSchatten/go-practice-metrics/internal/config"
	handlers "github.com/JSchatten/go-practice-metrics/internal/handler"
	storage "github.com/JSchatten/go-practice-metrics/internal/service"
	"github.com/gin-gonic/gin"
)

func main() {

	cfg, err := config.InitServerFlags()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	storageObj := storage.NewMemStorage()
	router := gin.Default()

	router.POST("/update/:type/:name/:value", handlers.UpdateHandler(storageObj))
	router.GET("/value/:type/:name", handlers.ValueHandler(storageObj))
	router.GET("/", handlers.RootHandler(storageObj))

	fmt.Printf("Server started at %s\n", cfg.ServerAddr)
	router.Run(cfg.ServerAddr)
}
