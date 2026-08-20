package bookinghttp_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	bookinghttp "github.com/joslmt/stayforlong-tech-challenge/internal/booking/adapters/in/http"
	"github.com/joslmt/stayforlong-tech-challenge/internal/booking/application"
	"github.com/joslmt/stayforlong-tech-challenge/internal/booking/domain"
)

var errUnexpectedService = errors.New("database credentials must not leak")

type failingMetrics struct{}

func (failingMetrics) Calculate(context.Context, domain.BookingBatch) (application.MetricsResult, error) {
	return application.MetricsResult{}, errUnexpectedService
}

type failingSchedule struct{}

func (failingSchedule) Build(context.Context, domain.BookingBatch) (application.RevenueResult, error) {
	return application.RevenueResult{}, errUnexpectedService
}

func TestHandler_GivenApplicationFailure_WhenCallingEitherEndpoint_ThenMapsSafe500(t *testing.T) {
	// Given
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := bookinghttp.NewHandler(failingMetrics{}, failingSchedule{}, logger)
	payload := `[{"request_id":"00000000-0000-4000-8000-000000000001","check_in":"2026-09-01","nights":1,"selling_rate":100,"margin":20}]`

	tests := []struct {
		name string
		act  func(http.ResponseWriter, *http.Request)
	}{
		{name: "data", act: handler.Data},
		{name: "revenue", act: handler.Revenue},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/"+test.name, bytes.NewBufferString(payload))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()

			// When
			test.act(response, request)

			// Then
			if response.Code != http.StatusInternalServerError || response.Body.String() != "{\"message\":\"Internal server error\"}\n" {
				t.Fatalf("expected safe message-only 500, got %d %s", response.Code, response.Body.String())
			}
		})
	}
}
