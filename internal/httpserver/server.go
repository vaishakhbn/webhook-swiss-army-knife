package httpserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

const defaultPort = "10000"

// NewHandler builds the HTTP API. Add future Swiss-army-knife endpoints here.
func NewHandler(logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"name": "webhook-swiss-army-knife",
			"endpoints": map[string]string{
				"POST /authorization":              "logs the Authorization request header",
				"POST /authorization/{request_id}": "logs the header with a request identifier",
				"GET /healthz":                     "health check",
			},
		})
	})

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	logAuthorization := func(w http.ResponseWriter, r *http.Request) {
		authorization := r.Header.Get("Authorization")
		requestID := r.PathValue("requestID")
		logger.Info("authorization header received",
			"authorization", authorization,
			"present", authorization != "",
			"request_id", requestID,
			"client_ip", clientIP(r),
			"x_forwarded_for", r.Header.Get("X-Forwarded-For"),
			"cf_ray", r.Header.Get("CF-Ray"),
			"remote_addr", r.RemoteAddr,
		)

		writeJSON(w, http.StatusOK, map[string]any{
			"logged":     authorization != "",
			"request_id": requestID,
		})
	}

	mux.HandleFunc("POST /authorization", logAuthorization)
	mux.HandleFunc("POST /authorization/{requestID}", logAuthorization)

	return securityHeaders(mux)
}

// clientIP returns the original client address supplied by Render's proxy. Render
// places it first in X-Forwarded-For. RemoteAddr remains the fallback for local
// development and requests that do not pass through Render's public proxy.
func clientIP(r *http.Request) string {
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		first, _, _ := strings.Cut(forwardedFor, ",")
		if ip := net.ParseIP(strings.TrimSpace(first)); ip != nil {
			return ip.String()
		}
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}

// Run starts the production HTTP server on Render's PORT, or 10000 locally.
func Run(logger *slog.Logger) error {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           NewHandler(logger),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	logger.Info("server listening", "address", server.Addr)
	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("listen: %w", err)
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}
