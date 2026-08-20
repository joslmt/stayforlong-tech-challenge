package domain_test

import (
	"errors"
	"testing"

	"github.com/joslmt/stayforlong-tech-challenge/internal/booking/domain"
	"github.com/joslmt/stayforlong-tech-challenge/internal/booking/testkit"
)

func TestNewBookingBatch_GivenEmptyBatch_WhenCreating_ThenRejectsIt(t *testing.T) {
	// Given / When
	_, err := domain.NewBookingBatch(nil)

	// Then
	assertRuleCode(t, err, domain.CodeEmptyBookingList)
}

func TestNewBookingBatch_GivenDuplicateIDs_WhenCreating_ThenRejectsThem(t *testing.T) {
	// Given
	first := testkit.BookingMother(t, testkit.DefaultBooking())
	secondValues := testkit.DefaultBooking()
	secondValues.CheckIn = "2026-09-03"
	second := testkit.BookingMother(t, secondValues)

	// When
	_, err := domain.NewBookingBatch([]domain.BookingRequest{first, second})

	// Then
	assertRuleCode(t, err, domain.CodeDuplicateID)
}

func TestNewBookingBatch_GivenMoreThanMaximum_WhenCreating_ThenRejectsBeforeProcessingItems(t *testing.T) {
	// Given: zero-value entities prove cardinality is the first invariant checked.
	bookings := make([]domain.BookingRequest, domain.MaxBookingsPerRequest+1)

	// When
	_, err := domain.NewBookingBatch(bookings)

	// Then
	assertRuleCode(t, err, domain.CodeTooManyBookings)
}

func TestBookingBatch_GivenZeroValue_WhenValidated_ThenCannotBypassInvariants(t *testing.T) {
	var batch domain.BookingBatch
	assertRuleCode(t, batch.Validate(), domain.CodeEmptyBookingList)
}

func TestNewBookingBatch_GivenZeroValueBooking_WhenCreating_ThenRejectsInvalidEntity(t *testing.T) {
	// Given / When: callers can construct a zero-value entity even though its fields are private.
	_, err := domain.NewBookingBatch([]domain.BookingRequest{{}})

	// Then
	assertRuleCode(t, err, domain.CodeInvalidRequestID)
}

func assertRuleCode(t *testing.T, err error, expected domain.ErrorCode) {
	t.Helper()
	var ruleError *domain.RuleError
	if !errors.As(err, &ruleError) || ruleError.Code != expected {
		t.Fatalf("expected rule error %d, got %v", expected, err)
	}
}
