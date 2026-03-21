package auth

import (
	"github.com/yoshioka0101/voiceblog/backend/internal/feature/auth/infra/googleauth"
	"github.com/yoshioka0101/voiceblog/backend/internal/feature/auth/usecase"
	userdomain "github.com/yoshioka0101/voiceblog/backend/internal/feature/user/domain"
)

type Feature struct {
	UseCase *usecase.UseCase
}

func New(googleClientID string, userRepo userdomain.Repository) *Feature {
	verifier := googleauth.NewTokenVerifier(googleClientID)

	return &Feature{
		UseCase: usecase.New(verifier, userRepo),
	}
}
