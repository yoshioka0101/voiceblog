package user

import (
	"database/sql"

	"github.com/yoshioka0101/voiceblog/backend/internal/feature/user/domain"
	"github.com/yoshioka0101/voiceblog/backend/internal/feature/user/infra/postgres"
	"github.com/yoshioka0101/voiceblog/backend/internal/feature/user/usecase"
)

type Feature struct {
	Repository domain.Repository
	UseCase    *usecase.UseCase
}

func New(db *sql.DB) *Feature {
	repo := postgres.NewRepository(db)

	return &Feature{
		Repository: repo,
		UseCase:    usecase.New(repo),
	}
}
