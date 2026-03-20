package user

import (
	"context"

	domain "github.com/yoshioka0101/voiceblog/backend/internal/domain/user"
)

type UseCase struct {
	repo domain.Repository
}

func NewUseCase(repo domain.Repository) *UseCase {
	return &UseCase{repo: repo}
}

func (uc *UseCase) FindOrCreate(ctx context.Context, provider, subject, email, name string) (*domain.User, error) {
	return uc.repo.FindOrCreate(ctx, provider, subject, email, name)
}
