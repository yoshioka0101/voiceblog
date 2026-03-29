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
	"github.com/stephenafamo/bob/dialect/psql/dm"
	"github.com/stephenafamo/bob/dialect/psql/im"
	"github.com/stephenafamo/bob/dialect/psql/sm"

	dbctx "github.com/yoshioka0101/voiceblog/backend/internal/db"
	entity "github.com/yoshioka0101/voiceblog/backend/internal/entity/articlesharetarget"
	"github.com/yoshioka0101/voiceblog/backend/models"
)

type ArticleShareTargetRepository struct {
	db bob.Executor
}

func NewArticleShareTargetRepository(db *sql.DB) *ArticleShareTargetRepository {
	return &ArticleShareTargetRepository{db: bob.NewDB(db)}
}

func NewArticleShareTargetRepositoryWithExecutor(exec bob.Executor) *ArticleShareTargetRepository {
	return &ArticleShareTargetRepository{db: exec}
}

func (r *ArticleShareTargetRepository) StoreArticleShareTarget(ctx context.Context, value *entity.ArticleShareTarget) (*entity.ArticleShareTarget, error) {
	exec := dbctx.ExecutorFromContext(ctx, r.db)
	now := time.Now()
	modelValue, err := models.ArticleShareTargets.Insert(
		&models.ArticleShareTargetSetter{
			ArticleID:   omit.From(value.ArticleID),
			Provider:    omit.From(value.Provider),
			ExternalID:  omit.From(value.ExternalID),
			ExternalURL: omit.From(value.ExternalURL),
			PublishedAt: omitnull.FromPtr(value.PublishedAt),
			UpdatedAt:   omit.From(now),
		},
		im.OnConflict("article_id", "provider").DoUpdate(
			im.SetExcluded("external_id", "external_url", "published_at", "updated_at"),
		),
	).One(ctx, exec)
	if err != nil {
		return nil, fmt.Errorf("upsert article share target: %w", err)
	}

	return toArticleShareTargetEntity(modelValue), nil
}

func (r *ArticleShareTargetRepository) ListArticleShareTargetsByArticleID(ctx context.Context, articleID int64) ([]*entity.ArticleShareTarget, error) {
	exec := dbctx.ExecutorFromContext(ctx, r.db)
	modelValues, err := models.ArticleShareTargets.Query(
		sm.Where(models.ArticleShareTargets.Columns.ArticleID.EQ(psql.Arg(articleID))),
		sm.OrderBy(models.ArticleShareTargets.Columns.Provider).Asc(),
	).All(ctx, exec)
	if err != nil {
		return nil, fmt.Errorf("list article share targets: %w", err)
	}

	result := make([]*entity.ArticleShareTarget, 0, len(modelValues))
	for _, v := range modelValues {
		result = append(result, toArticleShareTargetEntity(v))
	}

	return result, nil
}

func (r *ArticleShareTargetRepository) FindArticleShareTargetByArticleIDAndProvider(ctx context.Context, articleID int64, provider string) (*entity.ArticleShareTarget, error) {
	exec := dbctx.ExecutorFromContext(ctx, r.db)
	modelValue, err := models.ArticleShareTargets.Query(
		sm.Where(
			models.ArticleShareTargets.Columns.ArticleID.EQ(psql.Arg(articleID)).And(
				models.ArticleShareTargets.Columns.Provider.EQ(psql.Arg(provider)),
			),
		),
	).One(ctx, exec)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find article share target: %w", err)
	}

	return toArticleShareTargetEntity(modelValue), nil
}

func (r *ArticleShareTargetRepository) DeleteArticleShareTargetByArticleIDAndProvider(ctx context.Context, articleID int64, provider string) error {
	exec := dbctx.ExecutorFromContext(ctx, r.db)
	_, err := models.ArticleShareTargets.Delete(
		dm.Where(
			models.ArticleShareTargets.Columns.ArticleID.EQ(psql.Arg(articleID)).And(
				models.ArticleShareTargets.Columns.Provider.EQ(psql.Arg(provider)),
			),
		),
	).Exec(ctx, exec)
	if err != nil {
		return fmt.Errorf("delete article share target: %w", err)
	}

	return nil
}

func toArticleShareTargetEntity(v *models.ArticleShareTarget) *entity.ArticleShareTarget {
	if v == nil {
		return nil
	}

	return &entity.ArticleShareTarget{
		ID:          v.ID,
		ArticleID:   v.ArticleID,
		Provider:    v.Provider,
		ExternalID:  v.ExternalID,
		ExternalURL: v.ExternalURL,
		PublishedAt: v.PublishedAt.Ptr(),
		CreatedAt:   v.CreatedAt,
		UpdatedAt:   v.UpdatedAt,
	}
}
