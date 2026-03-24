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

	entity "github.com/yoshioka0101/voiceblog/backend/internal/entity/transcription"
	transcriptionusecase "github.com/yoshioka0101/voiceblog/backend/internal/usecase/transcription"
	"github.com/yoshioka0101/voiceblog/backend/models"
)

type TranscriptionRepository struct {
	db bob.DB
}

func NewTranscriptionRepository(db *sql.DB) *TranscriptionRepository {
	return &TranscriptionRepository{db: bob.NewDB(db)}
}

func (r *TranscriptionRepository) Create(ctx context.Context, transcriptionEntity *entity.Transcription) (*entity.Transcription, error) {
	value, err := models.Transcriptions.Insert(&models.TranscriptionSetter{
		UserID:       omit.From(transcriptionEntity.UserID),
		FullText:     omit.From(transcriptionEntity.FullText),
		SegmentsJSON: omit.From(types.NewJSON(json.RawMessage(transcriptionEntity.SegmentsJSON))),
	}).One(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("insert transcription: %w", err)
	}

	return toTranscriptionEntity(value), nil
}


func (r *TranscriptionRepository) FindByID(ctx context.Context, id int64) (*entity.Transcription, error) {
	value, err := models.FindTranscription(ctx, r.db, id)
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
