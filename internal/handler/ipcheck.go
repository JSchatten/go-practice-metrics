package handler

import (
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	logZero "github.com/rs/zerolog/log"
)

// IPCheckMiddleware создает middleware для проверки IP-адреса клиента.
//
// Параметры:
//   - trustedSubnet: строка в формате CIDR, определяющая доверенную подсеть.
//     Если пустая, проверка отключается.
//
// Поведение:
//   - Если trustedSubnet пустая, middleware пропускает все запросы.
//   - Если trustedSubnet задана, middleware извлекает IP-адрес из заголовка X-Real-IP.
//   - Если IP-адрес не найден, возвращается ошибка 400 Bad Request.
//   - Если IP-адрес не входит в доверенную подсеть, возвращается ошибка 403 Forbidden.
//   - В противном случае запрос передается дальше.
func IPCheckMiddleware(trustedSubnet string) gin.HandlerFunc {
	var _, ipNet, err = net.ParseCIDR(trustedSubnet)
	if err != nil {
		if trustedSubnet != "" {
			logZero.Logger.Error().Err(err).Str("cidr", trustedSubnet).Msg("Failed to parse trusted subnet CIDR")
		}
		// Если trustedSubnet пустая или некорректная, отключаем проверку
		ipNet = nil
	}

	return func(c *gin.Context) {
		// Если доверенная подсеть не задана, пропускаем проверку
		if ipNet == nil {
			c.Next()
			return
		}

		// Извлекаем IP-адрес из заголовка X-Real-IP
		realIP := c.GetHeader("X-Real-IP")
		if realIP == "" {
			logZero.Logger.Warn().Msg("X-Real-IP header is missing")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "X-Real-IP header is required"})
			return
		}

		// Проверяем, что IP-адрес входит в доверенную подсеть
		if !ipNet.Contains(net.ParseIP(realIP)) {
			logZero.Logger.Warn().Str("ip", realIP).Str("subnet", trustedSubnet).Msg("Client IP is not in trusted subnet")
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "client IP is not in trusted subnet"})
			return
		}

		// IP-адрес прошел проверку, продолжаем обработку запроса
		c.Next()
	}
}
