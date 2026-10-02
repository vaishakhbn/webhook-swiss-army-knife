package httpserver

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAuthorizationEndpointLogsHeader(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil))
	request := httptest.NewRequest(http.MethodPost, "/authorization", nil)
	request.Header.Set("Authorization", "Bearer test-token")
	recorder := httptest.NewRecorder()

	NewHandler(logger).ServeHTTP(recorder, request)
	response := recorder.Result()
	t.Cleanup(func() { _ = response.Body.Close() })

	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d; want %d", response.StatusCode, http.StatusOK)
	}
	if !strings.Contains(logs.String(), `"authorization":"Bearer test-token"`) {
		t.Fatalf("log did not contain authorization header: %s", logs.String())
	}

	var body struct {
		Logged bool `json:"logged"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if !body.Logged {
		t.Fatal("logged = false; want true")
	}
}

func TestAuthorizationEndpointLogsRequestID(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil))
	request := httptest.NewRequest(http.MethodPost, "/authorization/550e8400-e29b-41d4-a716-446655440000", nil)
	request.Header.Set("Authorization", "Basic dXNlcjpwYXNz")
	recorder := httptest.NewRecorder()

	NewHandler(logger).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d; want %d", recorder.Code, http.StatusOK)
	}
	if !strings.Contains(logs.String(), `"authorization":"Basic dXNlcjpwYXNz"`) {
		t.Fatalf("log did not contain authorization header: %s", logs.String())
	}
	if !strings.Contains(logs.String(), `"request_id":"550e8400-e29b-41d4-a716-446655440000"`) {
		t.Fatalf("log did not contain request ID: %s", logs.String())
	}
}

func TestAuthorizationEndpointLogsRenderClientIP(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil))
	request := httptest.NewRequest(http.MethodPost, "/authorization", nil)
	request.Header.Set("X-Forwarded-For", "203.0.113.42, 198.51.100.7")
	request.Header.Set("CF-Ray", "abc123-SJC")
	request.RemoteAddr = "10.0.0.5:4321"
	recorder := httptest.NewRecorder()

	NewHandler(logger).ServeHTTP(recorder, request)

	if !strings.Contains(logs.String(), `"client_ip":"203.0.113.42"`) {
		t.Fatalf("log did not contain Render client IP: %s", logs.String())
	}
	if !strings.Contains(logs.String(), `"x_forwarded_for":"203.0.113.42, 198.51.100.7"`) {
		t.Fatalf("log did not contain forwarded chain: %s", logs.String())
	}
	if !strings.Contains(logs.String(), `"cf_ray":"abc123-SJC"`) {
		t.Fatalf("log did not contain CF-Ray ID: %s", logs.String())
	}
}

func TestAuthorizationEndpointFallsBackToRemoteIP(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil))
	request := httptest.NewRequest(http.MethodPost, "/authorization", nil)
	request.RemoteAddr = "192.0.2.10:4321"
	recorder := httptest.NewRecorder()

	NewHandler(logger).ServeHTTP(recorder, request)

	if !strings.Contains(logs.String(), `"client_ip":"192.0.2.10"`) {
		t.Fatalf("log did not contain remote client IP fallback: %s", logs.String())
	}
}

func TestOnlyPOSTIsAccepted(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/authorization", nil)

	NewHandler(logger).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d; want %d", recorder.Code, http.StatusMethodNotAllowed)
	}
}

func TestHealth(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	NewHandler(logger).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d; want %d", recorder.Code, http.StatusOK)
	}
}
