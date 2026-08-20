package domain_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/joslmt/stayforlong-tech-challenge/internal/booking/domain"
	"github.com/joslmt/stayforlong-tech-challenge/internal/booking/testkit"
)

func TestScheduleOptimizer_GivenConflicts_WhenOptimising_ThenAppliesDomainPriorities(t *testing.T) {
	// Given: first+second occupy seven nights; the conflicting alternative occupies six.
	first := scheduleBooking(t, 1, "2026-09-01", 4, 400, 10, 0)
	second := scheduleBooking(t, 2, "2026-09-05", 3, 300, 10, 1)
	alternative := scheduleBooking(t, 3, "2026-09-03", 6, 900, 50, 2)
	batch := testkit.BookingBatchMother(t, first, second, alternative)

	// When
	schedule, err := domain.NewScheduleOptimizer().Optimize(context.Background(), batch)

	// Then
	if err != nil {
		t.Fatalf("optimise schedule: %v", err)
	}
	got := schedule.Bookings()
	if len(got) != 2 || !reflect.DeepEqual(
		[]string{got[0].ID().String(), got[1].ID().String()},
		[]string{testkit.UUIDv4(1), testkit.UUIDv4(2)},
	) {
		t.Fatalf("expected first and second bookings, got %#v", got)
	}
}

func TestNewSchedule_GivenConflictingBookings_WhenCreating_ThenRejectsInvalidSchedule(t *testing.T) {
	// Given
	first := scheduleBooking(t, 1, "2026-09-01", 4, 400, 10, 0)
	conflicting := scheduleBooking(t, 2, "2026-09-03", 2, 200, 10, 1)

	// When
	_, err := domain.NewSchedule([]domain.BookingRequest{first, conflicting})

	// Then
	assertRuleCode(t, err, domain.CodeInternal)
}

func TestScheduleOptimizer_GivenCancellationDuringCalculation_WhenOptimising_ThenStopsPromptly(t *testing.T) {
	// Given: this context deterministically cancels at the optimiser's periodic poll.
	ctx := &cancelAfterFirstCheckContext{}
	batch := testkit.BookingBatchMother(t, scheduleBooking(t, 1, "2026-09-01", 1, 100, 10, 0))

	// When
	_, err := domain.NewScheduleOptimizer().Optimize(ctx, batch)

	// Then
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
}

type cancelAfterFirstCheckContext struct{ checks int }

func (*cancelAfterFirstCheckContext) Deadline() (time.Time, bool) { return time.Time{}, false }
func (*cancelAfterFirstCheckContext) Done() <-chan struct{}       { return nil }
func (*cancelAfterFirstCheckContext) Value(any) any               { return nil }
func (ctx *cancelAfterFirstCheckContext) Err() error {
	ctx.checks++
	if ctx.checks > 1 {
		return context.Canceled
	}
	return nil
}

func scheduleBooking(t testing.TB, sequence int, checkIn string, nights, rate, margin int32, input int) domain.BookingRequest {
	t.Helper()
	values := testkit.DefaultBooking()
	values.RequestID = testkit.UUIDv4(sequence)
	values.CheckIn = checkIn
	values.Nights = nights
	values.SellingRate = rate
	values.Margin = margin
	values.InputIndex = input
	return testkit.BookingMother(t, values)
}
