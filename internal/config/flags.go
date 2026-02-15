// Package config предоставляет структуры и функции для инициализации и обработки флагов командной строки и переменных окружения
// для сервисов агента и сервера метрик.
//
// Пакет содержит:
//   - Структуры флагов: AgentFlags, ServerFlags, ServerFileFlags, ServerAuditFlags.
//   - Константы по умолчанию для флагов.
//   - Функции инициализации: InitAgentFlags, InitServerFlags.
//   - Обработку переменных окружения и приоритет флагов над env.
//   - Валидацию значений флагов.
package config

const (
	// common
	constServerAddr = "localhost:8080"
	// agent
	constPollInterval   = 2
	constReportInterval = 10
	constRateLimit      = 10
	// server
	constFilePath        = "./metrics.json"
	constFileIntervalSec = 300
	constRestoreFromFile = true
)
