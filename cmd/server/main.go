package main

import (
	"flag"
	"fmt"
	"os"

	handlers "github.com/JSchatten/go-practice-metrics/internal/handler"
	storage "github.com/JSchatten/go-practice-metrics/internal/service"
	"github.com/gin-gonic/gin"
)

var address string

func main() {
	flag.StringVar(&address, "a", "localhost:8080", "Server address (default: localhost:8080)")
	flag.Parse()

	if flag.NArg() > 0 {
		fmt.Fprintf(os.Stderr, "Error: unknown flags: %v\n", flag.Args())
		os.Exit(1)
	}

	storageObj := storage.NewMemStorage()
	router := gin.Default()

	router.POST("/update/:type/:name/:value", handlers.UpdateHandler(storageObj))
	router.GET("/value/:type/:name", handlers.ValueHandler(storageObj))
	router.GET("/", handlers.RootHandler(storageObj))

	fmt.Printf("Server started at %s\n", address)
	router.Run(address)
}
