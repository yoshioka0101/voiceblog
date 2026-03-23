package di

import (
	"database/sql"

	promptEntity "github.com/yoshioka0101/voiceblog/backend/internal/entity/prompt"
	transcriptionEntity "github.com/yoshioka0101/voiceblog/backend/internal/entity/transcription"
	domainUser "github.com/yoshioka0101/voiceblog/backend/internal/entity/user"
	"github.com/yoshioka0101/voiceblog/backend/internal/infra/googleauth"
	"github.com/yoshioka0101/voiceblog/backend/internal/repository/postgresql"
	authUseCase "github.com/yoshioka0101/voiceblog/backend/internal/usecase/auth"
	promptUseCase "github.com/yoshioka0101/voiceblog/backend/internal/usecase/prompt"
	transcriptionUseCase "github.com/yoshioka0101/voiceblog/backend/internal/usecase/transcription"
	userUseCase "github.com/yoshioka0101/voiceblog/backend/internal/usecase/user"
)

type Repositories struct {
	User          domainUser.Repository
	Prompt        promptEntity.Repository
	Transcription transcriptionEntity.Repository
}

type UseCases struct {
	Auth          *authUseCase.UseCase
	User          *userUseCase.UseCase
	Prompt        *promptUseCase.UseCase
	Transcription *transcriptionUseCase.UseCase
}

type Container struct {
	Repositories *Repositories
	UseCases     *UseCases
}

func New(db *sql.DB, googleClientID string) *Container {
	userRepo := postgresql.NewUserRepository(db)
	verifier := googleauth.NewTokenVerifier(googleClientID)
	promptRepo := postgresql.NewPromptRepository(db)
	transcriptionRepo := postgresql.NewTranscriptionRepository(db)

	repos := &Repositories{
		User:          userRepo,
		Prompt:        promptRepo,
		Transcription: transcriptionRepo,
	}
	useCases := &UseCases{
		Auth:          authUseCase.NewUseCase(verifier, repos.User),
		User:          userUseCase.NewUseCase(repos.User),
		Prompt:        promptUseCase.NewUseCase(repos.Prompt),
		Transcription: transcriptionUseCase.NewUseCase(repos.Transcription),
	}

	return &Container{
		Repositories: repos,
		UseCases:     useCases,
	}
}
