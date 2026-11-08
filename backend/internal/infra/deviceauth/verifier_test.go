package deviceauth

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"testing"

	domainAuth "github.com/yoshioka0101/voiceblog/backend/internal/entity/auth"
)

func deviceToken(secret []byte) string {
	return TokenPrefix + base64.RawURLEncoding.EncodeToString(secret)
}

func TestVerifier_Verify_ValidToken(t *testing.T) {
	secret := bytes.Repeat([]byte{0xab}, SecretLength)

	got, err := NewVerifier().Verify(context.Background(), deviceToken(secret))
	if err != nil {
		t.Fatalf("Verify returned error: %v", err)
	}

	sum := sha256.Sum256(secret)
	if got.Provider != Provider {
		t.Errorf("Provider = %q, want %q", got.Provider, Provider)
	}
	if got.Subject != hex.EncodeToString(sum[:]) {
		t.Errorf("Subject = %q, want sha256 hex of secret", got.Subject)
	}
	if got.Email != "" || got.Name != "" {
		t.Errorf("Email/Name should be empty for device users, got %q / %q", got.Email, got.Name)
	}
}

func TestVerifier_Verify_SameSecretSameSubject(t *testing.T) {
	secret := bytes.Repeat([]byte{0x01}, SecretLength)
	v := NewVerifier()

	a, err := v.Verify(context.Background(), deviceToken(secret))
	if err != nil {
		t.Fatal(err)
	}
	b, err := v.Verify(context.Background(), deviceToken(secret))
	if err != nil {
		t.Fatal(err)
	}
	if a.Subject != b.Subject {
		t.Fatalf("subjects differ for the same secret: %q vs %q", a.Subject, b.Subject)
	}
}

func TestVerifier_Verify_InvalidTokens(t *testing.T) {
	cases := map[string]string{
		"no prefix":      base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{1}, SecretLength)),
		"empty secret":   TokenPrefix,
		"not base64url":  TokenPrefix + "!!!!",
		"too short":      deviceToken(bytes.Repeat([]byte{1}, SecretLength-1)),
		"too long":       deviceToken(bytes.Repeat([]byte{1}, SecretLength+1)),
		"padded base64":  TokenPrefix + base64.URLEncoding.EncodeToString(bytes.Repeat([]byte{1}, SecretLength)),
		"standard alpha": TokenPrefix + base64.RawStdEncoding.EncodeToString(bytes.Repeat([]byte{0xff}, SecretLength)),
	}

	for name, token := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := NewVerifier().Verify(context.Background(), token)
			if !errors.Is(err, ErrInvalidDeviceToken) {
				t.Fatalf("err = %v, want ErrInvalidDeviceToken", err)
			}
		})
	}
}

type stubVerifier struct {
	called bool
	result *domainAuth.VerifiedIdentity
}

func (s *stubVerifier) Verify(context.Context, string) (*domainAuth.VerifiedIdentity, error) {
	s.called = true
	return s.result, nil
}

func TestPrefixRouter_RoutesByPrefix(t *testing.T) {
	device := &stubVerifier{result: &domainAuth.VerifiedIdentity{Provider: "device"}}
	google := &stubVerifier{result: &domainAuth.VerifiedIdentity{Provider: "google"}}
	router := NewPrefixRouter(device, google)

	got, err := router.Verify(context.Background(), TokenPrefix+"anything")
	if err != nil {
		t.Fatal(err)
	}
	if got.Provider != "device" || !device.called || google.called {
		t.Fatalf("vbd_ token should go to device verifier only")
	}

	device.called = false
	got, err = router.Verify(context.Background(), "eyJhbGciOiJSUzI1NiJ9.payload.sig")
	if err != nil {
		t.Fatal(err)
	}
	if got.Provider != "google" || device.called || !google.called {
		t.Fatalf("JWT should go to fallback verifier only")
	}
}
