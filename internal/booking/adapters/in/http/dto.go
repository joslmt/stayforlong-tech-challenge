package bookinghttp

import "github.com/joslmt/stayforlong-tech-challenge/internal/booking/domain"

type bookingRequestDTO struct {
	RequestID   string `json:"request_id"`
	CheckIn     string `json:"check_in"`
	Nights      int32  `json:"nights"`
	SellingRate int32  `json:"selling_rate"`
	Margin      int32  `json:"margin"`
}

type dataResponse struct {
	AverageNight float64 `json:"avg_night"`
	MinimumNight float64 `json:"min_night"`
	MaximumNight float64 `json:"max_night"`
}

type revenueResponse struct {
	RequestIDs   []string `json:"request_ids"`
	TotalProfit  float64  `json:"total_profit"`
	AverageNight float64  `json:"avg_night"`
	MinimumNight float64  `json:"min_night"`
	MaximumNight float64  `json:"max_night"`
}

type errorResponse struct {
	Message string `json:"message"`
}

func euroFloat(cents domain.ProfitCents) float64 {
	return float64(cents) / 100
}
