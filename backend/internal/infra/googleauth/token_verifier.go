package googleauth

import (
	"context"
	"fmt"

	"github.com/golang-jwt/jwt/v5"

	domainAuth "github.com/yoshioka0101/voiceblog/backend/internal/domain/auth"
)

type claims struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
	Name  string `json:"name"`
	jwt.RegisteredClaims
}

// TokenVerifier は Google id_token を検証する domain/auth.TokenVerifier の実装。
type TokenVerifier struct {
	cache    *jwksCache
	audience string
}

func NewTokenVerifier(audience string) *TokenVerifier {
	return &TokenVerifier{
		cache:    newJWKSCache(),
		audience: audience,
	}
}

func (v *TokenVerifier) Verify(_ context.Context, tokenStr string) (*domainAuth.VerifiedIdentity, error) {
	c := &claims{}
	_, err := jwt.ParseWithClaims(tokenStr, c, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("missing kid in token header")
		}
		return v.cache.getKey(kid)
	}, jwt.WithAudience(v.audience))
	if err != nil {
		return nil, fmt.Errorf("verify token: %w", err)
	}

	iss, err := c.GetIssuer()
	if err != nil || (iss != "accounts.google.com" && iss != "https://accounts.google.com") {
		return nil, fmt.Errorf("invalid issuer: %s", iss)
	}

	return &domainAuth.VerifiedIdentity{
		Provider: "google",
		Subject:  c.Sub,
		Email:    c.Email,
		Name:     c.Name,
	}, nil
}
