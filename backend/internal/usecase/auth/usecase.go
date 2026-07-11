package auth

import (
	"context"
	"fmt"
	"github.com/yoshioka0101/voiceblog/backend/internal/entity/repository"

	domainAuth "github.com/yoshioka0101/voiceblog/backend/internal/entity/auth"
	domainUser "github.com/yoshioka0101/voiceblog/backend/internal/entity/user"
)

type UseCase struct {
	verifier domainAuth.TokenVerifier
	userRepo repository.UserRepository
}

func NewUseCase(verifier domainAuth.TokenVerifier, userRepo repository.UserRepository) *UseCase {
	return &UseCase{verifier: verifier, userRepo: userRepo}
}

// Authenticate はトークンを検証し、対応するユーザーを返す（存在しなければ作成）。
func (uc *UseCase) Authenticate(ctx context.Context, token string) (*domainUser.User, error) {
	identity, err := uc.verifier.Verify(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("verify token: %w", err)
	}
	return uc.userRepo.FindOrCreate(ctx, identity.Provider, identity.Subject, identity.Email, identity.Name)
}
