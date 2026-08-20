package bookinghttp

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"

	"github.com/joslmt/stayforlong-tech-challenge/internal/booking/application"
	"github.com/joslmt/stayforlong-tech-challenge/internal/booking/domain"
)

const MaxBodyBytes int64 = 5 << 20

type Handler struct {
	metrics   application.MetricsCalculator
	scheduler application.ScheduleBuilder
	logger    *slog.Logger
}

func NewHandler(metrics application.MetricsCalculator, scheduler application.ScheduleBuilder, logger *slog.Logger) Handler {
	return Handler{metrics: metrics, scheduler: scheduler, logger: logger}
}

func (h Handler) Data(response http.ResponseWriter, request *http.Request) {
	bookings, err := h.decodeBookings(response, request)
	if err != nil {
		h.writeMappedError(response, request, err)
		return
	}
	result, err := h.metrics.Calculate(request.Context(), bookings)
	if err != nil {
		h.writeMappedError(response, request, err)
		return
	}
	h.writeJSON(response, http.StatusOK, dataResponse{
		AverageNight: euroFloat(result.Metrics.Average().RoundedCents()),
		MinimumNight: euroFloat(result.Metrics.Minimum().RoundedCents()),
		MaximumNight: euroFloat(result.Metrics.Maximum().RoundedCents()),
	})
}

func (h Handler) Revenue(response http.ResponseWriter, request *http.Request) {
	bookings, err := h.decodeBookings(response, request)
	if err != nil {
		h.writeMappedError(response, request, err)
		return
	}
	result, err := h.scheduler.Build(request.Context(), bookings)
	if err != nil {
		h.writeMappedError(response, request, err)
		return
	}
	h.writeJSON(response, http.StatusOK, revenueResponse{
		RequestIDs:   result.RequestIDs,
		TotalProfit:  euroFloat(result.TotalProfit),
		AverageNight: euroFloat(result.Metrics.Average().RoundedCents()),
		MinimumNight: euroFloat(result.Metrics.Minimum().RoundedCents()),
		MaximumNight: euroFloat(result.Metrics.Maximum().RoundedCents()),
	})
}

func (h Handler) decodeBookings(response http.ResponseWriter, request *http.Request) (domain.BookingBatch, error) {
	contentType, _, err := mime.ParseMediaType(request.Header.Get("Content-Type"))
	if err != nil || contentType != "application/json" {
		return domain.BookingBatch{}, newRequestError(1003, "Content-Type must be application/json")
	}

	request.Body = http.MaxBytesReader(response, request.Body, MaxBodyBytes)
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	var payload []bookingRequestDTO
	if err := decoder.Decode(&payload); err != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			return domain.BookingBatch{}, newRequestError(1002, "request body must not exceed 5 MiB")
		}
		return domain.BookingBatch{}, newRequestError(1003, "request body must be a valid JSON array: "+humanizeJSONError(err))
	}
	if err := ensureSingleJSONValue(decoder); err != nil {
		return domain.BookingBatch{}, newRequestError(1003, "request body must contain exactly one JSON array")
	}

	bookings := make([]domain.BookingRequest, 0, len(payload))
	for index, item := range payload {
		id, err := domain.NewRequestID(item.RequestID)
		if err != nil {
			return domain.BookingBatch{}, fieldError(index, err)
		}
		checkIn, err := domain.ParseLocalDate(item.CheckIn)
		if err != nil {
			return domain.BookingBatch{}, fieldError(index, err)
		}
		nights, err := domain.NewStayNights(item.Nights)
		if err != nil {
			return domain.BookingBatch{}, fieldError(index, err)
		}
		rate, err := domain.NewSellingRate(item.SellingRate)
		if err != nil {
			return domain.BookingBatch{}, fieldError(index, err)
		}
		margin, err := domain.NewMarginPercent(item.Margin)
		if err != nil {
			return domain.BookingBatch{}, fieldError(index, err)
		}
		bookings = append(bookings, domain.NewBookingRequest(id, checkIn, nights, rate, margin, index))
	}
	return domain.NewBookingBatch(bookings)
}

func fieldError(index int, err error) error {
	var ruleError *domain.RuleError
	if errors.As(err, &ruleError) {
		return &domain.RuleError{Code: ruleError.Code, Message: fmt.Sprintf("booking at index %d: %s", index, ruleError.Message)}
	}
	return err
}

func ensureSingleJSONValue(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return errors.New("unexpected trailing JSON value")
	}
	return nil
}

func humanizeJSONError(err error) string {
	var syntaxError *json.SyntaxError
	var typeError *json.UnmarshalTypeError
	switch {
	case errors.As(err, &syntaxError):
		return fmt.Sprintf("invalid JSON near byte %d", syntaxError.Offset)
	case errors.As(err, &typeError):
		if typeError.Field != "" {
			return fmt.Sprintf("field %s has an invalid type", typeError.Field)
		}
		return "a field has an invalid type"
	case errors.Is(err, io.EOF):
		return "the body is empty"
	default:
		return err.Error()
	}
}

func (h Handler) writeMappedError(response http.ResponseWriter, request *http.Request, err error) {
	var transportError *requestError
	if errors.As(err, &transportError) {
		h.logger.WarnContext(request.Context(), "request rejected", "error_code", transportError.code, "error", transportError.Error())
		h.writeJSON(response, http.StatusBadRequest, errorResponse{Message: transportError.message})
		return
	}
	var ruleError *domain.RuleError
	if errors.As(err, &ruleError) && ruleError.Code != domain.CodeInternal {
		h.logger.WarnContext(request.Context(), "request rejected", "error_code", int(ruleError.Code), "error", ruleError.Error())
		h.writeJSON(response, http.StatusBadRequest, errorResponse{Message: ruleError.Message})
		return
	}
	h.logger.ErrorContext(request.Context(), "request failed", "error_code", int(domain.CodeInternal), "error", err)
	h.writeJSON(response, http.StatusInternalServerError, errorResponse{Message: "Internal server error"})
}

type requestError struct {
	code    int
	message string
}

func newRequestError(code int, message string) error {
	return &requestError{code: code, message: message}
}

func (err *requestError) Error() string {
	return fmt.Sprintf("request error %d: %s", err.code, err.message)
}

func (Handler) writeJSON(response http.ResponseWriter, status int, payload any) {
	response.Header().Set("Content-Type", "application/json; charset=utf-8")
	response.WriteHeader(status)
	_ = json.NewEncoder(response).Encode(payload)
}
