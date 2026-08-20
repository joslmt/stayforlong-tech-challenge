package application

import (
	"context"

	"github.com/joslmt/stayforlong-tech-challenge/internal/booking/domain"
)

type RevenueResult struct {
	RequestIDs  []string
	TotalProfit domain.ProfitCents
	Metrics     domain.NightMetrics
}

type ScheduleService struct {
	optimizer domain.ScheduleOptimizer
}

func NewScheduleService(optimizer domain.ScheduleOptimizer) ScheduleService {
	return ScheduleService{optimizer: optimizer}
}

func (service ScheduleService) Build(ctx context.Context, batch domain.BookingBatch) (RevenueResult, error) {
	schedule, err := service.optimizer.Optimize(ctx, batch)
	if err != nil {
		return RevenueResult{}, err
	}

	bookings := schedule.Bookings()
	requestIDs := make([]string, 0, len(bookings))
	var totalProfit domain.ProfitCents
	for _, booking := range bookings {
		requestIDs = append(requestIDs, booking.ID().String())
		totalProfit += booking.Profit()
	}

	return RevenueResult{
		RequestIDs:  requestIDs,
		TotalProfit: totalProfit,
		Metrics:     domain.CalculateNightMetrics(bookings),
	}, nil
}
