package dto

import (
	"telephony/internal/domain"
	"telephony/models"
	"time"

	"github.com/go-openapi/strfmt"
)

// Метрики: только модели из Swagger (models.CallMetricsRequest, models.CallMetricsResponse и т.д.). Собственных DTO нет.

// CallMetricsRequestToFromTo конвертирует date_from/date_to из request в time.Time.
// Вызывать только после проверки в transport, что оба поля заданы.
func CallMetricsRequestToFromTo(req *models.CallMetricsRequest) (from, to time.Time) {
	if req == nil {
		return from, to
	}
	if req.DateFrom != nil {
		from = time.Time(*req.DateFrom)
	}
	if req.DateTo != nil {
		to = time.Time(*req.DateTo).Add(24*time.Hour - time.Nanosecond)
	}
	return from, to
}

// CallMetricsDomainToModel конвертирует domain.CallMetrics и domain.CallMetricsTimeseries в models.CallMetricsResponse.
func CallMetricsDomainToModel(summary *domain.CallMetrics, timeseries *domain.CallMetricsTimeseries) *models.CallMetricsResponse {
	if summary == nil {
		return &models.CallMetricsResponse{
			Summary:    &models.CallMetricsSummary{ByDirection: make(map[string]int32)},
			Timeseries: []*models.CallMetricsPoint{},
		}
	}
	byDir := make(map[string]int32, len(summary.ByDirection))
	for k, v := range summary.ByDirection {
		byDir[string(k)] = int32(v)
	}
	res := &models.CallMetricsResponse{
		Summary: &models.CallMetricsSummary{
			Total:       int32(summary.Total),
			Answered:    int32(summary.Answered),
			Missed:      int32(summary.Missed),
			ByDirection: byDir,
		},
		Timeseries: []*models.CallMetricsPoint{},
	}
	if timeseries != nil && len(timeseries.Points) > 0 {
		for _, p := range timeseries.Points {
			if p == nil {
				continue
			}
			res.Timeseries = append(res.Timeseries, &models.CallMetricsPoint{
				Date:     strfmt.Date(p.Date),
				Total:    int32(p.Total),
				Answered: int32(p.Answered),
				Missed:   int32(p.Missed),
			})
		}
	}
	return res
}
