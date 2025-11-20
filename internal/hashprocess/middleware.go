package hashprocess

import (
	"bytes"
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

		clientHash := c.Request.Header.Get("HashSHA256")
		if clientHash == "" {
			// fmt.Println("Client hash is empty")
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		// fmt.Println("clientHash", clientHash)
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		if !Verify(body, hashSecret, clientHash) {
			c.AbortWithStatus(http.StatusBadRequest)
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
