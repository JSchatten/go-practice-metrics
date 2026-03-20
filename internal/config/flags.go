// Package config предоставляет структуры и функции для инициализации и обработки флагов командной строки, переменных окружения
// и конфигурационных файлов JSON для сервисов агента и сервера метрик.
//
// Пакет содержит:
//   - Структуры флагов: AgentFlags, ServerFlags, ServerFileFlags, ServerAuditFlags.
//   - Структуры конфигурации: ServerConfig, AgentConfig.
//   - Константы по умолчанию для флагов.
//   - Функции инициализации: InitAgentFlags, InitServerFlags.
//   - Обработку переменных окружения, конфигурационных файлов и приоритет: флаги > env > файл конфигурации.
//   - Валидацию значений флагов.
package config

const (
	// common
	constServerAddr = "localhost:8080"
	// agent
	constPollInterval   = 2
	constReportInterval = 10
	constRateLimit      = 10
	constGRPCServerAddr = "localhost:9000"
	// server
	constFilePath        = "./metrics.json"
	constFileIntervalSec = 300
	constRestoreFromFile = true
	constGRPCServerPort  = 9000
)
