package application_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/joslmt/stayforlong-tech-challenge/internal/booking/application"
	"github.com/joslmt/stayforlong-tech-challenge/internal/booking/domain"
	"github.com/joslmt/stayforlong-tech-challenge/internal/booking/testkit"
)

func TestBuild_GivenConflictingBookings_WhenOptimising_ThenMaximisesOccupiedNights(t *testing.T) {
	// Given: A+B occupy 7 compatible nights, while C conflicts and occupies 6.
	bookings := []domain.BookingRequest{
		mother(t, 1, "2026-09-01", 4, 400, 10, 0),
		mother(t, 2, "2026-09-05", 3, 300, 10, 1),
		mother(t, 3, "2026-09-03", 6, 900, 50, 2),
	}

	// When
	result, err := newScheduleService().Build(context.Background(), testkit.BookingBatchMother(t, bookings...))

	// Then
	if err != nil {
		t.Fatalf("build schedule: %v", err)
	}
	if !reflect.DeepEqual(result.RequestIDs, []string{testkit.UUIDv4(1), testkit.UUIDv4(2)}) {
		t.Fatalf("expected first and second UUIDs, got %v", result.RequestIDs)
	}
}

func TestBuild_GivenEqualOccupancy_WhenOptimising_ThenUsesProfitTieBreaker(t *testing.T) {
	// Given
	bookings := []domain.BookingRequest{
		mother(t, 1, "2026-09-01", 4, 400, 10, 0),
		mother(t, 2, "2026-09-01", 4, 400, 20, 1),
	}

	// When
	result, err := newScheduleService().Build(context.Background(), testkit.BookingBatchMother(t, bookings...))

	// Then
	if err != nil || !reflect.DeepEqual(result.RequestIDs, []string{testkit.UUIDv4(2)}) {
		t.Fatalf("expected higher-profit booking, got result=%v err=%v", result.RequestIDs, err)
	}
}

func TestBuild_GivenEqualOccupancyAndProfit_WhenOptimising_ThenUsesInputOrderTieBreaker(t *testing.T) {
	// Given
	bookings := []domain.BookingRequest{
		mother(t, 1, "2026-09-01", 4, 400, 20, 0),
		mother(t, 2, "2026-09-01", 4, 400, 20, 1),
	}

	// When
	result, err := newScheduleService().Build(context.Background(), testkit.BookingBatchMother(t, bookings...))

	// Then
	if err != nil || !reflect.DeepEqual(result.RequestIDs, []string{testkit.UUIDv4(1)}) {
		t.Fatalf("expected earliest input booking, got result=%v err=%v", result.RequestIDs, err)
	}
}

func TestBuild_GivenReverseInputDates_WhenReturning_ThenIDsAreChronological(t *testing.T) {
	// Given
	bookings := []domain.BookingRequest{
		mother(t, 1, "2026-09-05", 2, 200, 20, 0),
		mother(t, 2, "2026-09-01", 2, 200, 20, 1),
	}

	// When
	result, err := newScheduleService().Build(context.Background(), testkit.BookingBatchMother(t, bookings...))

	// Then
	if err != nil || !reflect.DeepEqual(result.RequestIDs, []string{testkit.UUIDv4(2), testkit.UUIDv4(1)}) {
		t.Fatalf("expected chronological IDs, got result=%v err=%v", result.RequestIDs, err)
	}
}

func TestBuild_GivenCancelledContext_WhenStarting_ThenStops(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := newScheduleService().Build(ctx, testkit.BookingBatchMother(t, mother(t, 1, "2026-09-01", 2, 200, 20, 0)))
	if err == nil {
		t.Fatal("expected context cancellation")
	}
}

func TestBuild_GivenTenThousandNonConflictingBookings_WhenOptimising_ThenSelectsAllEfficiently(t *testing.T) {
	// Given
	bookings := make([]domain.BookingRequest, domain.MaxBookingsPerRequest)
	start, _ := time.Parse("2006-01-02", "2026-01-01")
	for index := range bookings {
		bookings[index] = mother(t, index+1, start.AddDate(0, 0, index).Format("2006-01-02"), 1, 100, 10, index)
	}

	// When
	result, err := newScheduleService().Build(context.Background(), testkit.BookingBatchMother(t, bookings...))

	// Then
	if err != nil || len(result.RequestIDs) != domain.MaxBookingsPerRequest {
		t.Fatalf("expected all %d bookings, got %d err=%v", domain.MaxBookingsPerRequest, len(result.RequestIDs), err)
	}
}

func mother(t testing.TB, idSequence int, checkIn string, nights, rate, margin int32, input int) domain.BookingRequest {
	t.Helper()
	values := testkit.DefaultBooking()
	values.RequestID, values.CheckIn, values.Nights = testkit.UUIDv4(idSequence), checkIn, nights
	values.SellingRate, values.Margin, values.InputIndex = rate, margin, input
	return testkit.BookingMother(t, values)
}

func newScheduleService() application.ScheduleService {
	return application.NewScheduleService(domain.NewScheduleOptimizer())
}
