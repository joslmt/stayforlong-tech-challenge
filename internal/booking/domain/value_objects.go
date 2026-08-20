package domain

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

const (
	RequestIDLength = 36
	MaxStayNights   = 365
)

// UUID v4 requires the version nibble to be 4 and the RFC 4122 variant to be 8, 9, a, or b.
var requestIDPattern = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

type RequestID struct{ value string }

func NewRequestID(value string) (RequestID, error) {
	if len(value) != RequestIDLength || !requestIDPattern.MatchString(value) {
		return RequestID{}, newRuleError(CodeInvalidRequestID, "request_id must be a valid UUID v4")
	}
	return RequestID{value: strings.ToLower(value)}, nil
}

func (id RequestID) String() string { return id.value }

// LocalDate models a business calendar date without inventing a timezone not present in the API.
type LocalDate struct{ value time.Time }

func ParseLocalDate(value string) (LocalDate, error) {
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil || parsed.Format("2006-01-02") != value {
		return LocalDate{}, newRuleError(CodeInvalidCheckIn, "check_in must be a valid date in YYYY-MM-DD format")
	}
	return LocalDate{value: parsed}, nil
}

func (d LocalDate) String() string              { return d.value.Format("2006-01-02") }
func (d LocalDate) AddDays(days int) LocalDate  { return LocalDate{value: d.value.AddDate(0, 0, days)} }
func (d LocalDate) Before(other LocalDate) bool { return d.value.Before(other.value) }
func (d LocalDate) Equal(other LocalDate) bool  { return d.value.Equal(other.value) }

type StayNights struct{ value int32 }

func NewStayNights(value int32) (StayNights, error) {
	if value < 1 || value > MaxStayNights {
		return StayNights{}, newRuleError(CodeInvalidNights, fmt.Sprintf("nights must be between 1 and %d", MaxStayNights))
	}
	return StayNights{value: value}, nil
}

func (n StayNights) Int32() int32 { return n.value }

// SellingRate is the retail price for the complete stay, expressed in whole euros by the API.
type SellingRate struct{ euros int32 }

func NewSellingRate(euros int32) (SellingRate, error) {
	if euros <= 0 {
		return SellingRate{}, newRuleError(CodeInvalidRate, "selling_rate must be greater than zero euros")
	}
	return SellingRate{euros: euros}, nil
}

func (r SellingRate) Euros() int32 { return r.euros }

type MarginPercent struct{ value int32 }

func NewMarginPercent(value int32) (MarginPercent, error) {
	if value < 0 || value > 100 {
		return MarginPercent{}, newRuleError(CodeInvalidMargin, "margin must be between 0 and 100 percent")
	}
	return MarginPercent{value: value}, nil
}

func (m MarginPercent) Int32() int32 { return m.value }

// ProfitCents is exact for integer euro selling rates and integer percentage margins:
// euros * 100 cents * margin / 100 = euros * margin cents.
type ProfitCents int64

func CalculateProfit(rate SellingRate, margin MarginPercent) ProfitCents {
	return ProfitCents(int64(rate.euros) * int64(margin.value))
}

// RoundRatioCentsHalfUp rounds a positive rational number of cents at the API boundary.
func RoundRatioCentsHalfUp(numerator ProfitCents, denominator int64) ProfitCents {
	if denominator <= 0 {
		return 0
	}
	return ProfitCents((int64(numerator)*2 + denominator) / (2 * denominator))
}
