// Package hashprocess предоставляет middleware для проверки и установки SHA256-хэша тела HTTP-запроса и ответа.
//
// Основные функции:
//   - HashCheckMiddleware: проверяет хэш тела запроса из заголовка HashSHA256 при наличии секретного ключа.
//   - Записывает хэш тела ответа в заголовок HashSHA256 при отправке.
//
// Используется для обеспечения целостности данных между агентом и сервером.
package hashprocess

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// var hashSecretHandler string

func HashCheckMiddleware(hashSecret string) gin.HandlerFunc {

	// hashSecretHandler = hashSecret

	// fmt.Println("HashCheckMiddleware.hashSecret", hashSecret)
	// fmt.Println("HashCheckMiddleware.hashSecret", hashSecret)
	// fmt.Println("HashCheckMiddleware.hashSecret", hashSecret)

	// fmt.Println("HashCheckMiddleware.hashSecretHandler", hashSecretHandler)
	// fmt.Println("HashCheckMiddleware.hashSecretHandler", hashSecretHandler)
	// fmt.Println("HashCheckMiddleware.hashSecretHandler", hashSecretHandler)

	return func(c *gin.Context) {
		if hashSecret == "" {
			c.Next() // Скипаем по условию
			return
		}

		// clientHash := r.Header.Get("HashSHA256")
		// if clientHash == "" {
		// 	next.ServeHTTP(w, r)
		// 	return
		// }

		// bodyBytes, err := io.ReadAll(r.Body)
		// if err != nil {
		// 	logger.Log.Errorf("Error reading body: %s", err)
		// 	next.ServeHTTP(w, r)
		// 	return
		// }

		// r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// s := hmac.New(sha256.New, []byte(h.cfg.Key))
		// s.Write(bodyBytes)
		// bodyHash := s.Sum(nil)

		// data, err := hex.DecodeString(clientHash)
		// if err != nil {
		// 	logger.Log.Errorf("Error decoding client hash: %s", err)
		// 	return
		// }

		// if !hmac.Equal(bodyHash, data) {
		// 	logger.Log.Errorf("Auth Token Failed")
		// 	//w.WriteHeader(http.StatusBadRequest)
		// 	next.ServeHTTP(w, r)
		// 	return
		// }

		// clientHash := c.Request.Header.Get("HashSHA256")
		clientHash := c.Request.Header.Get("HashSHA256")

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
			// fmt.Println("HashCheckMiddleware.body", body)
			// fmt.Println("HashCheckMiddleware.hashSecret", hashSecret)
			// fmt.Println("HashCheckMiddleware.clientHash", clientHash)
			// fmt.Println("HashCheckMiddleware.err", err)

			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "Invalid hash",
				"code":  http.StatusBadRequest,
				"err":   err,
			})
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
