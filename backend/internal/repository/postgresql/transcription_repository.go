package postgresql

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aarondl/opt/omit"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/types"

	dbctx "github.com/yoshioka0101/voiceblog/backend/internal/db"
	entity "github.com/yoshioka0101/voiceblog/backend/internal/entity/transcription"
	transcriptionusecase "github.com/yoshioka0101/voiceblog/backend/internal/usecase/transcription"
	"github.com/yoshioka0101/voiceblog/backend/models"
)

type TranscriptionRepository struct {
	db bob.Executor
}

func NewTranscriptionRepository(db *sql.DB) *TranscriptionRepository {
	return &TranscriptionRepository{db: bob.NewDB(db)}
}

func NewTranscriptionRepositoryWithExecutor(exec bob.Executor) *TranscriptionRepository {
	return &TranscriptionRepository{db: exec}
}

func (r *TranscriptionRepository) CreateTranscription(ctx context.Context, transcriptionEntity *entity.Transcription) (*entity.Transcription, error) {
	exec := dbctx.ExecutorFromContext(ctx, r.db)
	value, err := models.Transcriptions.Insert(&models.TranscriptionSetter{
		UserID:       omit.From(transcriptionEntity.UserID),
		FullText:     omit.From(transcriptionEntity.FullText),
		SegmentsJSON: omit.From(types.NewJSON(json.RawMessage(transcriptionEntity.SegmentsJSON))),
	}).One(ctx, exec)
	if err != nil {
		return nil, fmt.Errorf("insert transcription: %w", err)
	}

	return toTranscriptionEntity(value), nil
}

func (r *TranscriptionRepository) FindTranscriptionByID(ctx context.Context, id int64) (*entity.Transcription, error) {
	exec := dbctx.ExecutorFromContext(ctx, r.db)
	value, err := models.FindTranscription(ctx, exec, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, transcriptionusecase.ErrNotFound
		}
		return nil, fmt.Errorf("find transcription: %w", err)
	}

	return toTranscriptionEntity(value), nil
}

func toTranscriptionEntity(value *models.Transcription) *entity.Transcription {
	if value == nil {
		return nil
	}

	return &entity.Transcription{
		ID:           value.ID,
		UserID:       value.UserID,
		FullText:     value.FullText,
		SegmentsJSON: json.RawMessage(append([]byte(nil), value.SegmentsJSON.Val...)),
		CreatedAt:    value.CreatedAt,
		UpdatedAt:    value.UpdatedAt,
	}
}
