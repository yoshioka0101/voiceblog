package googleauth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestJWKSCacheRefresh_UsesCacheControlMaxAge(t *testing.T) {
	now := time.Date(2026, 3, 22, 0, 0, 0, 0, time.UTC)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=120")
		if err := json.NewEncoder(w).Encode(jwksResponse{Keys: []jwkKey{mustJWK(t, "kid-1")}}); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
	defer server.Close()

	cache := newJWKSCache()
	cache.client = server.Client()
	cache.endpoint = server.URL
	cache.now = func() time.Time { return now }

	if err := cache.refresh(context.Background()); err != nil {
		t.Fatalf("refresh failed: %v", err)
	}

	want := now.Add(120 * time.Second)
	if !cache.expiresAt.Equal(want) {
		t.Fatalf("expiresAt = %v, want %v", cache.expiresAt, want)
	}
}

func TestJWKSCacheRefresh_UsesFallbackTTLWithoutMaxAge(t *testing.T) {
	now := time.Date(2026, 3, 22, 0, 0, 0, 0, time.UTC)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if err := json.NewEncoder(w).Encode(jwksResponse{Keys: []jwkKey{mustJWK(t, "kid-1")}}); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
	defer server.Close()

	cache := newJWKSCache()
	cache.client = server.Client()
	cache.endpoint = server.URL
	cache.now = func() time.Time { return now }

	if err := cache.refresh(context.Background()); err != nil {
		t.Fatalf("refresh failed: %v", err)
	}

	want := now.Add(defaultTTL)
	if !cache.expiresAt.Equal(want) {
		t.Fatalf("expiresAt = %v, want %v", cache.expiresAt, want)
	}
}

func TestJWKSCacheRefresh_RejectsNonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	cache := newJWKSCache()
	cache.client = server.Client()
	cache.endpoint = server.URL

	if err := cache.refresh(context.Background()); err == nil {
		t.Fatal("expected refresh to fail")
	}
}

func TestJWKSCacheGetKey_RespectsContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if err := json.NewEncoder(w).Encode(jwksResponse{Keys: []jwkKey{mustJWK(t, "kid-1")}}); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
	defer server.Close()

	cache := newJWKSCache()
	cache.client = server.Client()
	cache.endpoint = server.URL

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := cache.getKey(ctx, "kid-1")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
}

func mustJWK(t *testing.T, kid string) jwkKey {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate rsa key: %v", err)
	}

	return jwkKey{
		Kid: kid,
		N:   base64.RawURLEncoding.EncodeToString(privateKey.PublicKey.N.Bytes()),
		E:   base64.RawURLEncoding.EncodeToString(bigEndianBytes(privateKey.PublicKey.E)),
	}
}

func bigEndianBytes(value int) []byte {
	switch {
	case value <= 0xFF:
		return []byte{byte(value)}
	case value <= 0xFFFF:
		return []byte{byte(value >> 8), byte(value)}
	default:
		return []byte{byte(value >> 16), byte(value >> 8), byte(value)}
	}
}
