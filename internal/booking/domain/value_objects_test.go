package domain_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/joslmt/stayforlong-tech-challenge/internal/booking/domain"
	"github.com/joslmt/stayforlong-tech-challenge/internal/booking/testkit"
)

func TestNewRequestID_GivenUUIDv4_WhenCreating_ThenItSucceedsAndNormalisesCase(t *testing.T) {
	// Given / When
	id, err := domain.NewRequestID("550E8400-E29B-41D4-A716-446655440000")

	// Then
	if err != nil || id.String() != "550e8400-e29b-41d4-a716-446655440000" {
		t.Fatalf("expected valid request ID, got id=%q err=%v", id.String(), err)
	}
}

func TestValueObjects_GivenInclusiveBoundaryValues_WhenCreating_ThenTheyAreAccepted(t *testing.T) {
	// Given / When / Then: constraints are inclusive at their documented boundaries.
	for _, nights := range []int32{1, domain.MaxStayNights} {
		if _, err := domain.NewStayNights(nights); err != nil {
			t.Fatalf("expected %d nights to be valid: %v", nights, err)
		}
	}
	for _, margin := range []int32{0, 100} {
		if _, err := domain.NewMarginPercent(margin); err != nil {
			t.Fatalf("expected %d%% margin to be valid: %v", margin, err)
		}
	}
	for _, date := range []string{"2020-01-01", "2099-12-31"} {
		if _, err := domain.ParseLocalDate(date); err != nil {
			t.Fatalf("expected historical/future date %s to be valid: %v", date, err)
		}
	}
}

func TestNewRequestID_GivenNonV4UUID_WhenCreating_ThenItIsRejected(t *testing.T) {
	tests := []string{
		"bookata_XY123",
		"550e8400-e29b-11d4-a716-446655440000", // UUID v1.
		"550e8400-e29b-41d4-7716-446655440000", // Invalid RFC 4122 variant.
		"550e8400e29b41d4a716446655440000",     // Missing canonical separators.
	}
	for _, value := range tests {
		_, err := domain.NewRequestID(value)
		var ruleError *domain.RuleError
		if !errors.As(err, &ruleError) || ruleError.Code != domain.CodeInvalidRequestID {
			t.Fatalf("expected %q to return invalid request ID error, got %v", value, err)
		}
	}
}

func TestValueObjects_GivenInvalidInput_WhenCreating_ThenReturnTypedRuleError(t *testing.T) {
	tests := []struct {
		name string
		act  func() error
		code domain.ErrorCode
	}{
		{"empty request ID", func() error { _, err := domain.NewRequestID(""); return err }, domain.CodeInvalidRequestID},
		{"invalid date", func() error { _, err := domain.ParseLocalDate("2026-02-30"); return err }, domain.CodeInvalidCheckIn},
		{"zero nights", func() error { _, err := domain.NewStayNights(0); return err }, domain.CodeInvalidNights},
		{"too many nights", func() error { _, err := domain.NewStayNights(366); return err }, domain.CodeInvalidNights},
		{"zero selling rate", func() error { _, err := domain.NewSellingRate(0); return err }, domain.CodeInvalidRate},
		{"negative margin", func() error { _, err := domain.NewMarginPercent(-1); return err }, domain.CodeInvalidMargin},
		{"margin over 100", func() error { _, err := domain.NewMarginPercent(101); return err }, domain.CodeInvalidMargin},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given / When
			err := test.act()
			// Then
			var ruleError *domain.RuleError
			if !errors.As(err, &ruleError) || ruleError.Code != test.code {
				t.Fatalf("expected rule error code %d, got %v", test.code, err)
			}
		})
	}
}

func TestNewMarginPercent_GivenZero_WhenCreating_ThenItIsAllowed(t *testing.T) {
	margin, err := domain.NewMarginPercent(0)
	if err != nil || margin.Int32() != 0 {
		t.Fatalf("expected a valid zero margin, got %d err=%v", margin.Int32(), err)
	}
}

func TestRoundRatioCentsHalfUp_GivenHalfCent_WhenRounding_ThenRoundsAwayFromZero(t *testing.T) {
	if got := domain.RoundRatioCentsHalfUp(1, 2); got != 1 {
		t.Fatalf("expected 0.5 cents to round to 1 cent, got %d", got)
	}
}

func TestRoundRatioCentsHalfUp_GivenInvalidDenominator_WhenRounding_ThenReturnsZero(t *testing.T) {
	if got := domain.RoundRatioCentsHalfUp(100, 0); got != 0 {
		t.Fatalf("expected invalid denominator to return zero, got %d", got)
	}
}

func TestBookingRequest_GivenValidValues_WhenRead_ThenExposesImmutableValueObjects(t *testing.T) {
	booking := testkit.BookingMother(t, testkit.DefaultBooking())
	if booking.SellingRate().Euros() != 200 || booking.Margin().Int32() != 20 || booking.CheckIn().String() != "2026-09-01" {
		t.Fatalf("unexpected booking values: rate=%d margin=%d check_in=%s",
			booking.SellingRate().Euros(), booking.Margin().Int32(), booking.CheckIn().String())
	}
}

func TestRuleError_GivenInvalidValue_WhenRendered_ThenIncludesStableNumericCode(t *testing.T) {
	_, err := domain.NewSellingRate(0)
	if err == nil || !strings.Contains(err.Error(), "domain error 1401") {
		t.Fatalf("expected stable code in internal error, got %v", err)
	}
}
