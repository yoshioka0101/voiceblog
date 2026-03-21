package usecase

import (
	"context"

	userdomain "github.com/yoshioka0101/voiceblog/backend/internal/feature/user/domain"
)

type UseCase struct {
	repo userdomain.Repository
}

func New(repo userdomain.Repository) *UseCase {
	return &UseCase{repo: repo}
}

func (uc *UseCase) FindOrCreate(ctx context.Context, provider, subject, email, name string) (*userdomain.User, error) {
	return uc.repo.FindOrCreate(ctx, provider, subject, email, name)
}
