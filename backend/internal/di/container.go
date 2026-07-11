package di

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/yoshioka0101/voiceblog/backend/internal/entity/repository"

	dbtx "github.com/yoshioka0101/voiceblog/backend/internal/db"
	"github.com/yoshioka0101/voiceblog/backend/internal/infra/crypto"
	"github.com/yoshioka0101/voiceblog/backend/internal/infra/gemini"
	"github.com/yoshioka0101/voiceblog/backend/internal/infra/googleauth"
	"github.com/yoshioka0101/voiceblog/backend/internal/infra/hatena"
	"github.com/yoshioka0101/voiceblog/backend/internal/infra/qiita"
	"github.com/yoshioka0101/voiceblog/backend/internal/repository/postgresql"
	articleUseCase "github.com/yoshioka0101/voiceblog/backend/internal/usecase/article"
	authUseCase "github.com/yoshioka0101/voiceblog/backend/internal/usecase/auth"
	integrationUseCase "github.com/yoshioka0101/voiceblog/backend/internal/usecase/integration"
	promptUseCase "github.com/yoshioka0101/voiceblog/backend/internal/usecase/prompt"
	promptRunJobUseCase "github.com/yoshioka0101/voiceblog/backend/internal/usecase/promptrunjob"
	publishUseCase "github.com/yoshioka0101/voiceblog/backend/internal/usecase/publish"
	transcriptionUseCase "github.com/yoshioka0101/voiceblog/backend/internal/usecase/transcription"
	userUseCase "github.com/yoshioka0101/voiceblog/backend/internal/usecase/user"
)

type Repositories struct {
	User               repository.UserRepository
	Prompt             repository.PromptRepository
	Transcription      repository.TranscriptionRepository
	Article            repository.ArticleRepository
	PromptRunJob       repository.PromptRunJobRepository
	ExternalToken      repository.ExternalTokenRepository
	ArticleShareTarget repository.ArticleShareTargetRepository
}

type UseCases struct {
	Auth          *authUseCase.UseCase
	User          *userUseCase.UseCase
	Prompt        *promptUseCase.UseCase
	Transcription *transcriptionUseCase.UseCase
	Article       *articleUseCase.UseCase
	PromptRunJob  *promptRunJobUseCase.UseCase
	Integration   *integrationUseCase.UseCase
	Publish       *publishUseCase.UseCase
}

type Container struct {
	Repositories *Repositories
	UseCases     *UseCases
}

func New(db *sql.DB, googleClientID, geminiAPIKey, geminiModel, tokenEncryptionKey string) (*Container, error) {
	userRepo := postgresql.NewUserRepository(db)
	verifier := googleauth.NewTokenVerifier(googleClientID)
	promptRepo := postgresql.NewPromptRepository(db)
	transcriptionRepo := postgresql.NewTranscriptionRepository(db)
	articleRepo := postgresql.NewArticleRepository(db)
	promptRunJobRepo := postgresql.NewPromptRunJobRepository(db)
	externalTokenRepo := postgresql.NewExternalTokenRepository(db)
	articleShareTargetRepo := postgresql.NewArticleShareTargetRepository(db)
	geminiClient := gemini.NewClient(geminiAPIKey, geminiModel)
	txRunner := dbtx.NewTxRunner(db)

	repos := &Repositories{
		User:               userRepo,
		Prompt:             promptRepo,
		Transcription:      transcriptionRepo,
		Article:            articleRepo,
		PromptRunJob:       promptRunJobRepo,
		ExternalToken:      externalTokenRepo,
		ArticleShareTarget: articleShareTargetRepo,
	}

	useCases := &UseCases{
		Auth:          authUseCase.NewUseCase(verifier, repos.User),
		User:          userUseCase.NewUseCase(repos.User),
		Prompt:        promptUseCase.NewUseCase(repos.Prompt, txRunner),
		Transcription: transcriptionUseCase.NewUseCase(repos.Transcription),
		Article:       articleUseCase.NewUseCase(repos.Article, repos.PromptRunJob, repos.Transcription, txRunner).WithGenerator(repos.Prompt, geminiClient),
		PromptRunJob:  promptRunJobUseCase.NewUseCase(repos.PromptRunJob, repos.Transcription, repos.Prompt, geminiClient, txRunner),
	}

	if tokenEncryptionKey != "" {
		encryptor, err := crypto.NewAESEncryptor(tokenEncryptionKey)
		if err != nil {
			return nil, fmt.Errorf("create token encryptor: %w", err)
		}

		qiitaClient := qiita.NewClient()
		hatenaClient := hatena.NewClient()

		verifiers := map[string]integrationUseCase.Verifier{
			"qiita":  &qiitaVerifier{client: qiitaClient},
			"hatena": &hatenaVerifier{client: hatenaClient},
		}
		integrationUC := integrationUseCase.NewUseCase(repos.ExternalToken, encryptor, verifiers)
		useCases.Integration = integrationUC

		publishers := map[string]publishUseCase.Publisher{
			"qiita":  &qiitaAdapter{client: qiitaClient},
			"hatena": &hatenaAdapter{client: hatenaClient},
		}
		useCases.Publish = publishUseCase.NewUseCase(repos.Article, repos.ArticleShareTarget, integrationUC, publishers)
	}

	return &Container{
		Repositories: repos,
		UseCases:     useCases,
	}, nil
}

type qiitaVerifier struct {
	client *qiita.Client
}

func (v *qiitaVerifier) Verify(ctx context.Context, token string) error {
	return v.client.Verify(ctx, token)
}

type hatenaVerifier struct {
	client *hatena.Client
}

func (v *hatenaVerifier) Verify(ctx context.Context, token string) error {
	return v.client.Verify(ctx, token)
}

type qiitaAdapter struct {
	client *qiita.Client
}

func (a *qiitaAdapter) Publish(ctx context.Context, token, title, content string) (*publishUseCase.PublishResult, error) {
	r, err := a.client.Publish(ctx, token, title, content)
	if err != nil {
		return nil, err
	}
	return &publishUseCase.PublishResult{ID: r.ID, URL: r.URL}, nil
}

type hatenaAdapter struct {
	client *hatena.Client
}

func (a *hatenaAdapter) Publish(ctx context.Context, token, title, content string) (*publishUseCase.PublishResult, error) {
	r, err := a.client.Publish(ctx, token, title, content)
	if err != nil {
		return nil, err
	}
	return &publishUseCase.PublishResult{ID: r.ID, URL: r.URL}, nil
}
