// Package deviceauth はログイン不要の iOS アプリ向けに、端末ごとの匿名ユーザーを認証する。
//
// iOS アプリは初回起動時に 32 byte の乱数（デバイス秘密値）を生成して Keychain に保存し、
// `Authorization: Bearer vbd_<base64url(秘密値)>` で API を呼ぶ。
// サーバーは秘密値そのものを保存せず、SHA-256 ハッシュをユーザーの auth_subject として使う。
package deviceauth

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"

	domainAuth "github.com/yoshioka0101/voiceblog/backend/internal/entity/auth"
)

const (
	// TokenPrefix はデバイストークンを Google ID トークンと見分けるための接頭辞。
	TokenPrefix = "vbd_"
	// Provider は users.auth_provider に保存する値。
	Provider = "device"
	// SecretLength はデバイス秘密値のバイト長。
	SecretLength = 32
)

var ErrInvalidDeviceToken = errors.New("invalid device token")

// Verifier はデバイストークンを検証する。
type Verifier struct{}

func NewVerifier() *Verifier {
	return &Verifier{}
}

func (v *Verifier) Verify(_ context.Context, token string) (*domainAuth.VerifiedIdentity, error) {
	encoded, ok := strings.CutPrefix(token, TokenPrefix)
	if !ok {
		return nil, ErrInvalidDeviceToken
	}

	secret, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil || len(secret) != SecretLength {
		return nil, ErrInvalidDeviceToken
	}

	sum := sha256.Sum256(secret)
	return &domainAuth.VerifiedIdentity{
		Provider: Provider,
		Subject:  hex.EncodeToString(sum[:]),
	}, nil
}

// PrefixRouter は `vbd_` で始まるトークンをデバイス検証へ、それ以外を fallback（Google）へ振り分ける。
type PrefixRouter struct {
	device   domainAuth.TokenVerifier
	fallback domainAuth.TokenVerifier
}

func NewPrefixRouter(device, fallback domainAuth.TokenVerifier) *PrefixRouter {
	return &PrefixRouter{device: device, fallback: fallback}
}

func (r *PrefixRouter) Verify(ctx context.Context, token string) (*domainAuth.VerifiedIdentity, error) {
	if strings.HasPrefix(token, TokenPrefix) {
		return r.device.Verify(ctx, token)
	}
	return r.fallback.Verify(ctx, token)
}
