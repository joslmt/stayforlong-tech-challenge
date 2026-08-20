package application

import (
	"context"

	"github.com/joslmt/stayforlong-tech-challenge/internal/booking/domain"
)

type MetricsResult struct {
	Metrics domain.NightMetrics
}

type MetricsService struct{}

func NewMetricsService() MetricsService { return MetricsService{} }

func (MetricsService) Calculate(ctx context.Context, batch domain.BookingBatch) (MetricsResult, error) {
	if err := ctx.Err(); err != nil {
		return MetricsResult{}, err
	}
	if err := batch.Validate(); err != nil {
		return MetricsResult{}, err
	}
	return MetricsResult{Metrics: domain.CalculateNightMetrics(batch.Bookings())}, nil
}
