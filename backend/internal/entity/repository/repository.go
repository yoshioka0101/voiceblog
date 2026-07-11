// Package repository はドメインの永続化契約を一箇所に集約する。
// 実装は internal/repository/postgresql が提供し、di で注入される。
package repository

import (
	"context"

	article "github.com/yoshioka0101/voiceblog/backend/internal/entity/article"
	share "github.com/yoshioka0101/voiceblog/backend/internal/entity/articlesharetarget"
	externaltoken "github.com/yoshioka0101/voiceblog/backend/internal/entity/externaltoken"
	prompt "github.com/yoshioka0101/voiceblog/backend/internal/entity/prompt"
	promptrunjob "github.com/yoshioka0101/voiceblog/backend/internal/entity/promptrunjob"
	transcription "github.com/yoshioka0101/voiceblog/backend/internal/entity/transcription"
	"github.com/yoshioka0101/voiceblog/backend/internal/entity/user"
)

type UserRepository interface {
	FindOrCreate(ctx context.Context, provider, subject, email, name string) (*user.User, error)
}

type PromptRepository interface {
	ListVisiblePromptsByUserID(ctx context.Context, userID int64) ([]*prompt.Prompt, error)
	CreatePrompt(ctx context.Context, value *prompt.Prompt) (*prompt.Prompt, error)
	FindPromptByID(ctx context.Context, id int64) (*prompt.Prompt, error)
	UpdatePrompt(ctx context.Context, value *prompt.Prompt) (*prompt.Prompt, error)
	DeletePrompt(ctx context.Context, id int64) error
}

type TranscriptionRepository interface {
	CreateTranscription(ctx context.Context, value *transcription.Transcription) (*transcription.Transcription, error)
	FindTranscriptionByID(ctx context.Context, id int64) (*transcription.Transcription, error)
}

type ArticleRepository interface {
	CreateArticle(ctx context.Context, value *article.Article) (*article.Article, error)
	FindArticleByID(ctx context.Context, id int64) (*article.Article, error)
	FindArticleByPromptRunJobID(ctx context.Context, promptRunJobID int64) (*article.Article, error)
	ListArticlesByUserID(ctx context.Context, userID int64) ([]*article.Article, error)
	UpdateArticle(ctx context.Context, value *article.Article) (*article.Article, error)
	DeleteArticle(ctx context.Context, id int64) error
	UpsertArticleByPromptRunJobID(ctx context.Context, value *article.Article) (*article.Article, error)
}

type PromptRunJobRepository interface {
	CreatePromptRunJob(ctx context.Context, value *promptrunjob.Job) (*promptrunjob.Job, error)
	FindPromptRunJobByID(ctx context.Context, id int64) (*promptrunjob.Job, error)
	UpdatePromptRunJob(ctx context.Context, value *promptrunjob.Job) (*promptrunjob.Job, error)
}

type ExternalTokenRepository interface {
	StoreExternalToken(ctx context.Context, value *externaltoken.ExternalToken) (*externaltoken.ExternalToken, error)
	FindByUserIDAndProvider(ctx context.Context, userID int64, provider string) (*externaltoken.ExternalToken, error)
	DeleteByUserIDAndProvider(ctx context.Context, userID int64, provider string) error
	ListByUserID(ctx context.Context, userID int64) ([]*externaltoken.ExternalToken, error)
}

type ArticleShareTargetRepository interface {
	StoreArticleShareTarget(ctx context.Context, value *share.ArticleShareTarget) (*share.ArticleShareTarget, error)
	ListArticleShareTargetsByArticleID(ctx context.Context, articleID int64) ([]*share.ArticleShareTarget, error)
	FindArticleShareTargetByArticleIDAndProvider(ctx context.Context, articleID int64, provider string) (*share.ArticleShareTarget, error)
	DeleteArticleShareTargetByArticleIDAndProvider(ctx context.Context, articleID int64, provider string) error
}
