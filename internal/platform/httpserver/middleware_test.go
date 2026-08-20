package httpserver

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBufferedResponse_GivenImplicitStatus_WhenWriting_ThenFlushesCompleteResponse(t *testing.T) {
	// Given
	buffer := newBufferedResponse()
	buffer.Header().Set("X-Test", "preserved")

	// When
	if _, err := buffer.Write([]byte("complete body")); err != nil {
		t.Fatalf("write buffer: %v", err)
	}
	destination := httptest.NewRecorder()
	destination.Header().Set("X-Stale", "remove")
	if err := buffer.flushTo(destination); err != nil {
		t.Fatalf("flush buffer: %v", err)
	}

	// Then
	if destination.Code != http.StatusOK || destination.Body.String() != "complete body" {
		t.Fatalf("unexpected response: %d %q", destination.Code, destination.Body.String())
	}
	if destination.Header().Get("X-Test") != "preserved" || destination.Header().Get("X-Stale") != "" {
		t.Fatalf("unexpected headers: %#v", destination.Header())
	}
}

func TestBufferedResponse_GivenRepeatedStatus_WhenWriting_ThenKeepsFirstStatus(t *testing.T) {
	buffer := newBufferedResponse()
	buffer.WriteHeader(http.StatusCreated)
	buffer.WriteHeader(http.StatusNoContent)

	if buffer.status != http.StatusCreated {
		t.Fatalf("expected first status %d, got %d", http.StatusCreated, buffer.status)
	}
}
