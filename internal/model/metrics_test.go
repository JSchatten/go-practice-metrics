package model

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetrics_String(t *testing.T) {
	t.Run("gauge - valid", func(t *testing.T) {
		value := 3.1415
		metric := &Metrics{
			ID:    "cpu_usage",
			MType: Gauge,
			Value: &value,
		}

		result := metric.String()

		var parsed map[string]interface{}
		err := json.Unmarshal([]byte(result), &parsed)
		require.NoError(t, err)

		assert.Equal(t, "cpu_usage", parsed["id"])
		assert.Equal(t, "gauge", parsed["type"])
		assert.Equal(t, 3.1415, parsed["value"])
		_, hasDelta := parsed["delta"]
		assert.False(t, hasDelta, "delta should not be present")
	})

	t.Run("counter - valid", func(t *testing.T) {
		delta := int64(42)
		metric := &Metrics{
			ID:    "requests_total",
			MType: Counter,
			Delta: &delta,
		}

		result := metric.String()

		var parsed map[string]interface{}
		err := json.Unmarshal([]byte(result), &parsed)
		require.NoError(t, err)

		assert.Equal(t, "requests_total", parsed["id"])
		assert.Equal(t, "counter", parsed["type"])
		assert.Equal(t, float64(42), parsed["delta"]) // int64 -> float64 при unmarshal
		_, hasValue := parsed["value"]
		assert.False(t, hasValue, "value should not be present")
	})

	t.Run("with hash", func(t *testing.T) {
		value := 100.0
		metric := &Metrics{
			ID:    "memory",
			MType: Gauge,
			Value: &value,
			Hash:  "abc123",
		}

		result := metric.String()

		var parsed map[string]interface{}
		err := json.Unmarshal([]byte(result), &parsed)
		require.NoError(t, err)

		assert.Equal(t, "memory", parsed["id"])
		assert.Equal(t, "gauge", parsed["type"])
		assert.Equal(t, 100.0, parsed["value"])
		assert.Equal(t, "abc123", parsed["hash"])
	})

	t.Run("nil metric", func(t *testing.T) {
		var metric *Metrics = nil

		result := metric.String()

		assert.Equal(t, "<nil>", result)
	})

	t.Run("empty fields", func(t *testing.T) {
		// Хотя Validate() не пройдёт, String() должен работать
		metric := &Metrics{
			ID:    "",
			MType: "",
			Delta: nil,
			Value: nil,
		}

		result := metric.String()

		// Должен вернуть валидный JSON, даже если поля пустые
		var parsed map[string]interface{}
		err := json.Unmarshal([]byte(result), &parsed)
		require.NoError(t, err)

		assert.Equal(t, "", parsed["id"])
		assert.Equal(t, "", parsed["type"])
		_, hasDelta := parsed["delta"]
		assert.False(t, hasDelta)
		_, hasValue := parsed["value"]
		assert.False(t, hasValue)
	})

	t.Run("zero values - should be included Counter", func(t *testing.T) {
		delta := int64(0)
		metric := &Metrics{
			ID:    "test_zero",
			MType: Counter,
			Delta: &delta,
		}

		result := metric.String()

		var parsed map[string]interface{}
		err := json.Unmarshal([]byte(result), &parsed)
		require.NoError(t, err)

		assert.Equal(t, "test_zero", parsed["id"])
		assert.Equal(t, "counter", parsed["type"])
		assert.Equal(t, float64(0), parsed["delta"]) // 0 — это валидно

	})

	t.Run("zero values - should be included Gauge", func(t *testing.T) {
		value := 0.0
		metric := &Metrics{
			ID:    "test_zero",
			MType: Counter,
			Value: &value,
		}

		result := metric.String()

		var parsed map[string]interface{}
		err := json.Unmarshal([]byte(result), &parsed)
		require.NoError(t, err)

		assert.Equal(t, "test_zero", parsed["id"])
		assert.Equal(t, "counter", parsed["type"])
		assert.Equal(t, float64(0), parsed["value"])
	})
}

func TestNewMetrics_ValidGauge(t *testing.T) {
	metric, err := NewMetrics("cpu_load", "gauge", "99.5")
	require.NoError(t, err)
	assert.Equal(t, "cpu_load", metric.ID)
	assert.Equal(t, "gauge", metric.MType)
	assert.NotNil(t, metric.Value)
	assert.Nil(t, metric.Delta)
	assert.InDelta(t, 99.5, *metric.Value, 1e-6)
}

func TestNewMetrics_ValidCounter(t *testing.T) {
	metric, err := NewMetrics("requests", "counter", "123")
	require.NoError(t, err)
	assert.Equal(t, "requests", metric.ID)
	assert.Equal(t, "counter", metric.MType)
	assert.NotNil(t, metric.Delta)
	assert.Nil(t, metric.Value)
	assert.Equal(t, int64(123), *metric.Delta)
}

func TestNewMetrics_CounterZeroValue(t *testing.T) {
	// Значение "0" — валидно
	metric, err := NewMetrics("requests", "counter", "0")
	require.NoError(t, err)
	assert.Equal(t, int64(0), *metric.Delta)
}

func TestNewMetrics_GaugeZeroValue(t *testing.T) {
	metric, err := NewMetrics("cpu", "gauge", "0.0")
	require.NoError(t, err)
	assert.InDelta(t, 0.0, *metric.Value, 1e-6)
}

func TestNewMetrics_EmptyID(t *testing.T) {
	metric, err := NewMetrics("", "gauge", "1.0")
	assert.Error(t, err)
	assert.Nil(t, metric)
	assert.Equal(t, ErrEmptyMetricID, err)
}

func TestNewMetrics_EmptyType(t *testing.T) {
	metric, err := NewMetrics("cpu", "", "1.0")
	assert.Error(t, err)
	assert.Nil(t, metric)
	assert.Equal(t, ErrEmptyMetricType, err)
}

func TestNewMetrics_UnknownType(t *testing.T) {
	metric, err := NewMetrics("cpu", "unknown", "1.0")
	assert.Error(t, err)
	assert.Nil(t, metric)
	assert.Equal(t, ErrUnknownMetricType, err)
}

func TestNewMetrics_MissingValue(t *testing.T) {
	metric, err := NewMetrics("cpu", "gauge", "")
	assert.Error(t, err)
	assert.Nil(t, metric)
	assert.Equal(t, ErrValueRequired, err)

	metric, err = NewMetrics("requests", "counter", "")
	assert.Error(t, err)
	assert.Nil(t, metric)
	assert.Equal(t, ErrValueRequired, err)
}

func TestNewMetrics_InvalidCounterValue(t *testing.T) {
	metric, err := NewMetrics("requests", "counter", "abc")
	assert.Error(t, err)
	assert.Nil(t, metric)
	assert.Equal(t, ErrInvalidCounterValue, err)
}

func TestNewMetrics_InvalidGaugeValue(t *testing.T) {
	metric, err := NewMetrics("cpu", "gauge", "xyz")
	assert.Error(t, err)
	assert.Nil(t, metric)
	assert.Equal(t, ErrInvalidGaugeValue, err)
}

func TestMetrics_Validate_ValidGauge(t *testing.T) {
	metric := Metrics{
		ID:    "cpu_load",
		MType: "gauge",
		Value: floatPtr(1.5),
	}
	err := metric.Validate()
	assert.NoError(t, err)
}

func TestMetrics_Validate_ValidCounter(t *testing.T) {
	metric := Metrics{
		ID:    "requests",
		MType: "counter",
		Delta: intPtr(100),
	}
	err := metric.Validate()
	assert.NoError(t, err)
}

func TestMetrics_Validate_EmptyID(t *testing.T) {
	metric := Metrics{
		ID:    "",
		MType: "gauge",
		Value: floatPtr(1.0),
	}
	err := metric.Validate()
	assert.Error(t, err)
	assert.Equal(t, ErrEmptyMetricID, err)
}

func TestMetrics_Validate_EmptyType(t *testing.T) {
	metric := Metrics{
		ID:    "cpu",
		MType: "",
		Value: floatPtr(1.0),
	}
	err := metric.Validate()
	assert.Error(t, err)
	assert.Equal(t, ErrEmptyMetricType, err)
}

func TestMetrics_Validate_MissingValue(t *testing.T) {
	metric := Metrics{
		ID:    "cpu",
		MType: "gauge",
		Value: nil,
	}
	err := metric.Validate()
	assert.Error(t, err)
	assert.Equal(t, ErrValueRequired, err)
}

func TestMetrics_Validate_MissingDelta(t *testing.T) {
	metric := Metrics{
		ID:    "requests",
		MType: "counter",
		Delta: nil,
	}
	err := metric.Validate()
	assert.Error(t, err)
	assert.Equal(t, ErrDeltaRequired, err)
}

func TestMetrics_Validate_UnknownType(t *testing.T) {
	metric := Metrics{
		ID:    "unknown",
		MType: "unknown_type",
		Delta: intPtr(1),
	}
	err := metric.Validate()
	assert.Error(t, err)
	assert.Equal(t, ErrUnknownMetricType, err)
}

func TestMetrics_Validate_NilMetric(t *testing.T) {
	var metric *Metrics
	err := metric.Validate() // должна быть проверка на nil
	assert.Error(t, err)
	assert.Equal(t, ErrNilMetric, err) // потому что ID == ""
}

// Вспомогательные функции
func intPtr(i int64) *int64 {
	return &i
}

func floatPtr(f float64) *float64 {
	return &f
}
