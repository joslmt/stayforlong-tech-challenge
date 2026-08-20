package testkit

import (
	"fmt"
	"testing"

	"github.com/joslmt/stayforlong-tech-challenge/internal/booking/domain"
)

type BookingValues struct {
	RequestID   string
	CheckIn     string
	Nights      int32
	SellingRate int32
	Margin      int32
	InputIndex  int
}

func DefaultBooking() BookingValues {
	return BookingValues{
		RequestID: UUIDv4(1), CheckIn: "2026-09-01", Nights: 2,
		SellingRate: 200, Margin: 20, InputIndex: 0,
	}
}

// UUIDv4 returns deterministic UUID v4 fixtures without hiding random data in tests.
func UUIDv4(sequence int) string {
	return fmt.Sprintf("00000000-0000-4000-8000-%012x", sequence)
}

func BookingMother(t testing.TB, values BookingValues) domain.BookingRequest {
	t.Helper()
	id, err := domain.NewRequestID(values.RequestID)
	if err != nil {
		t.Fatalf("booking mother request ID: %v", err)
	}
	checkIn, err := domain.ParseLocalDate(values.CheckIn)
	if err != nil {
		t.Fatalf("booking mother check-in: %v", err)
	}
	nights, err := domain.NewStayNights(values.Nights)
	if err != nil {
		t.Fatalf("booking mother nights: %v", err)
	}
	rate, err := domain.NewSellingRate(values.SellingRate)
	if err != nil {
		t.Fatalf("booking mother rate: %v", err)
	}
	margin, err := domain.NewMarginPercent(values.Margin)
	if err != nil {
		t.Fatalf("booking mother margin: %v", err)
	}
	return domain.NewBookingRequest(id, checkIn, nights, rate, margin, values.InputIndex)
}

func BookingBatchMother(t testing.TB, bookings ...domain.BookingRequest) domain.BookingBatch {
	t.Helper()
	batch, err := domain.NewBookingBatch(bookings)
	if err != nil {
		t.Fatalf("booking batch mother: %v", err)
	}
	return batch
}
