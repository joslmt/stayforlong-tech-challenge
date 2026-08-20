package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joslmt/stayforlong-tech-challenge/internal/bootstrap"
)

const defaultAddress = ":8080"
const defaultHealthcheckURL = "http://127.0.0.1:8080/healthz"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		if err := healthcheck(environmentOrDefault("HEALTHCHECK_URL", defaultHealthcheckURL)); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	address := environmentOrDefault("HTTP_ADDRESS", defaultAddress)
	server := &http.Server{
		Addr: address, Handler: bootstrap.NewHTTPHandler(logger),
		ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second,
		WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second,
	}

	shutdownSignals, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	go func() {
		<-shutdownSignals.Done()
		shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownContext); err != nil {
			logger.Error("graceful shutdown failed", "error", err)
		}
	}()

	logger.Info("server starting", "address", address)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("server stopped unexpectedly", "error", err)
		os.Exit(1)
	}
}

func healthcheck(url string) error {
	client := &http.Client{Timeout: 2 * time.Second}
	response, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("healthcheck request: %w", err)
	}
	statusCode := response.StatusCode
	if err := response.Body.Close(); err != nil {
		return fmt.Errorf("healthcheck close response: %w", err)
	}
	if statusCode != http.StatusNoContent {
		return fmt.Errorf("healthcheck returned status %d", statusCode)
	}
	return nil
}

func environmentOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
