package bootstrap

import (
	"log/slog"
	"net/http"

	bookinghttp "github.com/joslmt/stayforlong-tech-challenge/internal/booking/adapters/in/http"
	"github.com/joslmt/stayforlong-tech-challenge/internal/booking/application"
	"github.com/joslmt/stayforlong-tech-challenge/internal/booking/domain"
	"github.com/joslmt/stayforlong-tech-challenge/internal/platform/httpserver"
	webassets "github.com/joslmt/stayforlong-tech-challenge/web"
)

// NewHTTPHandler is the composition root for the process HTTP interface.
func NewHTTPHandler(logger *slog.Logger) http.Handler {
	metricsService := application.NewMetricsService()
	scheduleService := application.NewScheduleService(domain.NewScheduleOptimizer())
	bookingHandler := bookinghttp.NewHandler(metricsService, scheduleService, logger)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(response http.ResponseWriter, _ *http.Request) {
		response.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("POST /data", bookingHandler.Data)
	mux.HandleFunc("POST /revenue", bookingHandler.Revenue)
	mux.Handle("GET /", webassets.Handler())
	return httpserver.Middleware(logger, mux)
}
