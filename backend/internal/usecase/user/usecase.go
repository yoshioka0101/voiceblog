package user

import (
	"context"

	entity "github.com/yoshioka0101/voiceblog/backend/internal/entity/user"
)

type UseCase struct {
	repo entity.Repository
}

func NewUseCase(repo entity.Repository) *UseCase {
	return &UseCase{repo: repo}
}

func (uc *UseCase) FindOrCreate(ctx context.Context, provider, subject, email, name string) (*entity.User, error) {
	return uc.repo.FindOrCreate(ctx, provider, subject, email, name)
}
