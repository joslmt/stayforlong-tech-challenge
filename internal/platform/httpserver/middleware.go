package httpserver

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"
)

type contextKey string

const requestIDContextKey contextKey = "request_id"

// bufferedResponse prevents a handler panic from leaking a partial success
// response after its headers have already been written.
type bufferedResponse struct {
	header      http.Header
	body        bytes.Buffer
	status      int
	wroteHeader bool
}

func newBufferedResponse() *bufferedResponse {
	return &bufferedResponse{header: make(http.Header), status: http.StatusOK}
}

func (response *bufferedResponse) Header() http.Header { return response.header }

func (response *bufferedResponse) WriteHeader(status int) {
	if response.wroteHeader {
		return
	}
	response.status = status
	response.wroteHeader = true
}

func (response *bufferedResponse) Write(payload []byte) (int, error) {
	if !response.wroteHeader {
		response.WriteHeader(http.StatusOK)
	}
	return response.body.Write(payload)
}

func (response *bufferedResponse) flushTo(destination http.ResponseWriter) error {
	copyHeaders(destination.Header(), response.header)
	destination.WriteHeader(response.status)
	_, err := io.Copy(destination, &response.body)
	return err
}

func Middleware(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(destination http.ResponseWriter, request *http.Request) {
		response := newBufferedResponse()
		setSecurityHeaders(response.Header())

		requestID := request.Header.Get("X-Request-ID")
		if requestID == "" || len(requestID) > 128 {
			requestID = newRequestID()
		}
		response.Header().Set("X-Request-ID", requestID)
		request = request.WithContext(context.WithValue(request.Context(), requestIDContextKey, requestID))
		startedAt := time.Now()

		defer func() {
			if recovered := recover(); recovered != nil {
				logger.ErrorContext(request.Context(), "panic recovered",
					"request_id", requestID, "error_code", 9000, "panic", recovered, "stack", string(debug.Stack()))
				writeInternalServerError(destination, response.Header())
				response.status = http.StatusInternalServerError
			} else if err := response.flushTo(destination); err != nil {
				logger.ErrorContext(request.Context(), "response write failed", "request_id", requestID, "error", err)
			}

			if request.URL.Path != "/healthz" || response.status >= http.StatusBadRequest {
				logger.InfoContext(request.Context(), "request completed",
					"request_id", requestID, "method", request.Method, "path", request.URL.Path,
					"status", response.status, "duration_ms", time.Since(startedAt).Milliseconds())
			}
		}()

		next.ServeHTTP(response, request)
	})
}

func setSecurityHeaders(header http.Header) {
	header.Set("Content-Security-Policy", "default-src 'self'; connect-src 'self'; img-src 'self'; style-src 'self'; script-src 'self'; base-uri 'none'; frame-ancestors 'none'")
	header.Set("Referrer-Policy", "no-referrer")
	header.Set("X-Content-Type-Options", "nosniff")
	header.Set("X-Frame-Options", "DENY")
}

func writeInternalServerError(destination http.ResponseWriter, preserved http.Header) {
	copyHeaders(destination.Header(), preserved)
	destination.Header().Set("Content-Type", "application/json; charset=utf-8")
	destination.Header().Del("Content-Length")
	destination.WriteHeader(http.StatusInternalServerError)
	_ = json.NewEncoder(destination).Encode(map[string]string{"message": "Internal server error"})
}

func copyHeaders(destination, source http.Header) {
	for key := range destination {
		destination.Del(key)
	}
	for key, values := range source {
		destination[key] = append([]string(nil), values...)
	}
}

func newRequestID() string {
	buffer := make([]byte, 12)
	if _, err := rand.Read(buffer); err != nil {
		return time.Now().UTC().Format("20060102150405.000000000")
	}
	return hex.EncodeToString(buffer)
}
