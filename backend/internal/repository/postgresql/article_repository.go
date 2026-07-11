package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/aarondl/opt/omit"
	"github.com/aarondl/opt/omitnull"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/dialect"
	"github.com/stephenafamo/bob/dialect/psql/sm"

	dbctx "github.com/yoshioka0101/voiceblog/backend/internal/db"
	entity "github.com/yoshioka0101/voiceblog/backend/internal/entity/article"
	articleusecase "github.com/yoshioka0101/voiceblog/backend/internal/usecase/article"
	"github.com/yoshioka0101/voiceblog/backend/models"
)

type ArticleRepository struct {
	db bob.Executor
}

func NewArticleRepository(db *sql.DB) *ArticleRepository {
	return &ArticleRepository{db: bob.NewDB(db)}
}

func NewArticleRepositoryWithExecutor(exec bob.Executor) *ArticleRepository {
	return &ArticleRepository{db: exec}
}

func (r *ArticleRepository) CreateArticle(ctx context.Context, value *entity.Article) (*entity.Article, error) {
	exec := dbctx.ExecutorFromContext(ctx, r.db)
	modelValue, err := models.Articles.Insert(&models.ArticleSetter{
		UserID:  omit.From(value.UserID),
		Title:   omit.From(value.Title),
		Content: omit.From(value.Content),
	}).One(ctx, exec)
	if err != nil {
		return nil, fmt.Errorf("insert article: %w", err)
	}

	return toArticleEntity(modelValue), nil
}

func (r *ArticleRepository) FindArticleByID(ctx context.Context, id int64) (*entity.Article, error) {
	exec := dbctx.ExecutorFromContext(ctx, r.db)
	mods := []bob.Mod[*dialect.SelectQuery]{
		sm.Where(
			models.Articles.Columns.ID.EQ(psql.Arg(id)).And(
				models.Articles.Columns.DeletedAt.IsNull(),
			),
		),
	}
	if dbctx.InTx(ctx) {
		mods = append(mods, sm.ForUpdate("articles"))
	}

	modelValue, err := models.Articles.Query(mods...).One(ctx, exec)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, articleusecase.ErrNotFound
		}
		return nil, fmt.Errorf("find article: %w", err)
	}

	return toArticleEntity(modelValue), nil
}

func (r *ArticleRepository) ListArticlesByUserID(ctx context.Context, userID int64) ([]*entity.Article, error) {
	exec := dbctx.ExecutorFromContext(ctx, r.db)
	modelValues, err := models.Articles.Query(
		sm.Where(
			models.Articles.Columns.UserID.EQ(psql.Arg(userID)).And(
				models.Articles.Columns.DeletedAt.IsNull(),
			),
		),
		sm.OrderBy(models.Articles.Columns.UpdatedAt).Desc(),
		sm.OrderBy(models.Articles.Columns.ID).Desc(),
	).All(ctx, exec)
	if err != nil {
		return nil, fmt.Errorf("list articles: %w", err)
	}

	result := make([]*entity.Article, 0, len(modelValues))
	for _, modelValue := range modelValues {
		result = append(result, toArticleEntity(modelValue))
	}

	return result, nil
}

func (r *ArticleRepository) UpdateArticle(ctx context.Context, value *entity.Article) (*entity.Article, error) {
	exec := dbctx.ExecutorFromContext(ctx, r.db)
	modelValue := &models.Article{ID: value.ID}
	if err := modelValue.Update(ctx, exec, &models.ArticleSetter{
		Title:     omit.From(value.Title),
		Content:   omit.From(value.Content),
		UpdatedAt: omit.From(time.Now()),
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, articleusecase.ErrNotFound
		}
		return nil, fmt.Errorf("update article: %w", err)
	}

	return toArticleEntity(modelValue), nil
}

func (r *ArticleRepository) DeleteArticle(ctx context.Context, id int64) error {
	exec := dbctx.ExecutorFromContext(ctx, r.db)
	exists, err := models.Articles.Query(
		sm.Where(
			models.Articles.Columns.ID.EQ(psql.Arg(id)).And(
				models.Articles.Columns.DeletedAt.IsNull(),
			),
		),
	).Exists(ctx, exec)
	if err != nil {
		return fmt.Errorf("check article exists: %w", err)
	}
	if !exists {
		return articleusecase.ErrNotFound
	}

	modelValue := &models.Article{ID: id}
	now := time.Now()
	if err := modelValue.Update(ctx, exec, &models.ArticleSetter{
		DeletedAt: omitnull.From(now),
		UpdatedAt: omit.From(now),
	}); err != nil {
		return fmt.Errorf("soft delete article: %w", err)
	}

	return nil
}

func toArticleEntity(value *models.Article) *entity.Article {
	if value == nil {
		return nil
	}

	return &entity.Article{
		ID:        value.ID,
		UserID:    value.UserID,
		Title:     value.Title,
		Content:   value.Content,
		DeletedAt: value.DeletedAt.Ptr(),
		CreatedAt: value.CreatedAt,
		UpdatedAt: value.UpdatedAt,
	}
}
