// Package audit реализует систему аудита действий в приложении.
//
// Предоставляет:
//   - Гибкий механизм наблюдения за событиями (Observer Pattern)
//   - Поддержку нескольких бэкендов: файл и HTTP
//   - Асинхронную отправку событий
//
// Основные компоненты:
//   - AuditEvent - структура события аудита
//   - AuditObserver - интерфейс наблюдателя
//   - AuditManager - центральный диспетчер уведомлений
//   - Реализации: FileAuditObserver, HTTPAuditObserver
//
// Использование:
//   - Регистрируются наблюдатели (например, файл или HTTP-эндпоинт)
//   - Middleware собирает данные и передаёт их через AuditManager
//   - События логируются в файл или отправляются по HTTP
package audit

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"

	logZero "github.com/rs/zerolog/log"
)

// AuditEvent представляет событие аудита - факт изменения метрик.
//
// Поля:
//   - Timestamp: время события в Unix-секундах
//   - Metrics: список имён метрик, задействованных в запросе
//   - IPAddress: IP-адрес клиента
//
// Используется всеми реализациями AuditObserver.
type AuditEvent struct {
	Timestamp int64    `json:"ts"`
	Metrics   []string `json:"metrics"`
	IPAddress string   `json:"ip_address"`
}

// AuditObserver - интерфейс, который должен реализовывать каждый наблюдатель аудита.
//
// Реализации:
//   - FileAuditObserver: запись в файл
//   - HTTPAuditObserver: отправка POST-запроса
//
// Метод OnAuditEvent вызывается AuditManager при наступлении события.
type AuditObserver interface {
	OnAuditEvent(event AuditEvent)
}

// FileAuditObserver - реализация наблюдателя, сохраняющего события в файл.
//
// Поведение:
//   - Каждое событие записывается как JSON-строка в указанный файл
//   - Использует мьютекс для потокобезопасности
//   - При ошибке записи логирует ошибку через zerolog
type FileAuditObserver struct {
	filePath string
	mu       sync.Mutex
	client   *http.Client
}

// NewFileAuditObserver создаёт новый наблюдатель, пишущий в файл.
//
// Параметр filePath - путь к файлу, куда будут записываться события.
func NewFileAuditObserver(filePath string) *FileAuditObserver {
	return &FileAuditObserver{
		filePath: filePath,
		client:   &http.Client{Timeout: 5 * time.Second},
	}
}

// OnAuditEvent сериализует событие в JSON и записывает в файл.
//
// Поведение:
//   - Блокирует доступ к файлу с помощью мьютекса
//   - Открывает файл в режиме добавления (O_APPEND)
//   - Записывает строку и переводит каретку
func (f *FileAuditObserver) OnAuditEvent(event AuditEvent) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	file, err := os.OpenFile(f.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logZero.Error().Err(err).Msg("AUDIT [FILE] Failed to open file")
		return
	}
	defer file.Close()

	_, err = file.WriteString(string(data) + "\n")
	if err != nil {
		logZero.Error().Err(err).Msg("AUDIT [FILE] Failed to write")
	}
}

// HTTPAuditObserver - реализация наблюдателя, отправляющего события по HTTP.
//
// Поведение:
//   - Отправляет POST-запрос с JSON-телом на указанный URL
//   - Использует отдельную горутину для асинхронной отправки
//   - Имеет таймаут 5 секунд
//
// Полезно для интеграции с внешними системами мониторинга
type HTTPAuditObserver struct {
	url    string
	client *http.Client
}

// NewHTTPAuditObserver создаёт наблюдатель, отправляющий события на указанный URL.
//
// Параметр url - адрес, куда отправляются POST-запросы (должен принимать JSON).
//
// Пример:
//
//	observer := audit.NewHTTPAuditObserver("https://analytics.example.com/audit")
func NewHTTPAuditObserver(url string) *HTTPAuditObserver {
	return &HTTPAuditObserver{
		url:    url,
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

// OnAuditEvent отправляет событие на удалённый сервер в отдельной горутине.
//
// Поведение:
//   - Событие сериализуется в JSON
//   - Запускается горутина для отправки POST-запроса
//   - Ответ сервера игнорируется (тело закрывается)
//
// Ошибки:
//   - Логируются через zerolog
//   - Не блокируют основной поток
//
// Примечание: если сервер недоступен, запрос падает - повторов нет.
func (h *HTTPAuditObserver) OnAuditEvent(event AuditEvent) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	go func() {
		resp, err := h.client.Post(h.url, "application/json", bytes.NewReader(data))
		if err != nil {
			logZero.Error().Err(err).Msg("AUDIT [HTTP] Request failed")
			return
		}
		resp.Body.Close()
	}()
}

// AuditManager - центральный менеджер аудита.
//
// Управляет наблюдателями и рассылает события.
// Потокобезопасен при регистрации, но не при уведомлении (наблюдатели сами должны быть thread-safe).
type AuditManager struct {
	observers []AuditObserver
	logger    zerolog.Logger
}

// NewAuditManager создаёт новый менеджер аудита.
//
// Принимает логгер, который можно использовать в будущем.
//
// Пример:
//
//	auditManager := audit.NewAuditManager(logZero.Logger)
func NewAuditManager(logger zerolog.Logger) *AuditManager {
	return &AuditManager{
		observers: make([]AuditObserver, 0),
		logger:    logger,
	}
}

// Register добавляет нового наблюдателя в менеджер.
//
// Если observer == nil - игнорируется.
//
// Пример:
//
//	manager.Register(audit.NewFileAuditObserver("/var/log/audit.log"))
//	manager.Register(audit.NewHTTPAuditObserver("https://example.com/audit"))
func (am *AuditManager) Register(observer AuditObserver) {
	if observer != nil {
		am.observers = append(am.observers, observer)
	}
}

// Notify рассылает событие всем зарегистрированным наблюдателям.
//
// Если наблюдателей нет - ничего не делает.
//
// Вызовы OnAuditEvent происходят последовательно.
// Если один наблюдатель падает - другие всё равно вызываются.
func (am *AuditManager) Notify(event AuditEvent) {
	if len(am.observers) == 0 {
		return
	}
	for _, obs := range am.observers {
		obs.OnAuditEvent(event)
	}
}
