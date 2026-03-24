package entity

import "context"

type Repository interface {
	CreateArticle(ctx context.Context, value *Article) (*Article, error)
	FindArticleByID(ctx context.Context, id int64) (*Article, error)
	FindArticleByPromptRunJobID(ctx context.Context, promptRunJobID int64) (*Article, error)
	ListArticlesByUserID(ctx context.Context, userID int64) ([]*Article, error)
	UpdateArticle(ctx context.Context, value *Article) (*Article, error)
	DeleteArticle(ctx context.Context, id int64) error
	UpsertArticleByPromptRunJobID(ctx context.Context, value *Article) (*Article, error)
}
