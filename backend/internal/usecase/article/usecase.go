package usecase

import (
	"context"
	"strings"

	"github.com/yoshioka0101/voiceblog/backend/internal/apperr"
	entity "github.com/yoshioka0101/voiceblog/backend/internal/entity/article"
	promptrunjobEntity "github.com/yoshioka0101/voiceblog/backend/internal/entity/promptrunjob"
	transcriptionEntity "github.com/yoshioka0101/voiceblog/backend/internal/entity/transcription"
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

type UseCase struct {
	repo              entity.Repository
	promptRunJobRepo  promptrunjobEntity.Repository
	transcriptionRepo transcriptionEntity.Repository
}

func NewUseCase(
	repo entity.Repository,
	promptRunJobRepo promptrunjobEntity.Repository,
	transcriptionRepo transcriptionEntity.Repository,
) *UseCase {
	return &UseCase{
		repo:              repo,
		promptRunJobRepo:  promptRunJobRepo,
		transcriptionRepo: transcriptionRepo,
	}
}

func (uc *UseCase) CreateArticle(ctx context.Context, input CreateArticleInput) (*entity.Article, error) {
	title := strings.TrimSpace(input.Title)
	content := strings.TrimSpace(input.Content)
	if title == "" || content == "" {
		return nil, apperr.BadRequest( "title and content are required")
	}

	article := &entity.Article{
		UserID:  input.UserID,
		Title:   title,
		Content: content,
	}

	if input.PromptRunJobID == nil {
		return uc.repo.CreateArticle(ctx, article)
	}

	job, err := uc.promptRunJobRepo.FindPromptRunJobByID(ctx, *input.PromptRunJobID)
	if err != nil {
		return nil, err
	}

	transcriptionValue, err := uc.transcriptionRepo.FindByID(ctx, job.TranscriptionID)
	if err != nil {
		return nil, err
	}
	if transcriptionValue.UserID != input.UserID {
		return nil, ErrForbidden
	}

	article.PromptRunJobID = input.PromptRunJobID

	return uc.repo.UpsertArticleByPromptRunJobID(ctx, article)
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
		return nil, apperr.BadRequest( "at least one field is required")
	}

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
			return nil, apperr.BadRequest( "title must not be blank")
		}
		value.Title = title
	}

	if input.Content != nil {
		content := strings.TrimSpace(*input.Content)
		if content == "" {
			return nil, apperr.BadRequest( "content must not be blank")
		}
		value.Content = content
	}

	return uc.repo.UpdateArticle(ctx, value)
}

func (uc *UseCase) DeleteArticle(ctx context.Context, userID, articleID int64) error {
	value, err := uc.repo.FindArticleByID(ctx, articleID)
	if err != nil {
		return err
	}
	if value.UserID != userID {
		return ErrForbidden
	}

	return uc.repo.DeleteArticle(ctx, articleID)
}
