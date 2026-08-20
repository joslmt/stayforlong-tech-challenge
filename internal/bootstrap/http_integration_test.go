package bootstrap_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	bookinghttp "github.com/joslmt/stayforlong-tech-challenge/internal/booking/adapters/in/http"
	"github.com/joslmt/stayforlong-tech-challenge/internal/bootstrap"
	"github.com/joslmt/stayforlong-tech-challenge/internal/platform/httpserver"
)

func TestData_GivenValidPayload_WhenPosted_ThenReturnsMetrics(t *testing.T) {
	// Given
	server := newServer(t)
	payload := `[{"request_id":"00000000-0000-4000-8000-000000000001","check_in":"2026-09-01","nights":3,"selling_rate":100,"margin":20}]`

	// When
	response := postJSON(t, server.URL+"/data", payload)

	// Then
	defer closeResponseBody(t, response)
	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", response.StatusCode, readBody(t, response))
	}
	var body map[string]float64
	decodeBody(t, response, &body)
	if body["avg_night"] != 6.67 || body["min_night"] != 6.67 || body["max_night"] != 6.67 {
		t.Fatalf("unexpected metrics: %#v", body)
	}
}

func TestRevenue_GivenConflicts_WhenPosted_ThenReturnsChronologicalSelectedSchedule(t *testing.T) {
	// Given
	server := newServer(t)
	payload := `[
		{"request_id":"00000000-0000-4000-8000-000000000001","check_in":"2026-09-05","nights":3,"selling_rate":300,"margin":10},
		{"request_id":"00000000-0000-4000-8000-000000000002","check_in":"2026-09-01","nights":4,"selling_rate":400,"margin":10},
		{"request_id":"00000000-0000-4000-8000-000000000003","check_in":"2026-09-03","nights":6,"selling_rate":900,"margin":50}
	]`

	// When
	response := postJSON(t, server.URL+"/revenue", payload)

	// Then
	defer closeResponseBody(t, response)
	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", response.StatusCode, readBody(t, response))
	}
	var body struct {
		RequestIDs []string `json:"request_ids"`
		Total      float64  `json:"total_profit"`
	}
	decodeBody(t, response, &body)
	if strings.Join(body.RequestIDs, ",") != "00000000-0000-4000-8000-000000000002,00000000-0000-4000-8000-000000000001" || body.Total != 70 {
		t.Fatalf("unexpected revenue response: %#v", body)
	}
}

func TestData_GivenUnknownField_WhenPosted_ThenReturnsHumanFriendly400(t *testing.T) {
	server := newServer(t)
	payload := `[{"request_id":"00000000-0000-4000-8000-000000000001","check_in":"2026-09-01","nights":1,"selling_rate":100,"margin":20,"unexpected":true}]`
	response := postJSON(t, server.URL+"/data", payload)
	defer closeResponseBody(t, response)
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.StatusCode)
	}
	var body map[string]any
	decodeBody(t, response, &body)
	if len(body) != 1 || body["message"] == "" {
		t.Fatalf("expected message-only error, got %#v", body)
	}
}

func TestData_GivenSameUUIDWithDifferentCase_WhenPosted_ThenReturnsHumanFriendly400(t *testing.T) {
	server := newServer(t)
	payload := `[
		{"request_id":"550e8400-e29b-41d4-a716-446655440000","check_in":"2026-09-01","nights":1,"selling_rate":100,"margin":20},
		{"request_id":"550E8400-E29B-41D4-A716-446655440000","check_in":"2026-09-02","nights":1,"selling_rate":100,"margin":20}
	]`
	response := postJSON(t, server.URL+"/data", payload)
	defer closeResponseBody(t, response)
	if response.StatusCode != http.StatusBadRequest || !strings.Contains(readBody(t, response), "must be unique") {
		t.Fatalf("expected duplicate ID message, got status %d", response.StatusCode)
	}
}

func TestEndpoints_GivenInvalidEdgeCases_WhenPosted_ThenReturnMessageOnly400(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		contentType string
		body        string
		messagePart string
	}{
		{"missing content type", "/data", "text/plain", `[]`, "Content-Type must be application/json"},
		{"empty body", "/data", "application/json", ``, "body is empty"},
		{"malformed JSON", "/data", "application/json", `[`, "valid JSON array"},
		{"syntax error", "/data", "application/json", `[{]`, "invalid JSON near byte"},
		{"invalid field type", "/data", "application/json", `[{"request_id":false}]`, "field 0.request_id has an invalid type"},
		{"trailing JSON", "/data", "application/json", `[] []`, "exactly one JSON array"},
		{"empty list", "/data", "application/json", `[]`, "at least one booking"},
		{"non-UUID request ID", "/data", "application/json", `[{"request_id":"bookata_XY123","check_in":"2026-09-01","nights":1,"selling_rate":100,"margin":20}]`, "valid UUID v4"},
		{"unknown field", "/data", "application/json", `[{"request_id":"00000000-0000-4000-8000-000000000001","check_in":"2026-09-01","nights":1,"selling_rate":100,"margin":20,"extra":true}]`, "unknown field"},
		{"invalid date", "/revenue", "application/json", `[{"request_id":"00000000-0000-4000-8000-000000000001","check_in":"2026-02-30","nights":1,"selling_rate":100,"margin":20}]`, "valid date"},
		{"zero nights", "/revenue", "application/json", `[{"request_id":"00000000-0000-4000-8000-000000000001","check_in":"2026-09-01","nights":0,"selling_rate":100,"margin":20}]`, "nights must be"},
		{"negative rate", "/data", "application/json", `[{"request_id":"00000000-0000-4000-8000-000000000001","check_in":"2026-09-01","nights":1,"selling_rate":-1,"margin":20}]`, "selling_rate must be"},
		{"margin over 100", "/data", "application/json", `[{"request_id":"00000000-0000-4000-8000-000000000001","check_in":"2026-09-01","nights":1,"selling_rate":100,"margin":101}]`, "margin must be"},
	}

	server := newServer(t)
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			request, err := http.NewRequest(http.MethodPost, server.URL+test.path, bytes.NewBufferString(test.body))
			if err != nil {
				t.Fatalf("create request: %v", err)
			}
			request.Header.Set("Content-Type", test.contentType)

			// When
			response, err := http.DefaultClient.Do(request)
			if err != nil {
				t.Fatalf("execute request: %v", err)
			}

			// Then
			defer closeResponseBody(t, response)
			var body map[string]any
			decodeBody(t, response, &body)
			if response.StatusCode != http.StatusBadRequest || len(body) != 1 ||
				!strings.Contains(body["message"].(string), test.messagePart) {
				t.Fatalf("expected message-only 400 containing %q, got %d %#v", test.messagePart, response.StatusCode, body)
			}
		})
	}
}

func TestData_GivenBodyOverLimit_WhenPosted_ThenReturnsHumanFriendly400(t *testing.T) {
	// Given
	server := newServer(t)
	request, err := http.NewRequest(http.MethodPost, server.URL+"/data", strings.NewReader(strings.Repeat(" ", int(bookinghttp.MaxBodyBytes)+1)))
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	request.Header.Set("Content-Type", "application/json")

	// When
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("execute request: %v", err)
	}
	defer closeResponseBody(t, response)

	// Then
	if response.StatusCode != http.StatusBadRequest || !strings.Contains(readBody(t, response), "must not exceed 5 MiB") {
		t.Fatalf("expected body-limit 400, got %d", response.StatusCode)
	}
}

func TestData_GivenHalfCentPerNight_WhenPosted_ThenRoundsHalfUpAtResponseBoundary(t *testing.T) {
	// Given: €1 at 1% is one cent profit across two nights, or half a cent per night.
	server := newServer(t)
	payload := `[{"request_id":"00000000-0000-4000-8000-000000000001","check_in":"2026-09-01","nights":2,"selling_rate":1,"margin":1}]`

	// When
	response := postJSON(t, server.URL+"/data", payload)
	defer closeResponseBody(t, response)
	var body map[string]float64
	decodeBody(t, response, &body)

	// Then
	if response.StatusCode != http.StatusOK || body["avg_night"] != 0.01 {
		t.Fatalf("expected half-up €0.01 average, got %d %#v", response.StatusCode, body)
	}
}

func TestRevenue_GivenZeroMarginBooking_WhenItAddsOccupancy_ThenSelectsIt(t *testing.T) {
	server := newServer(t)
	payload := `[{"request_id":"00000000-0000-4000-8000-000000000001","check_in":"2020-01-01","nights":2,"selling_rate":100,"margin":0}]`
	response := postJSON(t, server.URL+"/revenue", payload)
	defer closeResponseBody(t, response)
	var body struct {
		RequestIDs []string `json:"request_ids"`
		Total      float64  `json:"total_profit"`
	}
	decodeBody(t, response, &body)
	if response.StatusCode != http.StatusOK || strings.Join(body.RequestIDs, ",") != "00000000-0000-4000-8000-000000000001" || body.Total != 0 {
		t.Fatalf("expected zero-margin booking to be selected, got %d %#v", response.StatusCode, body)
	}
}

func TestMiddleware_GivenPanic_WhenServing_ThenReturnsSafe500(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := httpserver.Middleware(logger, http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		response.WriteHeader(http.StatusNoContent)
		_, _ = response.Write([]byte("partial sensitive response"))
		panic("sensitive detail")
	}))
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/panic", nil))
	if recorder.Code != http.StatusInternalServerError || recorder.Body.String() != "{\"message\":\"Internal server error\"}\n" {
		t.Fatalf("expected safe 500, got %d %s", recorder.Code, recorder.Body.String())
	}
}

func TestHealth_GivenHealthyApplication_WhenProbed_ThenReturnsQuiet204(t *testing.T) {
	// Given
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil))
	handler := bootstrap.NewHTTPHandler(logger)
	recorder := httptest.NewRecorder()

	// When
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	// Then
	if recorder.Code != http.StatusNoContent || recorder.Body.Len() != 0 {
		t.Fatalf("expected empty 204 health response, got %d %q", recorder.Code, recorder.Body.String())
	}
	if logs.Len() != 0 {
		t.Fatalf("expected successful health probe not to generate access logs, got %s", logs.String())
	}
}

func TestFrontend_GivenRootAndFAQ_WhenRequested_ThenBothAreAvailable(t *testing.T) {
	server := newServer(t)
	for _, path := range []string{"/", "/faq"} {
		response, err := http.Get(server.URL + path)
		if err != nil {
			t.Fatalf("GET %s: %v", path, err)
		}
		body := readBody(t, response)
		closeResponseBody(t, response)
		if response.StatusCode != http.StatusOK || !strings.Contains(body, "Stayforlong Tech Challenge") {
			t.Fatalf("expected frontend at %s, got %d", path, response.StatusCode)
		}
		if response.Header.Get("Content-Security-Policy") == "" || response.Header.Get("X-Content-Type-Options") != "nosniff" {
			t.Fatalf("expected browser security headers at %s", path)
		}
	}
}

func TestFrontend_GivenPlaygroundAssets_WhenRequested_ThenAPIExecutionFeedbackIsIncluded(t *testing.T) {
	server := newServer(t)

	pageResponse, err := http.Get(server.URL + "/")
	if err != nil {
		t.Fatalf("GET frontend: %v", err)
	}
	page := readBody(t, pageResponse)
	closeResponseBody(t, pageResponse)
	if !strings.Contains(page, `id="execution-feedback"`) {
		t.Fatal("expected accessible execution feedback in playground")
	}

	scriptResponse, err := http.Get(server.URL + "/app.js")
	if err != nil {
		t.Fatalf("GET frontend script: %v", err)
	}
	script := readBody(t, scriptResponse)
	closeResponseBody(t, scriptResponse)
	for _, expected := range []string{`data: "/data"`, `revenue: "/revenue"`, "Request completed successfully", "Request failed with HTTP"} {
		if !strings.Contains(script, expected) {
			t.Fatalf("expected frontend script to contain %q", expected)
		}
	}
}

func newServer(t *testing.T) *httptest.Server {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	server := httptest.NewServer(bootstrap.NewHTTPHandler(logger))
	t.Cleanup(server.Close)
	return server
}

func postJSON(t *testing.T, url, body string) *http.Response {
	t.Helper()
	response, err := http.Post(url, "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatalf("POST %s: %v", url, err)
	}
	return response
}

func decodeBody(t *testing.T, response *http.Response, target any) {
	t.Helper()
	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}

func readBody(t *testing.T, response *http.Response) string {
	t.Helper()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}
	return string(body)
}

func closeResponseBody(t *testing.T, response *http.Response) {
	t.Helper()
	if err := response.Body.Close(); err != nil {
		t.Errorf("close response body: %v", err)
	}
}
