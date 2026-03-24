package di

import (
	"database/sql"

	articleEntity "github.com/yoshioka0101/voiceblog/backend/internal/entity/article"
	promptEntity "github.com/yoshioka0101/voiceblog/backend/internal/entity/prompt"
	promptRunJobEntity "github.com/yoshioka0101/voiceblog/backend/internal/entity/promptrunjob"
	transcriptionEntity "github.com/yoshioka0101/voiceblog/backend/internal/entity/transcription"
	domainUser "github.com/yoshioka0101/voiceblog/backend/internal/entity/user"
	"github.com/yoshioka0101/voiceblog/backend/internal/infra/gemini"
	"github.com/yoshioka0101/voiceblog/backend/internal/infra/googleauth"
	"github.com/yoshioka0101/voiceblog/backend/internal/repository/postgresql"
	articleUseCase "github.com/yoshioka0101/voiceblog/backend/internal/usecase/article"
	authUseCase "github.com/yoshioka0101/voiceblog/backend/internal/usecase/auth"
	promptUseCase "github.com/yoshioka0101/voiceblog/backend/internal/usecase/prompt"
	promptRunJobUseCase "github.com/yoshioka0101/voiceblog/backend/internal/usecase/promptrunjob"
	transcriptionUseCase "github.com/yoshioka0101/voiceblog/backend/internal/usecase/transcription"
	userUseCase "github.com/yoshioka0101/voiceblog/backend/internal/usecase/user"
)

type Repositories struct {
	User          domainUser.Repository
	Prompt        promptEntity.Repository
	Transcription transcriptionEntity.Repository
	Article       articleEntity.Repository
	PromptRunJob  promptRunJobEntity.Repository
}

type UseCases struct {
	Auth          *authUseCase.UseCase
	User          *userUseCase.UseCase
	Prompt        *promptUseCase.UseCase
	Transcription *transcriptionUseCase.UseCase
	Article       *articleUseCase.UseCase
	PromptRunJob  *promptRunJobUseCase.UseCase
}

type Container struct {
	Repositories *Repositories
	UseCases     *UseCases
}

func New(db *sql.DB, googleClientID, geminiAPIKey string) *Container {
	userRepo := postgresql.NewUserRepository(db)
	verifier := googleauth.NewTokenVerifier(googleClientID)
	promptRepo := postgresql.NewPromptRepository(db)
	transcriptionRepo := postgresql.NewTranscriptionRepository(db)
	articleRepo := postgresql.NewArticleRepository(db)
	promptRunJobRepo := postgresql.NewPromptRunJobRepository(db)
	geminiClient := gemini.NewClient(geminiAPIKey)

	repos := &Repositories{
		User:          userRepo,
		Prompt:        promptRepo,
		Transcription: transcriptionRepo,
		Article:       articleRepo,
		PromptRunJob:  promptRunJobRepo,
	}
	useCases := &UseCases{
		Auth:          authUseCase.NewUseCase(verifier, repos.User),
		User:          userUseCase.NewUseCase(repos.User),
		Prompt:        promptUseCase.NewUseCase(repos.Prompt),
		Transcription: transcriptionUseCase.NewUseCase(repos.Transcription),
		Article:       articleUseCase.NewUseCase(repos.Article, repos.PromptRunJob, repos.Transcription),
		PromptRunJob:  promptRunJobUseCase.NewUseCase(repos.PromptRunJob, repos.Transcription, repos.Prompt, geminiClient),
	}

	return &Container{
		Repositories: repos,
		UseCases:     useCases,
	}
}
