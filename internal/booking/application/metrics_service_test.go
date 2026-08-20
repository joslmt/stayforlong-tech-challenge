package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/joslmt/stayforlong-tech-challenge/internal/booking/application"
	"github.com/joslmt/stayforlong-tech-challenge/internal/booking/domain"
	"github.com/joslmt/stayforlong-tech-challenge/internal/booking/testkit"
)

func TestCalculate_GivenBookings_WhenCalculatingMetrics_ThenReturnsWeightedNightValues(t *testing.T) {
	// Given: €20 across three nights and €5 across two nights.
	first := testkit.DefaultBooking()
	first.RequestID, first.Nights, first.SellingRate, first.Margin = testkit.UUIDv4(1), 3, 100, 20
	second := testkit.DefaultBooking()
	second.RequestID, second.Nights, second.SellingRate, second.Margin = testkit.UUIDv4(2), 2, 50, 10
	bookings := []domain.BookingRequest{
		testkit.BookingMother(t, first),
		testkit.BookingMother(t, second),
	}

	// When
	result, err := application.NewMetricsService().Calculate(context.Background(), testkit.BookingBatchMother(t, bookings...))

	// Then: (€20 + €5) / five nights = €5 average per night.
	if err != nil {
		t.Fatalf("calculate metrics: %v", err)
	}
	assertApplicationRatio(t, "average", result.Metrics.Average(), 2500, 5)
	assertApplicationRatio(t, "minimum", result.Metrics.Minimum(), 500, 2)
	assertApplicationRatio(t, "maximum", result.Metrics.Maximum(), 2000, 3)
}

func TestCalculate_GivenCancelledContext_WhenCalculatingMetrics_ThenStops(t *testing.T) {
	// Given
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// When
	_, err := application.NewMetricsService().Calculate(ctx, domain.BookingBatch{})

	// Then
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
}

func TestCalculate_GivenInvalidBatch_WhenCalculatingMetrics_ThenProtectsApplicationBoundary(t *testing.T) {
	// Given: the zero value demonstrates a non-HTTP caller cannot bypass batch rules.
	var batch domain.BookingBatch

	// When
	_, err := application.NewMetricsService().Calculate(context.Background(), batch)

	// Then
	var ruleError *domain.RuleError
	if !errors.As(err, &ruleError) || ruleError.Code != domain.CodeEmptyBookingList {
		t.Fatalf("expected empty-batch rule error, got %v", err)
	}
}

func assertApplicationRatio(t *testing.T, name string, got domain.ProfitPerNight, numerator domain.ProfitCents, denominator int64) {
	t.Helper()
	if got.Numerator() != numerator || got.Denominator() != denominator {
		t.Fatalf("expected %s ratio %d/%d, got %d/%d", name, numerator, denominator, got.Numerator(), got.Denominator())
	}
}
