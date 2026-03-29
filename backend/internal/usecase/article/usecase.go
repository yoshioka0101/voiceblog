package usecase

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/yoshioka0101/voiceblog/backend/internal/apperr"
	dbtx "github.com/yoshioka0101/voiceblog/backend/internal/db"
	entity "github.com/yoshioka0101/voiceblog/backend/internal/entity/article"
	promptEntity "github.com/yoshioka0101/voiceblog/backend/internal/entity/prompt"
	promptrunjobEntity "github.com/yoshioka0101/voiceblog/backend/internal/entity/promptrunjob"
	transcriptionEntity "github.com/yoshioka0101/voiceblog/backend/internal/entity/transcription"
	"github.com/yoshioka0101/voiceblog/backend/internal/usecase/articlegen"
)

var (
	ErrNotFound  = apperr.NotFound("article")
	ErrForbidden = apperr.ErrForbidden
)

type CreateArticleInput struct {
	UserID         int64
	PromptRunJobID *int64
	Title          string
	Content        string
}

type UpdateArticleInput struct {
	UserID    int64
	ArticleID int64
	Title     *string
	Content   *string
}

type GenerateArticleInput struct {
	UserID          int64
	TranscriptionID int64
	PromptID        int64
}

type UseCase struct {
	repo              entity.Repository
	promptRunJobRepo  promptrunjobEntity.Repository
	transcriptionRepo transcriptionEntity.Repository
	promptRepo        promptEntity.Repository
	generator         articlegen.Generator
	txRunner          dbtx.TxRunner
}

func NewUseCase(
	repo entity.Repository,
	promptRunJobRepo promptrunjobEntity.Repository,
	transcriptionRepo transcriptionEntity.Repository,
	txRunners ...dbtx.TxRunner,
) *UseCase {
	var txRunner dbtx.TxRunner
	if len(txRunners) > 0 {
		txRunner = txRunners[0]
	}

	return &UseCase{
		repo:              repo,
		promptRunJobRepo:  promptRunJobRepo,
		transcriptionRepo: transcriptionRepo,
		txRunner:          txRunner,
	}
}

func (uc *UseCase) WithGenerator(promptRepo promptEntity.Repository, generator articlegen.Generator) *UseCase {
	uc.promptRepo = promptRepo
	uc.generator = generator
	return uc
}

func (uc *UseCase) CreateArticle(ctx context.Context, input CreateArticleInput) (*entity.Article, error) {
	title := strings.TrimSpace(input.Title)
	content := strings.TrimSpace(input.Content)
	if title == "" || content == "" {
		return nil, apperr.BadRequest("title and content are required")
	}

	article := &entity.Article{
		UserID:  input.UserID,
		Title:   title,
		Content: content,
	}

	if input.PromptRunJobID == nil {
		return uc.repo.CreateArticle(ctx, article)
	}

	if uc.txRunner == nil {
		return uc.createArticleFromPromptRunJob(ctx, article, input.UserID, *input.PromptRunJobID)
	}

	var created *entity.Article
	err := uc.txRunner.RunInTx(ctx, func(txCtx context.Context) error {
		var err error
		created, err = uc.createArticleFromPromptRunJob(txCtx, article, input.UserID, *input.PromptRunJobID)
		return err
	})
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (uc *UseCase) createArticleFromPromptRunJob(ctx context.Context, article *entity.Article, userID int64, promptRunJobID int64) (*entity.Article, error) {
	job, err := uc.promptRunJobRepo.FindPromptRunJobByID(ctx, promptRunJobID)
	if err != nil {
		return nil, err
	}

	transcriptionValue, err := uc.transcriptionRepo.FindTranscriptionByID(ctx, job.TranscriptionID)
	if err != nil {
		return nil, err
	}
	if transcriptionValue.UserID != userID {
		return nil, ErrForbidden
	}

	article.PromptRunJobID = &promptRunJobID

	return uc.repo.UpsertArticleByPromptRunJobID(ctx, article)
}

func (uc *UseCase) GenerateArticle(ctx context.Context, input GenerateArticleInput) (*entity.Article, error) {
	if uc.promptRepo == nil || uc.generator == nil {
		return nil, apperr.InternalError("article generator is not configured")
	}

	transcriptionValue, err := uc.transcriptionRepo.FindTranscriptionByID(ctx, input.TranscriptionID)
	if err != nil {
		return nil, err
	}
	if transcriptionValue.UserID != input.UserID {
		return nil, ErrForbidden
	}

	promptValue, err := uc.promptRepo.FindPromptByID(ctx, input.PromptID)
	if err != nil {
		return nil, err
	}
	if !visiblePromptToUser(promptValue, input.UserID) {
		return nil, ErrForbidden
	}

	generated, err := uc.generator.GenerateArticle(ctx, articlegen.Input{
		PromptName:   promptValue.Name,
		PromptBody:   promptValue.Body,
		FullText:     transcriptionValue.FullText,
		SegmentsJSON: transcriptionValue.SegmentsJSON,
	})
	if err != nil {
		return nil, mapGenerateArticleError(err)
	}

	title, content, err := normalizeGeneratedArticle(generated)
	if err != nil {
		return nil, err
	}

	return uc.repo.CreateArticle(ctx, &entity.Article{
		UserID:  input.UserID,
		Title:   title,
		Content: content,
	})
}

func (uc *UseCase) ListArticlesByUserID(ctx context.Context, userID int64) ([]*entity.Article, error) {
	return uc.repo.ListArticlesByUserID(ctx, userID)
}

func (uc *UseCase) GetArticle(ctx context.Context, userID, articleID int64) (*entity.Article, error) {
	value, err := uc.repo.FindArticleByID(ctx, articleID)
	if err != nil {
		return nil, err
	}
	if value.UserID != userID {
		return nil, ErrForbidden
	}

	return value, nil
}

func (uc *UseCase) UpdateArticle(ctx context.Context, input UpdateArticleInput) (*entity.Article, error) {
	if input.Title == nil && input.Content == nil {
		return nil, apperr.BadRequest("at least one field is required")
	}

	if uc.txRunner == nil {
		return uc.updateArticle(ctx, input)
	}

	var article *entity.Article
	err := uc.txRunner.RunInTx(ctx, func(txCtx context.Context) error {
		var err error
		article, err = uc.updateArticle(txCtx, input)
		return err
	})
	if err != nil {
		return nil, err
	}

	return article, nil
}

func (uc *UseCase) updateArticle(ctx context.Context, input UpdateArticleInput) (*entity.Article, error) {
	value, err := uc.repo.FindArticleByID(ctx, input.ArticleID)
	if err != nil {
		return nil, err
	}
	if value.UserID != input.UserID {
		return nil, ErrForbidden
	}

	if input.Title != nil {
		title := strings.TrimSpace(*input.Title)
		if title == "" {
			return nil, apperr.BadRequest("title must not be blank")
		}
		value.Title = title
	}

	if input.Content != nil {
		content := strings.TrimSpace(*input.Content)
		if content == "" {
			return nil, apperr.BadRequest("content must not be blank")
		}
		value.Content = content
	}

	return uc.repo.UpdateArticle(ctx, value)
}

func (uc *UseCase) DeleteArticle(ctx context.Context, userID, articleID int64) error {
	if uc.txRunner == nil {
		return uc.deleteArticle(ctx, userID, articleID)
	}

	return uc.txRunner.RunInTx(ctx, func(txCtx context.Context) error {
		return uc.deleteArticle(txCtx, userID, articleID)
	})
}

func (uc *UseCase) deleteArticle(ctx context.Context, userID, articleID int64) error {
	value, err := uc.repo.FindArticleByID(ctx, articleID)
	if err != nil {
		return err
	}
	if value.UserID != userID {
		return ErrForbidden
	}

	return uc.repo.DeleteArticle(ctx, articleID)
}

func normalizeGeneratedArticle(generated *articlegen.GeneratedArticle) (string, string, error) {
	if generated == nil {
		return "", "", apperr.InternalError("generated article is empty")
	}

	title := strings.TrimSpace(generated.Title)
	content := strings.TrimSpace(generated.Content)
	if title == "" || content == "" {
		return "", "", apperr.InternalError("generated article is empty")
	}

	return title, content, nil
}

func visiblePromptToUser(value *promptEntity.Prompt, userID int64) bool {
	if !value.IsActive {
		return false
	}
	if value.UserID == nil {
		return true
	}
	return *value.UserID == userID
}

func mapGenerateArticleError(err error) error {
	var rateLimitErr *articlegen.RateLimitError
	if errors.As(err, &rateLimitErr) {
		return apperr.NewWithCode(http.StatusTooManyRequests, "rate_limited", rateLimitErr.Error())
	}

	return err
}
