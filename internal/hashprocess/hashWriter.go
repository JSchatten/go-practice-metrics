package hashprocess

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"

	"github.com/gin-gonic/gin"
)

func Sign(message []byte, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write(message)
	return hex.EncodeToString(h.Sum(nil))
}

func Verify(message []byte, key, signature string) bool {
	expected := Sign(message, key)
	return hmac.Equal([]byte(signature), []byte(expected))
}

type hashResponseWriter struct {
	gin.ResponseWriter
	body       *bytes.Buffer
	hashSecret string
}

func (w *hashResponseWriter) Write(data []byte) (int, error) {
	w.body.Write(data)
	// Подписываем тело ответа
	responseHash := Sign(w.body.Bytes(), w.hashSecret)
	w.ResponseWriter.Header().Set("HashSHA256", responseHash)
	return w.ResponseWriter.Write(data)
}
