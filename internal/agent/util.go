package agent

import (
	"net"

	MetricsModel "github.com/JSchatten/go-practice-metrics/internal/model"
)

func addError(errorsUpdating *[]error, err error) {
	if err != nil {
		*errorsUpdating = append(*errorsUpdating, err)
	}
}

func getMetricGauge(id string, value float64) *MetricsModel.Metrics {
	return &MetricsModel.Metrics{
		ID:    id,
		MType: MetricsModel.Gauge,
		Value: &value,
	}
}

func getMetricCount(id string, delta int64) *MetricsModel.Metrics {
	return &MetricsModel.Metrics{
		ID:    id,
		MType: MetricsModel.Counter,
		Delta: &delta,
	}
}

// getLocalIP возвращает локальный IP-адрес агента
func getLocalIP() string {
	// Пытаемся получить IP через подключение
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "127.0.0.1"
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}
