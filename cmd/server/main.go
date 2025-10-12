package main

import (
	"fmt"

	handlers "github.com/JSchatten/go-practice-metrics/internal/handler"
	storage "github.com/JSchatten/go-practice-metrics/internal/service"
	"github.com/gin-gonic/gin"
)

func main() {
	storageObj := storage.NewMemStorage()
	router := gin.Default()

	router.POST("/update/:type/:name/:value", handlers.UpdateHandler(storageObj))
	router.GET("/value/:type/:name", handlers.ValueHandler(storageObj))
	router.GET("/", handlers.RootHandler(storageObj))

	fmt.Println("Server started at http://localhost:8080")
	router.Run(":8080")
}
