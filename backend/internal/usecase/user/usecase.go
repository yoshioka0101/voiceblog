package user

import (
	"context"
	"github.com/yoshioka0101/voiceblog/backend/internal/entity/repository"

	entity "github.com/yoshioka0101/voiceblog/backend/internal/entity/user"
)

type UseCase struct {
	repo repository.UserRepository
}

func NewUseCase(repo repository.UserRepository) *UseCase {
	return &UseCase{repo: repo}
}

func (uc *UseCase) FindOrCreate(ctx context.Context, provider, subject, email, name string) (*entity.User, error) {
	return uc.repo.FindOrCreate(ctx, provider, subject, email, name)
}
