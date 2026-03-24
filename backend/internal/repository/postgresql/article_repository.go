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
	"github.com/stephenafamo/bob/dialect/psql/im"
	"github.com/stephenafamo/bob/dialect/psql/sm"

	entity "github.com/yoshioka0101/voiceblog/backend/internal/entity/article"
	articleusecase "github.com/yoshioka0101/voiceblog/backend/internal/usecase/article"
	"github.com/yoshioka0101/voiceblog/backend/models"
)

type ArticleRepository struct {
	db bob.DB
}

func NewArticleRepository(db *sql.DB) *ArticleRepository {
	return &ArticleRepository{db: bob.NewDB(db)}
}

func (r *ArticleRepository) CreateArticle(ctx context.Context, value *entity.Article) (*entity.Article, error) {
	modelValue, err := models.Articles.Insert(&models.ArticleSetter{
		UserID:         omit.From(value.UserID),
		PromptRunJobID: omitnull.FromPtr(value.PromptRunJobID),
		Title:          omit.From(value.Title),
		Content:        omit.From(value.Content),
	}).One(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("insert article: %w", err)
	}

	return toArticleEntity(modelValue), nil
}

func (r *ArticleRepository) FindArticleByID(ctx context.Context, id int64) (*entity.Article, error) {
	modelValue, err := models.Articles.Query(
		sm.Where(
			models.Articles.Columns.ID.EQ(psql.Arg(id)).And(
				models.Articles.Columns.DeletedAt.IsNull(),
			),
		),
	).One(ctx, r.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, articleusecase.ErrNotFound
		}
		return nil, fmt.Errorf("find article: %w", err)
	}

	return toArticleEntity(modelValue), nil
}

func (r *ArticleRepository) FindArticleByPromptRunJobID(ctx context.Context, promptRunJobID int64) (*entity.Article, error) {
	modelValue, err := models.Articles.Query(
		sm.Where(
			models.Articles.Columns.PromptRunJobID.EQ(psql.Arg(promptRunJobID)).And(
				models.Articles.Columns.DeletedAt.IsNull(),
			),
		),
	).One(ctx, r.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find article by prompt run job id: %w", err)
	}

	return toArticleEntity(modelValue), nil
}

func (r *ArticleRepository) ListArticlesByUserID(ctx context.Context, userID int64) ([]*entity.Article, error) {
	modelValues, err := models.Articles.Query(
		sm.Where(
			models.Articles.Columns.UserID.EQ(psql.Arg(userID)).And(
				models.Articles.Columns.DeletedAt.IsNull(),
			),
		),
		sm.OrderBy(models.Articles.Columns.UpdatedAt).Desc(),
		sm.OrderBy(models.Articles.Columns.ID).Desc(),
	).All(ctx, r.db)
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
	modelValue := &models.Article{ID: value.ID}
	if err := modelValue.Update(ctx, r.db, &models.ArticleSetter{
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
	exists, err := models.Articles.Query(
		sm.Where(
			models.Articles.Columns.ID.EQ(psql.Arg(id)).And(
				models.Articles.Columns.DeletedAt.IsNull(),
			),
		),
	).Exists(ctx, r.db)
	if err != nil {
		return fmt.Errorf("check article exists: %w", err)
	}
	if !exists {
		return articleusecase.ErrNotFound
	}

	modelValue := &models.Article{ID: id}
	now := time.Now()
	if err := modelValue.Update(ctx, r.db, &models.ArticleSetter{
		DeletedAt: omitnull.From(now),
		UpdatedAt: omit.From(now),
	}); err != nil {
		return fmt.Errorf("soft delete article: %w", err)
	}

	return nil
}

func (r *ArticleRepository) UpsertArticleByPromptRunJobID(ctx context.Context, value *entity.Article) (*entity.Article, error) {
	if value.PromptRunJobID == nil {
		return r.CreateArticle(ctx, value)
	}

	now := time.Now()
	var deletedAt *time.Time
	modelValue, err := models.Articles.Insert(
		&models.ArticleSetter{
			UserID:         omit.From(value.UserID),
			PromptRunJobID: omitnull.FromPtr(value.PromptRunJobID),
			Title:          omit.From(value.Title),
			Content:        omit.From(value.Content),
			DeletedAt:      omitnull.FromPtr(deletedAt),
			UpdatedAt:      omit.From(now),
		},
		im.OnConflict("prompt_run_job_id").DoUpdate(
			im.SetExcluded("title", "content", "updated_at"),
			im.SetCol("deleted_at").To(psql.Raw("NULL")),
		),
	).One(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("upsert article: %w", err)
	}

	return toArticleEntity(modelValue), nil
}

func toArticleEntity(value *models.Article) *entity.Article {
	if value == nil {
		return nil
	}

	return &entity.Article{
		ID:             value.ID,
		UserID:         value.UserID,
		PromptRunJobID: value.PromptRunJobID.Ptr(),
		Title:          value.Title,
		Content:        value.Content,
		DeletedAt:      value.DeletedAt.Ptr(),
		CreatedAt:      value.CreatedAt,
		UpdatedAt:      value.UpdatedAt,
	}
}
