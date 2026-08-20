package application

import (
	"context"

	"github.com/joslmt/stayforlong-tech-challenge/internal/booking/domain"
)

// MetricsCalculator is the inbound application port used by transport adapters.
type MetricsCalculator interface {
	Calculate(context.Context, domain.BookingBatch) (MetricsResult, error)
}

// ScheduleBuilder is the inbound application port used by transport adapters.
type ScheduleBuilder interface {
	Build(context.Context, domain.BookingBatch) (RevenueResult, error)
}
