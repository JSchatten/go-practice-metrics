package middleware

// import (
// 	//"compress/gzip"
// 	//gzip "github.com/klauspost/pgzip"

// 	"compress/gzip"
// 	"net/http"
// 	"path/filepath"
// 	"strings"
// 	"sync"

// 	"github.com/gin-gonic/gin"
// )

// const (
// 	BestCompression    = gzip.BestCompression
// 	BestSpeed          = gzip.BestSpeed
// 	DefaultCompression = gzip.DefaultCompression
// 	NoCompression      = gzip.NoCompression
// )

// var (
// 	compressionLevel int
// )

// func Gzip(level int) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		if !shouldCompress(c.Request) {
// 			c.Next()
// 			return
// 		}
// 		//TODO: skip compression if websocket
// 		//TODO: skip compression if already compressed
// 		compressionLevel = level
// 		gz := gzipWriterPool.Get().(*gzip.Writer)
// 		gz.Reset(c.Writer)

// 		c.Header("Content-Encoding", "gzip")
// 		c.Header("Vary", "Accept-Encoding")
// 		c.Writer = &gzipWriter{c.Writer, gz}
// 		defer func() {
// 			c.Header("Content-Length", "")
// 			gz.Close()
// 			gzipWriterPool.Put(gz)
// 		}()
// 		c.Next()
// 	}
// }

// type gzipWriter struct {
// 	gin.ResponseWriter
// 	writer *gzip.Writer
// }

// func (g *gzipWriter) Write(data []byte) (int, error) {
// 	return g.writer.Write(data)
// }

// func shouldCompress(req *http.Request) bool {
// 	if !strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") {
// 		return false
// 	}
// 	extension := filepath.Ext(req.URL.Path)
// 	if len(extension) < 4 { // fast path
// 		return true
// 	}

// 	switch extension {
// 	case ".png", ".gif", ".jpeg", ".jpg":
// 		return false
// 	default:
// 		return true
// 	}
// }

// var gzipWriterPool = sync.Pool{
// 	New: func() interface{} {
// 		v, _ := gzip.NewWriterLevel(nil, compressionLevel)
// 		return v
// 	},
// }
