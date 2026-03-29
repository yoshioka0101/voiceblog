package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/yoshioka0101/voiceblog/backend/internal/apperr"
	articleEntity "github.com/yoshioka0101/voiceblog/backend/internal/entity/article"
	shareEntity "github.com/yoshioka0101/voiceblog/backend/internal/entity/articlesharetarget"
	integrationUsecase "github.com/yoshioka0101/voiceblog/backend/internal/usecase/integration"
)

type PublishResult struct {
	ID  string
	URL string
}

type Publisher interface {
	Publish(ctx context.Context, token, title, content string) (*PublishResult, error)
}

type UseCase struct {
	articleRepo   articleEntity.Repository
	shareRepo     shareEntity.Repository
	integrationUC *integrationUsecase.UseCase
	publishers    map[string]Publisher
}

func NewUseCase(
	articleRepo articleEntity.Repository,
	shareRepo shareEntity.Repository,
	integrationUC *integrationUsecase.UseCase,
	publishers map[string]Publisher,
) *UseCase {
	return &UseCase{
		articleRepo:   articleRepo,
		shareRepo:     shareRepo,
		integrationUC: integrationUC,
		publishers:    publishers,
	}
}

func (uc *UseCase) PublishArticle(ctx context.Context, userID, articleID int64, provider string) (*shareEntity.ArticleShareTarget, error) {
	article, err := uc.articleRepo.FindArticleByID(ctx, articleID)
	if err != nil {
		return nil, err
	}
	if article.UserID != userID {
		return nil, apperr.ErrForbidden
	}

	publisher, ok := uc.publishers[provider]
	if !ok {
		return nil, apperr.BadRequestWithCode("unsupported_provider", fmt.Sprintf("unsupported provider: %s", provider))
	}

	token, err := uc.integrationUC.DecryptToken(ctx, userID, provider)
	if err != nil {
		return nil, err
	}

	result, err := publisher.Publish(ctx, token, article.Title, article.Content)
	if err != nil {
		return nil, fmt.Errorf("publish to %s: %w", provider, err)
	}

	now := time.Now()
	target, err := uc.shareRepo.StoreArticleShareTarget(ctx, &shareEntity.ArticleShareTarget{
		ArticleID:   articleID,
		Provider:    provider,
		ExternalID:  result.ID,
		ExternalURL: result.URL,
		PublishedAt: &now,
	})
	if err != nil {
		return nil, fmt.Errorf("save share target: %w", err)
	}

	return target, nil
}

func (uc *UseCase) GetShareTargets(ctx context.Context, userID, articleID int64) ([]*shareEntity.ArticleShareTarget, error) {
	article, err := uc.articleRepo.FindArticleByID(ctx, articleID)
	if err != nil {
		return nil, err
	}
	if article.UserID != userID {
		return nil, apperr.ErrForbidden
	}

	return uc.shareRepo.ListArticleShareTargetsByArticleID(ctx, articleID)
}
