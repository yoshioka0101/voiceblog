package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/aarondl/opt/omit"
	"github.com/aarondl/opt/omitnull"
	"github.com/stephenafamo/bob"

	entity "github.com/yoshioka0101/voiceblog/backend/internal/entity/promptrunjob"
	promptrunjobusecase "github.com/yoshioka0101/voiceblog/backend/internal/usecase/promptrunjob"
	"github.com/yoshioka0101/voiceblog/backend/models"
)

type PromptRunJobRepository struct {
	db bob.DB
}

func NewPromptRunJobRepository(db *sql.DB) *PromptRunJobRepository {
	return &PromptRunJobRepository{db: bob.NewDB(db)}
}

func (r *PromptRunJobRepository) CreatePromptRunJob(ctx context.Context, value *entity.Job) (*entity.Job, error) {
	modelValue, err := models.PromptRunJobs.Insert(&models.PromptRunJobSetter{
		TranscriptionID:  omit.From(value.TranscriptionID),
		PromptID:         omit.From(value.PromptID),
		Status:           omit.From(value.Status),
		AttemptCount:     omit.From(int32(value.AttemptCount)),
		NextRunAt:        omit.From(value.NextRunAt),
		ErrorMessage:     omitnull.FromPtr(value.ErrorMessage),
		GeneratedTitle:   omitnull.FromPtr(value.GeneratedTitle),
		GeneratedContent: omitnull.FromPtr(value.GeneratedContent),
	}).One(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("insert prompt run job: %w", err)
	}

	return toPromptRunJobEntity(modelValue), nil
}

func (r *PromptRunJobRepository) FindPromptRunJobByID(ctx context.Context, id int64) (*entity.Job, error) {
	modelValue, err := models.FindPromptRunJob(ctx, r.db, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, promptrunjobusecase.ErrNotFound
		}
		return nil, fmt.Errorf("find prompt run job: %w", err)
	}

	return toPromptRunJobEntity(modelValue), nil
}

func (r *PromptRunJobRepository) UpdatePromptRunJob(ctx context.Context, value *entity.Job) (*entity.Job, error) {
	modelValue := &models.PromptRunJob{ID: value.ID}
	if err := modelValue.Update(ctx, r.db, &models.PromptRunJobSetter{
		Status:           omit.From(value.Status),
		AttemptCount:     omit.From(int32(value.AttemptCount)),
		NextRunAt:        omit.From(value.NextRunAt),
		ErrorMessage:     omitnull.FromPtr(value.ErrorMessage),
		GeneratedTitle:   omitnull.FromPtr(value.GeneratedTitle),
		GeneratedContent: omitnull.FromPtr(value.GeneratedContent),
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, promptrunjobusecase.ErrNotFound
		}
		return nil, fmt.Errorf("update prompt run job: %w", err)
	}

	return toPromptRunJobEntity(modelValue), nil
}

func toPromptRunJobEntity(value *models.PromptRunJob) *entity.Job {
	if value == nil {
		return nil
	}

	return &entity.Job{
		ID:               value.ID,
		TranscriptionID:  value.TranscriptionID,
		PromptID:         value.PromptID,
		Status:           value.Status,
		AttemptCount:     int(value.AttemptCount),
		NextRunAt:        value.NextRunAt,
		ErrorMessage:     value.ErrorMessage.Ptr(),
		GeneratedTitle:   value.GeneratedTitle.Ptr(),
		GeneratedContent: value.GeneratedContent.Ptr(),
		CreatedAt:        value.CreatedAt,
	}
}
