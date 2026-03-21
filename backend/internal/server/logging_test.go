package server

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequestLogger(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	router := gin.New()
	router.Use(requestLogger(logger))
	router.GET("/users/:id", func(c *gin.Context) {
		c.String(http.StatusCreated, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/users/42", nil)
	req.RemoteAddr = "192.0.2.10:1234"
	req.Header.Set("User-Agent", "voiceblog-test")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("unexpected status: %d", recorder.Code)
	}

	entry := decodeSingleJSONLog(t, &logs)
	if entry["msg"] != "request completed" {
		t.Fatalf("unexpected message: %v", entry["msg"])
	}
	if entry["level"] != "INFO" {
		t.Fatalf("unexpected level: %v", entry["level"])
	}
	if entry["method"] != http.MethodGet {
		t.Fatalf("unexpected method: %v", entry["method"])
	}
	if entry["path"] != "/users/42" {
		t.Fatalf("unexpected path: %v", entry["path"])
	}
	if entry["route"] != "/users/:id" {
		t.Fatalf("unexpected route: %v", entry["route"])
	}
	if entry["client_ip"] != "192.0.2.10" {
		t.Fatalf("unexpected client_ip: %v", entry["client_ip"])
	}
	if entry["user_agent"] != "voiceblog-test" {
		t.Fatalf("unexpected user_agent: %v", entry["user_agent"])
	}
	if entry["status"] != float64(http.StatusCreated) {
		t.Fatalf("unexpected status field: %v", entry["status"])
	}
}

func TestRecoveryLogger(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	router := gin.New()
	router.Use(recoveryLogger(logger))
	router.GET("/panic", func(c *gin.Context) {
		panic("boom")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	req.RemoteAddr = "192.0.2.11:1234"

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("unexpected status: %d", recorder.Code)
	}

	entry := decodeSingleJSONLog(t, &logs)
	if entry["msg"] != "panic recovered" {
		t.Fatalf("unexpected message: %v", entry["msg"])
	}
	if entry["level"] != "ERROR" {
		t.Fatalf("unexpected level: %v", entry["level"])
	}
	if entry["method"] != http.MethodGet {
		t.Fatalf("unexpected method: %v", entry["method"])
	}
	if entry["path"] != "/panic" {
		t.Fatalf("unexpected path: %v", entry["path"])
	}
	if entry["route"] != "/panic" {
		t.Fatalf("unexpected route: %v", entry["route"])
	}
	if entry["panic"] != "boom" {
		t.Fatalf("unexpected panic: %v", entry["panic"])
	}
	if entry["status"] != float64(http.StatusInternalServerError) {
		t.Fatalf("unexpected status field: %v", entry["status"])
	}
	if stack, ok := entry["stack"].(string); !ok || stack == "" {
		t.Fatalf("missing stack trace: %v", entry["stack"])
	}
}

func decodeSingleJSONLog(t *testing.T, reader io.Reader) map[string]any {
	t.Helper()

	decoder := json.NewDecoder(reader)
	var entry map[string]any
	if err := decoder.Decode(&entry); err != nil {
		t.Fatalf("decode log entry: %v", err)
	}

	return entry
}
