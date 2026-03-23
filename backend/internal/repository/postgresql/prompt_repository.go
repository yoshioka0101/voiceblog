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
	"github.com/stephenafamo/bob/dialect/psql/sm"

	entity "github.com/yoshioka0101/voiceblog/backend/internal/entity/prompt"
	promptusecase "github.com/yoshioka0101/voiceblog/backend/internal/usecase/prompt"
	"github.com/yoshioka0101/voiceblog/backend/models"
)

type PromptRepository struct {
	db bob.DB
}

func NewPromptRepository(db *sql.DB) *PromptRepository {
	return &PromptRepository{db: bob.NewDB(db)}
}

func (r *PromptRepository) ListVisibleByUserID(ctx context.Context, userID int64) ([]*entity.Prompt, error) {
	values, err := models.Prompts.Query(
		sm.Where(
			models.Prompts.Columns.IsActive.EQ(psql.Arg(true)).And(
				psql.Or(
					models.Prompts.Columns.UserID.IsNull(),
					models.Prompts.Columns.UserID.EQ(psql.Arg(userID)),
				),
			),
		),
		sm.OrderBy(models.Prompts.Columns.IsDefault).Desc(),
		sm.OrderBy(psql.Raw("CASE WHEN prompts.user_id IS NULL THEN 0 ELSE 1 END")).Asc(),
		sm.OrderBy(models.Prompts.Columns.CreatedAt).Desc(),
		sm.OrderBy(models.Prompts.Columns.ID).Desc(),
	).All(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("list prompts: %w", err)
	}

	result := make([]*entity.Prompt, 0, len(values))
	for _, value := range values {
		result = append(result, toPromptEntity(value))
	}

	return result, nil
}

func (r *PromptRepository) Create(ctx context.Context, promptEntity *entity.Prompt) (*entity.Prompt, error) {
	value, err := models.Prompts.Insert(&models.PromptSetter{
		UserID:    omitnull.FromPtr(promptEntity.UserID),
		Name:      omit.From(promptEntity.Name),
		Body:      omit.From(promptEntity.Body),
		IsActive:  omit.From(promptEntity.IsActive),
		IsDefault: omit.From(false),
	}).One(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("insert prompt: %w", err)
	}

	return toPromptEntity(value), nil
}

func (r *PromptRepository) FindByID(ctx context.Context, id int64) (*entity.Prompt, error) {
	value, err := models.FindPrompt(ctx, r.db, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, promptusecase.ErrNotFound
		}
		return nil, fmt.Errorf("find prompt: %w", err)
	}

	return toPromptEntity(value), nil
}

func (r *PromptRepository) Update(ctx context.Context, promptEntity *entity.Prompt) (*entity.Prompt, error) {
	value := &models.Prompt{ID: promptEntity.ID}
	if err := value.Update(ctx, r.db, &models.PromptSetter{
		Name:      omit.From(promptEntity.Name),
		Body:      omit.From(promptEntity.Body),
		IsActive:  omit.From(promptEntity.IsActive),
		UpdatedAt: omit.From(time.Now()),
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, promptusecase.ErrNotFound
		}
		return nil, fmt.Errorf("update prompt: %w", err)
	}

	return toPromptEntity(value), nil
}

func (r *PromptRepository) Delete(ctx context.Context, id int64) error {
	exists, err := models.PromptExists(ctx, r.db, id)
	if err != nil {
		return fmt.Errorf("check prompt exists: %w", err)
	}
	if !exists {
		return promptusecase.ErrNotFound
	}

	if err := (&models.Prompt{ID: id}).Delete(ctx, r.db); err != nil {
		return fmt.Errorf("delete prompt: %w", err)
	}

	return nil
}

func toPromptEntity(value *models.Prompt) *entity.Prompt {
	if value == nil {
		return nil
	}

	return &entity.Prompt{
		ID:        value.ID,
		UserID:    value.UserID.Ptr(),
		Name:      value.Name,
		Body:      value.Body,
		IsActive:  value.IsActive,
		IsDefault: value.IsDefault,
		CreatedAt: value.CreatedAt,
		UpdatedAt: value.UpdatedAt,
	}
}
