package domain

import "fmt"

// ErrorCode is an internal, stable numeric classification. The HTTP API intentionally
// exposes only a human-friendly message to remain compatible with the challenge contract.
type ErrorCode int

const (
	CodeEmptyBookingList ErrorCode = 1001
	CodeTooManyBookings  ErrorCode = 1002
	CodeInvalidRequestID ErrorCode = 1101
	CodeDuplicateID      ErrorCode = 1102
	CodeInvalidCheckIn   ErrorCode = 1201
	CodeInvalidNights    ErrorCode = 1301
	CodeInvalidRate      ErrorCode = 1401
	CodeInvalidMargin    ErrorCode = 1501
	CodeInternal         ErrorCode = 9000
)

type RuleError struct {
	Code    ErrorCode
	Message string
}

func (e *RuleError) Error() string {
	return fmt.Sprintf("domain error %d: %s", e.Code, e.Message)
}

func newRuleError(code ErrorCode, message string) error {
	return &RuleError{Code: code, Message: message}
}
