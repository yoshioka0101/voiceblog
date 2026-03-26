package hatena

import (
	"context"
	"encoding/xml"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPublish(t *testing.T) {
	var receivedWSSE string
	var receivedBody string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedWSSE = r.Header.Get("X-WSSE")
		body, _ := io.ReadAll(r.Body)
		receivedBody = string(body)

		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`<?xml version="1.0" encoding="utf-8"?>
<entry xmlns="http://www.w3.org/2005/Atom">
  <id>tag:blog.hatena.ne.jp,2013:testuser-testblog-123</id>
  <link rel="alternate" href="https://testuser.hatenablog.com/entry/2026/03/26/test"/>
</entry>`))
	}))
	defer server.Close()

	client := &Client{
		baseURL:    server.URL,
		httpClient: server.Client(),
	}

	result, err := client.Publish(context.Background(), "testuser:testblog:testapikey", "Test Title", "Test Content")
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(receivedWSSE, "testuser") {
		t.Fatalf("WSSE should contain username, got %q", receivedWSSE)
	}

	var entry atomEntry
	if err := xml.Unmarshal([]byte(receivedBody[len(xml.Header):]), &entry); err != nil {
		// Try parsing the full body
		if err := xml.Unmarshal([]byte(receivedBody), &entry); err != nil {
			t.Fatalf("failed to parse request body: %v", err)
		}
	}
	if entry.Title != "Test Title" {
		t.Fatalf("got title %q, want %q", entry.Title, "Test Title")
	}

	if result.ID == "" {
		t.Fatal("expected non-empty ID")
	}
	if !strings.Contains(result.URL, "hatenablog.com") {
		t.Fatalf("got url %q, expected hatenablog.com URL", result.URL)
	}
}

func TestVerify(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/testuser/testblog/atom" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		wsse := r.Header.Get("X-WSSE")
		if !strings.Contains(wsse, "testuser") {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<?xml version="1.0" encoding="utf-8"?><service/>`))
	}))
	defer server.Close()

	client := &Client{
		baseURL:    server.URL,
		httpClient: server.Client(),
	}

	// Valid token
	if err := client.Verify(context.Background(), "testuser:testblog:testapikey"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Invalid format
	if err := client.Verify(context.Background(), "invalid-format"); err == nil {
		t.Fatal("expected error for invalid token format")
	}
}

func TestPublishInvalidToken(t *testing.T) {
	client := NewClient()
	_, err := client.Publish(context.Background(), "invalid-format", "Title", "Content")
	if err == nil {
		t.Fatal("expected error for invalid token format")
	}
}

func TestPublishError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Forbidden"))
	}))
	defer server.Close()

	client := &Client{
		baseURL:    server.URL,
		httpClient: server.Client(),
	}

	_, err := client.Publish(context.Background(), "user:blog:key", "Title", "Content")
	if err == nil {
		t.Fatal("expected error for 403")
	}
}
