package entity

import "context"

type Repository interface {
	StoreArticleShareTarget(ctx context.Context, value *ArticleShareTarget) (*ArticleShareTarget, error)
	ListArticleShareTargetsByArticleID(ctx context.Context, articleID int64) ([]*ArticleShareTarget, error)
	FindArticleShareTargetByArticleIDAndProvider(ctx context.Context, articleID int64, provider string) (*ArticleShareTarget, error)
	DeleteArticleShareTargetByArticleIDAndProvider(ctx context.Context, articleID int64, provider string) error
}
