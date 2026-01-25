package handler

import (
	"encoding/json"
	"net/http"

	model "github.com/JSchatten/go-practice-metrics/internal/model"
	storage "github.com/JSchatten/go-practice-metrics/internal/service"
	"github.com/gin-gonic/gin"
	logZero "github.com/rs/zerolog/log"
)

func UpdateHandler(storage storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		logZero.Logger.Info().Msg("UpdateHandlerJSON")
		var metricIn model.Metrics

		if c.Request.Method != http.MethodPost {
			MethodNotAllowed(c)
			return
		}

		// Это по-хорошему, т.к. есть sonic-avx, для build-тегов, но требования такие
		// if err := c.ShouldBindJSON(&metricIn); err != nil {
		// 	c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		// 	return
		// }

		decoder := json.NewDecoder(c.Request.Body)
		if err := decoder.Decode(&metricIn); err != nil {
			BodyInvalidJSON(c)
			return
		}

		// Валидация через метод
		if err := metricIn.Validate(); err != nil {
			logZero.Logger.Error().Err(err).Msgf("Validate error %+v", metricIn)
			switch err {
			case model.ErrEmptyMetricID, model.ErrEmptyMetricType:
				BodyMissingFields(c)
			case model.ErrDeltaRequired:
				DeltaNotProvided(c)
			case model.ErrValueRequired:
				ValueNotProvided(c)
			case model.ErrUnknownMetricType:
				BodyInvalidMetricType(c)
			default:
				BadRequestVerbose(c, err)
			}
			return
		}

		err := storage.UpdateMetric(c, &metricIn)
		if err != nil {
			FailedToUpdateMetric(c, err)
			return
		}
		// Добавляем метрики в контекст для аудита
		c.Set("audit.metrics", []string{metricIn.ID})

		c.JSON(http.StatusOK, metricIn)
		logZero.Logger.Info().Msgf("UpdateHandlerJSON metrics in = %s", metricIn.String())
	}
}

func UpdateHandlerBatchJSON(storage storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		logZero.Logger.Info().Msg("UpdateHandlerBatchJSON")
		if c.Request.Method != http.MethodPost {
			MethodNotAllowed(c)
			return
		}

		var metricsIn []model.Metrics
		decoder := json.NewDecoder(c.Request.Body)
		if err := decoder.Decode(&metricsIn); err != nil {
			BodyInvalidJSON(c)
			return
		}

		var metricNames []string
		// Валидация через метод
		for _, elem := range metricsIn {
			if err := elem.Validate(); err != nil {
				logZero.Logger.Error().Err(err).Msgf("Validate error %+v", elem)
				switch err {
				case model.ErrEmptyMetricID, model.ErrEmptyMetricType:
					BodyMissingFields(c)
				case model.ErrDeltaRequired:
					DeltaNotProvided(c)
				case model.ErrValueRequired:
					ValueNotProvided(c)
				case model.ErrUnknownMetricType:
					BodyInvalidMetricType(c)
				default:
					BadRequestVerbose(c, err)
				}
				// Прерываем, если в батче есть невалидные данные
				return
			}
			// Тут метрики добавляются, но вызов Set позже,
			// чтобы прервать в случа некорректного запроса
			metricNames = append(metricNames, elem.ID)
		}
		c.Set("audit.metrics", metricNames)

		err := storage.UpdateMetricBatch(c, &metricsIn)
		if err != nil {
			FailedToUpdateMetric(c, err)
			return
		}

		c.JSON(http.StatusOK, metricsIn)
	}
}
