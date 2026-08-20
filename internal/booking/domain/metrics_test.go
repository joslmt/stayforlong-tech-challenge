package domain_test

import (
	"testing"

	"github.com/joslmt/stayforlong-tech-challenge/internal/booking/domain"
	"github.com/joslmt/stayforlong-tech-challenge/internal/booking/testkit"
)

func TestCalculateNightMetrics_GivenBookings_WhenCalculating_ThenUsesWeightedAverageAndPerBookingExtremes(t *testing.T) {
	// Given
	first := testkit.DefaultBooking()
	first.RequestID, first.Nights, first.SellingRate, first.Margin = testkit.UUIDv4(1), 3, 100, 20 // €20 / 3 = €6.67
	second := testkit.DefaultBooking()
	second.RequestID, second.Nights, second.SellingRate, second.Margin = testkit.UUIDv4(2), 2, 50, 10 // €5 / 2 = €2.50
	bookings := []domain.BookingRequest{testkit.BookingMother(t, first), testkit.BookingMother(t, second)}

	// When
	metrics := domain.CalculateNightMetrics(bookings)

	// Then: exact ratios are retained; €25 / 5 nights is not converted here.
	assertRatio(t, "average", metrics.Average(), 2500, 5)
	assertRatio(t, "minimum", metrics.Minimum(), 500, 2)
	assertRatio(t, "maximum", metrics.Maximum(), 2000, 3)
}

func TestCalculateNightMetrics_GivenHalfCentPerNight_WhenCalculating_ThenKeepsExactRatioUntilConversion(t *testing.T) {
	// Given: €1 at 1% is one cent profit across two nights.
	values := testkit.DefaultBooking()
	values.Nights, values.SellingRate, values.Margin = 2, 1, 1

	// When
	metrics := domain.CalculateNightMetrics([]domain.BookingRequest{testkit.BookingMother(t, values)})

	// Then
	assertRatio(t, "average", metrics.Average(), 1, 2)
	if got := metrics.Average().RoundedCents(); got != 1 {
		t.Fatalf("expected boundary rounding to produce one cent, got %d", got)
	}
}

func TestCalculateNightMetrics_GivenNoBookings_WhenCalculating_ThenReturnsZeroMetrics(t *testing.T) {
	metrics := domain.CalculateNightMetrics(nil)
	if metrics.Average().Numerator() != 0 || metrics.Average().Denominator() != 0 {
		t.Fatalf("expected zero metrics, got %d/%d", metrics.Average().Numerator(), metrics.Average().Denominator())
	}
}

func TestBookingCanFollow_GivenCleaningFitsBeforeCheckIn_WhenChecking_ThenCheckoutDateIsCompatible(t *testing.T) {
	// Given: the first stay checks out on 5 September at 11:00; cleaning ends at 12:00.
	firstValues := testkit.DefaultBooking()
	firstValues.CheckIn, firstValues.Nights = "2026-09-01", 4
	secondValues := testkit.DefaultBooking()
	secondValues.RequestID, secondValues.CheckIn = testkit.UUIDv4(2), "2026-09-05"

	// When / Then: the next check-in is 15:00 that same calendar date.
	if !testkit.BookingMother(t, secondValues).CanFollow(testkit.BookingMother(t, firstValues)) {
		t.Fatal("expected checkout-date booking to fit after the cleaning window")
	}
}

func assertRatio(t *testing.T, name string, got domain.ProfitPerNight, numerator domain.ProfitCents, denominator int64) {
	t.Helper()
	if got.Numerator() != numerator || got.Denominator() != denominator {
		t.Fatalf("expected %s ratio %d/%d, got %d/%d", name, numerator, denominator, got.Numerator(), got.Denominator())
	}
}
