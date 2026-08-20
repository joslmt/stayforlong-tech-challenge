package domain

// ProfitPerNight preserves a nightly profit as an exact rational value.
type ProfitPerNight struct {
	profit ProfitCents
	nights int64
}

func newProfitPerNight(profit ProfitCents, nights int64) ProfitPerNight {
	return ProfitPerNight{profit: profit, nights: nights}
}

func (value ProfitPerNight) Numerator() ProfitCents { return value.profit }
func (value ProfitPerNight) Denominator() int64     { return value.nights }

// RoundedCents applies half-up rounding when an outbound adapter requests it.
func (value ProfitPerNight) RoundedCents() ProfitCents {
	return RoundRatioCentsHalfUp(value.profit, value.nights)
}

func (value ProfitPerNight) lessThan(other ProfitPerNight) bool {
	return int64(value.profit)*other.nights < int64(other.profit)*value.nights
}

// NightMetrics groups exact average, minimum, and maximum nightly profits.
type NightMetrics struct {
	average ProfitPerNight
	minimum ProfitPerNight
	maximum ProfitPerNight
}

func (metrics NightMetrics) Average() ProfitPerNight { return metrics.average }
func (metrics NightMetrics) Minimum() ProfitPerNight { return metrics.minimum }
func (metrics NightMetrics) Maximum() ProfitPerNight { return metrics.maximum }

func CalculateNightMetrics(bookings []BookingRequest) NightMetrics {
	if len(bookings) == 0 {
		return NightMetrics{}
	}

	var totalProfit ProfitCents
	var totalNights int64
	minimum := profitPerNightFor(bookings[0])
	maximum := minimum

	for _, booking := range bookings {
		nightlyProfit := profitPerNightFor(booking)
		totalProfit += booking.Profit()
		totalNights += int64(booking.Nights().Int32())

		if nightlyProfit.lessThan(minimum) {
			minimum = nightlyProfit
		}
		if maximum.lessThan(nightlyProfit) {
			maximum = nightlyProfit
		}
	}

	return NightMetrics{
		average: newProfitPerNight(totalProfit, totalNights),
		minimum: minimum,
		maximum: maximum,
	}
}

func profitPerNightFor(booking BookingRequest) ProfitPerNight {
	return newProfitPerNight(booking.Profit(), int64(booking.Nights().Int32()))
}
