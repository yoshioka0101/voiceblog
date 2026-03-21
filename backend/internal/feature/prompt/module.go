package prompt

import (
	"database/sql"

	"github.com/yoshioka0101/voiceblog/backend/internal/feature/prompt/domain"
	"github.com/yoshioka0101/voiceblog/backend/internal/feature/prompt/infra/postgres"
	"github.com/yoshioka0101/voiceblog/backend/internal/feature/prompt/usecase"
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
