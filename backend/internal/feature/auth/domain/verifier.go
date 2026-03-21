package domain

import "context"

// VerifiedIdentity はトークン検証後の認証結果を表す。
type VerifiedIdentity struct {
	Provider string
	Subject  string
	Email    string
	Name     string
}

// TokenVerifier は外部 IdP のトークンを検証する。
type TokenVerifier interface {
	Verify(ctx context.Context, token string) (*VerifiedIdentity, error)
}
