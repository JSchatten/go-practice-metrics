package agent

import MetricsModel "github.com/JSchatten/go-practice-metrics/internal/model"

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
