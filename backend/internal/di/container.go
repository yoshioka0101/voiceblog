package di

import (
	"database/sql"

	domainUser "github.com/yoshioka0101/voiceblog/backend/internal/domain/user"
	"github.com/yoshioka0101/voiceblog/backend/internal/infra/googleauth"
	"github.com/yoshioka0101/voiceblog/backend/internal/infra/postgres"
	authUseCase "github.com/yoshioka0101/voiceblog/backend/internal/usecase/auth"
	userUseCase "github.com/yoshioka0101/voiceblog/backend/internal/usecase/user"
)

type Repositories struct {
	User domainUser.Repository
}

type UseCases struct {
	Auth *authUseCase.UseCase
	User *userUseCase.UseCase
}

type Container struct {
	Repositories *Repositories
	UseCases     *UseCases
}

func New(db *sql.DB, googleClientID string) *Container {
	userRepo := postgres.NewUserRepository(db)
	verifier := googleauth.NewTokenVerifier(googleClientID)

	repos := &Repositories{
		User: userRepo,
	}
	useCases := &UseCases{
		Auth: authUseCase.NewUseCase(verifier, repos.User),
		User: userUseCase.NewUseCase(repos.User),
	}

	return &Container{
		Repositories: repos,
		UseCases:     useCases,
	}
}
