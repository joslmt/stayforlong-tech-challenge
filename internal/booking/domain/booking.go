package domain

type BookingRequest struct {
	id          RequestID
	checkIn     LocalDate
	nights      StayNights
	sellingRate SellingRate
	margin      MarginPercent
	inputIndex  int
}

func NewBookingRequest(
	id RequestID,
	checkIn LocalDate,
	nights StayNights,
	sellingRate SellingRate,
	margin MarginPercent,
	inputIndex int,
) BookingRequest {
	return BookingRequest{
		id: id, checkIn: checkIn, nights: nights,
		sellingRate: sellingRate, margin: margin, inputIndex: inputIndex,
	}
}

func (b BookingRequest) ID() RequestID            { return b.id }
func (b BookingRequest) CheckIn() LocalDate       { return b.checkIn }
func (b BookingRequest) Nights() StayNights       { return b.nights }
func (b BookingRequest) SellingRate() SellingRate { return b.sellingRate }
func (b BookingRequest) Margin() MarginPercent    { return b.margin }
func (b BookingRequest) InputIndex() int          { return b.inputIndex }
func (b BookingRequest) Profit() ProfitCents      { return CalculateProfit(b.sellingRate, b.margin) }
func (b BookingRequest) CheckOut() LocalDate      { return b.checkIn.AddDays(int(b.nights.value)) }

func (b BookingRequest) validate() error {
	if b.id.value == "" {
		return newRuleError(CodeInvalidRequestID, "request_id must be a valid UUID v4")
	}
	if b.checkIn.value.IsZero() {
		return newRuleError(CodeInvalidCheckIn, "check_in must be a valid date in YYYY-MM-DD format")
	}
	if b.nights.value < 1 || b.nights.value > MaxStayNights {
		return newRuleError(CodeInvalidNights, "nights must be between 1 and 365")
	}
	if b.sellingRate.euros <= 0 {
		return newRuleError(CodeInvalidRate, "selling_rate must be greater than zero euros")
	}
	if b.margin.value < 0 || b.margin.value > 100 {
		return newRuleError(CodeInvalidMargin, "margin must be between 0 and 100 percent")
	}
	return nil
}

// CanFollow encodes the fixed hotel operating times: checkout is 11:00,
// cleaning finishes at 12:00, and the next check-in is 15:00. Therefore a
// booking can start on the previous booking's checkout calendar date.
func (b BookingRequest) CanFollow(previous BookingRequest) bool {
	return !b.checkIn.Before(previous.CheckOut())
}

type Schedule struct{ bookings []BookingRequest }

func NewSchedule(bookings []BookingRequest) (Schedule, error) {
	for index := 1; index < len(bookings); index++ {
		if !bookings[index].CanFollow(bookings[index-1]) {
			return Schedule{}, newRuleError(CodeInternal, "selected schedule contains conflicting bookings")
		}
	}
	return Schedule{bookings: append([]BookingRequest(nil), bookings...)}, nil
}

func (s Schedule) Bookings() []BookingRequest {
	return append([]BookingRequest(nil), s.bookings...)
}
