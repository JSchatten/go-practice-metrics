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

type AuditEvent struct {
	Timestamp int64    `json:"ts"`
	Metrics   []string `json:"metrics"`
	IPAddress string   `json:"ip_address"`
}

type AuditObserver interface {
	OnAuditEvent(event AuditEvent)
}

type FileAuditObserver struct {
	filePath string
	mu       sync.Mutex
	client   *http.Client
}

func NewFileAuditObserver(filePath string) *FileAuditObserver {
	return &FileAuditObserver{
		filePath: filePath,
		client:   &http.Client{Timeout: 5 * time.Second},
	}
}

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

type HTTPAuditObserver struct {
	url    string
	client *http.Client
}

func NewHTTPAuditObserver(url string) *HTTPAuditObserver {
	return &HTTPAuditObserver{
		url:    url,
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

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

type AuditManager struct {
	observers []AuditObserver
	logger    zerolog.Logger
}

func NewAuditManager(logger zerolog.Logger) *AuditManager {
	return &AuditManager{
		observers: make([]AuditObserver, 0),
		logger:    logger,
	}
}

func (am *AuditManager) Register(observer AuditObserver) {
	if observer != nil {
		am.observers = append(am.observers, observer)
	}
}

func (am *AuditManager) Notify(event AuditEvent) {
	if len(am.observers) == 0 {
		return
	}
	for _, obs := range am.observers {
		obs.OnAuditEvent(event)
	}
}
