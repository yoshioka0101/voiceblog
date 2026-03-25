package entity

import "context"

type Repository interface {
	Upsert(ctx context.Context, value *ArticleShareTarget) (*ArticleShareTarget, error)
	ListByArticleID(ctx context.Context, articleID int64) ([]*ArticleShareTarget, error)
	FindByArticleIDAndProvider(ctx context.Context, articleID int64, provider string) (*ArticleShareTarget, error)
	DeleteByArticleIDAndProvider(ctx context.Context, articleID int64, provider string) error
}
