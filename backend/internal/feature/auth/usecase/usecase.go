package usecase

import (
	"context"
	"fmt"

	authdomain "github.com/yoshioka0101/voiceblog/backend/internal/feature/auth/domain"
	userdomain "github.com/yoshioka0101/voiceblog/backend/internal/feature/user/domain"
)

type UseCase struct {
	verifier authdomain.TokenVerifier
	userRepo userdomain.Repository
}

func New(verifier authdomain.TokenVerifier, userRepo userdomain.Repository) *UseCase {
	return &UseCase{verifier: verifier, userRepo: userRepo}
}

// Authenticate はトークンを検証し、対応するユーザーを返す（存在しなければ作成）。
func (uc *UseCase) Authenticate(ctx context.Context, token string) (*userdomain.User, error) {
	identity, err := uc.verifier.Verify(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("verify token: %w", err)
	}
	return uc.userRepo.FindOrCreate(ctx, identity.Provider, identity.Subject, identity.Email, identity.Name)
}
