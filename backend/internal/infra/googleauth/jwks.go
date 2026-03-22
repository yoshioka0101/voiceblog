package googleauth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	jwksURL            = "https://www.googleapis.com/oauth2/v3/certs"
	defaultTTL         = time.Hour
	defaultHTTPTimeout = 5 * time.Second
)

type jwksResponse struct {
	Keys []jwkKey `json:"keys"`
}

type jwkKey struct {
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type jwksCache struct {
	mu        sync.RWMutex
	keys      map[string]*rsa.PublicKey
	expiresAt time.Time
	ttl       time.Duration
	client    *http.Client
	endpoint  string
	now       func() time.Time
}

func newJWKSCache() *jwksCache {
	return &jwksCache{
		keys:     make(map[string]*rsa.PublicKey),
		ttl:      defaultTTL,
		client:   &http.Client{Timeout: defaultHTTPTimeout},
		endpoint: jwksURL,
		now:      time.Now,
	}
}

func (c *jwksCache) getKey(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	c.mu.RLock()
	if c.now().Before(c.expiresAt) {
		key, ok := c.keys[kid]
		c.mu.RUnlock()
		if ok {
			return key, nil
		}
	} else {
		c.mu.RUnlock()
	}

	if err := c.refresh(ctx); err != nil {
		return nil, err
	}

	c.mu.RLock()
	defer c.mu.RUnlock()
	key, ok := c.keys[kid]
	if !ok {
		return nil, fmt.Errorf("unknown kid: %s", kid)
	}
	return key, nil
}

func (c *jwksCache) refresh(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint, nil)
	if err != nil {
		return fmt.Errorf("create jwks request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("fetch jwks: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("unexpected jwks status: %d", resp.StatusCode)
	}

	var jwks jwksResponse
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return fmt.Errorf("decode jwks: %w", err)
	}

	keys := make(map[string]*rsa.PublicKey, len(jwks.Keys))
	for _, k := range jwks.Keys {
		nBytes, err := base64.RawURLEncoding.DecodeString(k.N)
		if err != nil {
			return fmt.Errorf("decode jwk n: %w", err)
		}
		eBytes, err := base64.RawURLEncoding.DecodeString(k.E)
		if err != nil {
			return fmt.Errorf("decode jwk e: %w", err)
		}
		pub := &rsa.PublicKey{
			N: new(big.Int).SetBytes(nBytes),
			E: int(new(big.Int).SetBytes(eBytes).Int64()),
		}
		keys[k.Kid] = pub
	}

	ttl := c.ttl
	if maxAge, ok := maxAgeFromCacheControl(resp.Header.Get("Cache-Control")); ok {
		ttl = maxAge
	}

	c.mu.Lock()
	c.keys = keys
	c.expiresAt = c.now().Add(ttl)
	c.mu.Unlock()

	return nil
}

func maxAgeFromCacheControl(value string) (time.Duration, bool) {
	for _, directive := range strings.Split(value, ",") {
		key, raw, ok := strings.Cut(strings.TrimSpace(directive), "=")
		if !ok || !strings.EqualFold(key, "max-age") {
			continue
		}

		seconds, err := strconv.Atoi(raw)
		if err != nil || seconds < 0 {
			return 0, false
		}

		return time.Duration(seconds) * time.Second, true
	}

	return 0, false
}
