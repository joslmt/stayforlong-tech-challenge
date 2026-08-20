package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthcheck_GivenConfiguredHealthyURL_WhenChecking_ThenSucceeds(t *testing.T) {
	// Given
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/custom-health" {
			t.Fatalf("expected configured health path, got %s", request.URL.Path)
		}
		response.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	// When / Then
	if err := healthcheck(server.URL + "/custom-health"); err != nil {
		t.Fatalf("expected healthy configured URL: %v", err)
	}
}

func TestHealthcheck_GivenUnhealthyStatus_WhenChecking_ThenFails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		response.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	err := healthcheck(server.URL)
	if err == nil || !strings.Contains(err.Error(), "status 503") {
		t.Fatalf("expected status error, got %v", err)
	}
}
