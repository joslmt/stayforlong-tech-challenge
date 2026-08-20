package domain

import (
	"errors"
	"fmt"
)

const MaxBookingsPerRequest = 10_000

// BookingBatch is the aggregate boundary for one metrics or scheduling request.
// It protects invariants independently of the transport used to create it.
type BookingBatch struct {
	bookings []BookingRequest
}

func NewBookingBatch(bookings []BookingRequest) (BookingBatch, error) {
	batch := BookingBatch{bookings: append([]BookingRequest(nil), bookings...)}
	if err := batch.Validate(); err != nil {
		return BookingBatch{}, err
	}
	return batch, nil
}

func (batch BookingBatch) Validate() error {
	if len(batch.bookings) == 0 {
		return newRuleError(CodeEmptyBookingList, "at least one booking request is required")
	}
	if len(batch.bookings) > MaxBookingsPerRequest {
		return newRuleError(CodeTooManyBookings, fmt.Sprintf("a maximum of %d booking requests is allowed", MaxBookingsPerRequest))
	}

	seen := make(map[string]struct{}, len(batch.bookings))
	for index, booking := range batch.bookings {
		if err := booking.validate(); err != nil {
			return fieldRuleError(index, err)
		}
		id := booking.ID().String()
		if _, exists := seen[id]; exists {
			return newRuleError(CodeDuplicateID, fmt.Sprintf("request_id %q must be unique", id))
		}
		seen[id] = struct{}{}
	}
	return nil
}

func (batch BookingBatch) Bookings() []BookingRequest {
	return append([]BookingRequest(nil), batch.bookings...)
}

func fieldRuleError(index int, err error) error {
	var ruleError *RuleError
	if !errors.As(err, &ruleError) {
		return err
	}
	return newRuleError(ruleError.Code, fmt.Sprintf("booking at index %d: %s", index, ruleError.Message))
}
