package hashprocess

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

func HashCheckMiddleware(hashSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if hashSecret == "" {
			c.Next() // Скипаем по условию
			return
		}

		// clientHash := c.Request.Header.Get("HashSHA256")
		clientHash := c.Request.Header.Get("Hash")

		if clientHash == "" {
			// fmt.Println("Client hash is empty")
			// c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			// 	"error": "HashSHA256 header is required",
			// 	"code":  http.StatusBadRequest,
			// })
			c.Next()
			return
		}
		// fmt.Println("clientHash", clientHash)
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "Failed to read request body",
				"code":  http.StatusBadRequest,
				"err":   err,
			})
			fmt.Println(err)
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		if !Verify(body, hashSecret, clientHash) {
			// c.AbortWithStatus(http.StatusBadRequest)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "Invalid hash",
				"code":  http.StatusBadRequest,
				"err":   err,
			})
			fmt.Println(err)
			return
		}

		// Перехват ответа с записью хеша
		writer := &hashResponseWriter{
			body:           &bytes.Buffer{},
			ResponseWriter: c.Writer,
			hashSecret:     hashSecret,
		}
		c.Writer = writer

		c.Next()
	}
}
