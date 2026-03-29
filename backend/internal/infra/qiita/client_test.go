package qiita

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Qiita keeps its own client tests because the contract is Bearer auth + JSON, unlike Hatena's WSSE + AtomPub flow.
// Ref: https://qiita.com/api/v2/docs
func TestPublish(t *testing.T) {
	var receivedAuth string
	var receivedBody createItemRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		if err := json.NewDecoder(r.Body).Decode(&receivedBody); err != nil {
			t.Fatal(err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(createItemResponse{
			ID:  "abc123",
			URL: "https://qiita.com/user/items/abc123",
		})
	}))
	defer server.Close()

	client := &Client{
		baseURL:    server.URL,
		httpClient: server.Client(),
	}

	result, err := client.Publish(context.Background(), "test-token", "Test Title", "Test Content")
	if err != nil {
		t.Fatal(err)
	}

	if receivedAuth != "Bearer test-token" {
		t.Fatalf("got auth %q, want %q", receivedAuth, "Bearer test-token")
	}
	if receivedBody.Title != "Test Title" {
		t.Fatalf("got title %q, want %q", receivedBody.Title, "Test Title")
	}
	if result.ID != "abc123" {
		t.Fatalf("got id %q, want %q", result.ID, "abc123")
	}
	if result.URL != "https://qiita.com/user/items/abc123" {
		t.Fatalf("got url %q, want %q", result.URL, "https://qiita.com/user/items/abc123")
	}
}

func TestVerify(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/authenticated_user" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer valid-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"user123"}`))
	}))
	defer server.Close()

	client := &Client{
		baseURL:    server.URL,
		httpClient: server.Client(),
	}

	// Valid token
	if err := client.Verify(context.Background(), "valid-token"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Invalid token
	if err := client.Verify(context.Background(), "bad-token"); err == nil {
		t.Fatal("expected error for invalid token")
	}
}

func TestPublishError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"Unauthorized"}`))
	}))
	defer server.Close()

	client := &Client{
		baseURL:    server.URL,
		httpClient: server.Client(),
	}

	_, err := client.Publish(context.Background(), "bad-token", "Title", "Content")
	if err == nil {
		t.Fatal("expected error for 401")
	}
}
