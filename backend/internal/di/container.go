package di

import (
	"context"
	"database/sql"

	articleEntity "github.com/yoshioka0101/voiceblog/backend/internal/entity/article"
	shareEntity "github.com/yoshioka0101/voiceblog/backend/internal/entity/articlesharetarget"
	tokenEntity "github.com/yoshioka0101/voiceblog/backend/internal/entity/externaltoken"
	promptEntity "github.com/yoshioka0101/voiceblog/backend/internal/entity/prompt"
	promptRunJobEntity "github.com/yoshioka0101/voiceblog/backend/internal/entity/promptrunjob"
	transcriptionEntity "github.com/yoshioka0101/voiceblog/backend/internal/entity/transcription"
	domainUser "github.com/yoshioka0101/voiceblog/backend/internal/entity/user"
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
	User               domainUser.Repository
	Prompt             promptEntity.Repository
	Transcription      transcriptionEntity.Repository
	Article            articleEntity.Repository
	PromptRunJob       promptRunJobEntity.Repository
	ExternalToken      tokenEntity.Repository
	ArticleShareTarget shareEntity.Repository
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

func New(db *sql.DB, googleClientID, geminiAPIKey, tokenEncryptionKey string) *Container {
	userRepo := postgresql.NewUserRepository(db)
	verifier := googleauth.NewTokenVerifier(googleClientID)
	promptRepo := postgresql.NewPromptRepository(db)
	transcriptionRepo := postgresql.NewTranscriptionRepository(db)
	articleRepo := postgresql.NewArticleRepository(db)
	promptRunJobRepo := postgresql.NewPromptRunJobRepository(db)
	externalTokenRepo := postgresql.NewExternalTokenRepository(db)
	articleShareTargetRepo := postgresql.NewArticleShareTargetRepository(db)
	geminiClient := gemini.NewClient(geminiAPIKey)

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
		Prompt:        promptUseCase.NewUseCase(repos.Prompt),
		Transcription: transcriptionUseCase.NewUseCase(repos.Transcription),
		Article:       articleUseCase.NewUseCase(repos.Article, repos.PromptRunJob, repos.Transcription),
		PromptRunJob:  promptRunJobUseCase.NewUseCase(repos.PromptRunJob, repos.Transcription, repos.Prompt, geminiClient),
	}

	if tokenEncryptionKey != "" {
		encryptor, err := crypto.NewAESEncryptor(tokenEncryptionKey)
		if err == nil {
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
	}

	return &Container{
		Repositories: repos,
		UseCases:     useCases,
	}
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
