package crypto

import (
	"crypto/rand"
	"encoding/hex"
	"testing"
)

func generateTestKey(t *testing.T) string {
	t.Helper()
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatal(err)
	}
	return hex.EncodeToString(key)
}

func TestRoundTrip(t *testing.T) {
	enc, err := NewAESEncryptor(generateTestKey(t))
	if err != nil {
		t.Fatal(err)
	}

	plaintext := []byte("my-secret-api-token")
	ciphertext, nonce, err := enc.Encrypt(plaintext)
	if err != nil {
		t.Fatal(err)
	}

	decrypted, err := enc.Decrypt(ciphertext, nonce)
	if err != nil {
		t.Fatal(err)
	}

	if string(decrypted) != string(plaintext) {
		t.Fatalf("got %q, want %q", decrypted, plaintext)
	}
}

func TestDifferentNonces(t *testing.T) {
	enc, err := NewAESEncryptor(generateTestKey(t))
	if err != nil {
		t.Fatal(err)
	}

	plaintext := []byte("same-text")
	ct1, n1, _ := enc.Encrypt(plaintext)
	ct2, n2, _ := enc.Encrypt(plaintext)

	if hex.EncodeToString(ct1) == hex.EncodeToString(ct2) {
		t.Fatal("two encryptions of same plaintext should produce different ciphertext")
	}
	if hex.EncodeToString(n1) == hex.EncodeToString(n2) {
		t.Fatal("nonces should differ")
	}
}

func TestWrongKey(t *testing.T) {
	enc1, _ := NewAESEncryptor(generateTestKey(t))
	enc2, _ := NewAESEncryptor(generateTestKey(t))

	ciphertext, nonce, _ := enc1.Encrypt([]byte("secret"))
	_, err := enc2.Decrypt(ciphertext, nonce)
	if err == nil {
		t.Fatal("decryption with wrong key should fail")
	}
}

func TestInvalidKeyLength(t *testing.T) {
	_, err := NewAESEncryptor("aabbcc")
	if err == nil {
		t.Fatal("should reject short key")
	}
}

func TestInvalidKeyHex(t *testing.T) {
	_, err := NewAESEncryptor("not-hex-at-all-not-hex-at-all-not-hex-at-all-not-hex-at-all-1234")
	if err == nil {
		t.Fatal("should reject invalid hex")
	}
}
