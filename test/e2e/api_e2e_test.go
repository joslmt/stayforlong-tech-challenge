//go:build e2e

package e2e_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestAPI_GivenProductionContainer_WhenUsedByAClient_ThenCompleteExperienceWorks(t *testing.T) {
	// Given: the production Dockerfile is the system under test.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	container, err := testcontainers.Run(
		ctx,
		"",
		testcontainers.WithDockerfile(testcontainers.FromDockerfile{
			Context:    projectRoot(t),
			Dockerfile: "Dockerfile",
		}),
		testcontainers.WithExposedPorts("8080/tcp"),
		testcontainers.WithWaitStrategy(
			wait.ForHTTP("/healthz").WithPort("8080/tcp").WithStatusCodeMatcher(
				func(status int) bool { return status == http.StatusNoContent },
			).WithStartupTimeout(3*time.Minute),
		),
	)
	testcontainers.CleanupContainer(t, container)
	if err != nil {
		t.Fatalf("start production container: %v", err)
	}
	baseURL, err := container.Endpoint(ctx, "http")
	if err != nil {
		t.Fatalf("resolve container endpoint: %v", err)
	}
	client := &http.Client{Timeout: 10 * time.Second}

	// When / Then: the frontend and FAQ are delivered by the same container.
	assertPageContains(t, client, baseURL+"/", "Stayforlong Tech Challenge")
	assertPageContains(t, client, baseURL+"/faq", "Frequently asked questions")

	// When / Then: /data exposes profit-per-night values.
	dataResponse := postJSON(t, client, baseURL+"/data",
		`[{"request_id":"00000000-0000-4000-8000-000000000001","check_in":"2026-09-01","nights":1,"selling_rate":50,"margin":20}]`)
	assertStatus(t, dataResponse, http.StatusOK)
	var dataBody struct {
		Average float64 `json:"avg_night"`
		Minimum float64 `json:"min_night"`
		Maximum float64 `json:"max_night"`
	}
	decodeAndClose(t, dataResponse, &dataBody)
	if dataBody.Average != 10 || dataBody.Minimum != 10 || dataBody.Maximum != 10 {
		t.Fatalf("unexpected /data response: %#v", dataBody)
	}

	// When / Then: /revenue returns chronological selected IDs and selected statistics.
	revenueResponse := postJSON(t, client, baseURL+"/revenue", `[
		{"request_id":"00000000-0000-4000-8000-000000000001","check_in":"2026-02-18","nights":4,"selling_rate":160,"margin":30},
		{"request_id":"00000000-0000-4000-8000-000000000002","check_in":"2026-01-01","nights":4,"selling_rate":480,"margin":10}
	]`)
	assertStatus(t, revenueResponse, http.StatusOK)
	var revenueBody struct {
		RequestIDs  []string `json:"request_ids"`
		TotalProfit float64  `json:"total_profit"`
		Average     float64  `json:"avg_night"`
	}
	decodeAndClose(t, revenueResponse, &revenueBody)
	if strings.Join(revenueBody.RequestIDs, ",") != "00000000-0000-4000-8000-000000000002,00000000-0000-4000-8000-000000000001" ||
		revenueBody.TotalProfit != 96 || revenueBody.Average != 12 {
		t.Fatalf("unexpected /revenue response: %#v", revenueBody)
	}

	// When / Then: invalid input preserves the message-only 400 contract.
	errorResponse := postJSON(t, client, baseURL+"/data", `[]`)
	assertStatus(t, errorResponse, http.StatusBadRequest)
	var errorBody map[string]any
	decodeAndClose(t, errorResponse, &errorBody)
	if len(errorBody) != 1 || errorBody["message"] != "at least one booking request is required" {
		t.Fatalf("unexpected error response: %#v", errorBody)
	}

	// Edge cases exercise the same production image, not an in-memory substitute.
	t.Run("rounds a half cent per night half-up", func(t *testing.T) {
		response := postJSON(t, client, baseURL+"/data",
			`[{"request_id":"00000000-0000-4000-8000-000000000001","check_in":"2026-09-01","nights":2,"selling_rate":1,"margin":1}]`)
		assertStatus(t, response, http.StatusOK)
		var body map[string]float64
		decodeAndClose(t, response, &body)
		if body["avg_night"] != 0.01 {
			t.Fatalf("expected half-up €0.01, got %#v", body)
		}
	})

	t.Run("rejects duplicate request IDs", func(t *testing.T) {
		response := postJSON(t, client, baseURL+"/revenue", `[
			{"request_id":"00000000-0000-4000-8000-000000000001","check_in":"2026-09-01","nights":1,"selling_rate":100,"margin":10},
			{"request_id":"00000000-0000-4000-8000-000000000001","check_in":"2026-09-02","nights":1,"selling_rate":100,"margin":10}
		]`)
		assertMessageOnlyError(t, response, http.StatusBadRequest, "must be unique")
	})

	t.Run("rejects non-UUID-v4 request IDs", func(t *testing.T) {
		response := postJSON(t, client, baseURL+"/data",
			`[{"request_id":"bookata_XY123","check_in":"2026-09-01","nights":1,"selling_rate":100,"margin":10}]`)
		assertMessageOnlyError(t, response, http.StatusBadRequest, "valid UUID v4")
	})

	t.Run("rejects unknown JSON fields", func(t *testing.T) {
		response := postJSON(t, client, baseURL+"/data",
			`[{"request_id":"00000000-0000-4000-8000-000000000001","check_in":"2026-09-01","nights":1,"selling_rate":100,"margin":10,"extra":true}]`)
		assertMessageOnlyError(t, response, http.StatusBadRequest, "unknown field")
	})

	t.Run("accepts a checkout-date booking after cleaning", func(t *testing.T) {
		response := postJSON(t, client, baseURL+"/revenue", `[
			{"request_id":"00000000-0000-4000-8000-000000000001","check_in":"2026-09-01","nights":4,"selling_rate":400,"margin":10},
			{"request_id":"00000000-0000-4000-8000-000000000002","check_in":"2026-09-05","nights":2,"selling_rate":200,"margin":0}
		]`)
		assertStatus(t, response, http.StatusOK)
		var body struct {
			RequestIDs  []string `json:"request_ids"`
			TotalProfit float64  `json:"total_profit"`
		}
		decodeAndClose(t, response, &body)
		if strings.Join(body.RequestIDs, ",") != "00000000-0000-4000-8000-000000000001,00000000-0000-4000-8000-000000000002" || body.TotalProfit != 40 {
			t.Fatalf("expected both bookings including zero-margin occupancy, got %#v", body)
		}
	})
}

func projectRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve current test file")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", ".."))
}

func assertPageContains(t *testing.T, client *http.Client, url, expected string) {
	t.Helper()
	response, err := client.Get(url)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	assertStatus(t, response, http.StatusOK)
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read %s: %v", url, err)
	}
	closeResponse(t, response)
	if !strings.Contains(string(body), expected) {
		t.Fatalf("expected %s to contain %q", url, expected)
	}
}

func postJSON(t *testing.T, client *http.Client, url, body string) *http.Response {
	t.Helper()
	request, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, bytes.NewBufferString(body))
	if err != nil {
		t.Fatalf("create POST %s: %v", url, err)
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("POST %s: %v", url, err)
	}
	return response
}

func assertStatus(t *testing.T, response *http.Response, expected int) {
	t.Helper()
	if response.StatusCode != expected {
		body, _ := io.ReadAll(response.Body)
		closeResponse(t, response)
		t.Fatalf("expected HTTP %d, got %d: %s", expected, response.StatusCode, body)
	}
}

func decodeAndClose(t *testing.T, response *http.Response, target any) {
	t.Helper()
	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		closeResponse(t, response)
		t.Fatalf("decode response: %v", err)
	}
	closeResponse(t, response)
}

func assertMessageOnlyError(t *testing.T, response *http.Response, expectedStatus int, messagePart string) {
	t.Helper()
	assertStatus(t, response, expectedStatus)
	var body map[string]any
	decodeAndClose(t, response, &body)
	message, ok := body["message"].(string)
	if len(body) != 1 || !ok || !strings.Contains(message, messagePart) {
		t.Fatalf("expected message-only error containing %q, got %#v", messagePart, body)
	}
}

func closeResponse(t *testing.T, response *http.Response) {
	t.Helper()
	if err := response.Body.Close(); err != nil {
		t.Errorf("close response body: %v", err)
	}
}
