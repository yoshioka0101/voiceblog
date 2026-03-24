package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"

	"github.com/yoshioka0101/voiceblog/backend/internal/apperr"
	entity "github.com/yoshioka0101/voiceblog/backend/internal/entity/transcription"
)

var ErrNotFound = apperr.NotFound("transcription")

type CreateTranscriptionInput struct {
	UserID       int64
	FullText     string
	SegmentsJSON json.RawMessage
}

type UseCase struct {
	repo entity.Repository
}

func NewUseCase(repo entity.Repository) *UseCase {
	return &UseCase{repo: repo}
}

func (uc *UseCase) CreateTranscription(ctx context.Context, input CreateTranscriptionInput) (*entity.Transcription, error) {
	fullText := strings.TrimSpace(input.FullText)
	if fullText == "" {
		return nil, apperr.BadRequest( "full_text is required")
	}

	segmentsJSON, err := normalizeSegmentsJSON(input.SegmentsJSON)
	if err != nil {
		return nil, err
	}

	return uc.repo.Create(ctx, &entity.Transcription{
		UserID:       input.UserID,
		FullText:     fullText,
		SegmentsJSON: segmentsJSON,
	})
}


func normalizeSegmentsJSON(raw json.RawMessage) (json.RawMessage, error) {
	if len(bytes.TrimSpace(raw)) == 0 {
		return nil, apperr.BadRequest( "segments_json must not be empty")
	}

	var segments []map[string]any
	if err := json.Unmarshal(raw, &segments); err != nil {
		return nil, apperr.BadRequest( "segments_json must be valid JSON")
	}
	if len(segments) == 0 {
		return nil, apperr.BadRequest( "segments_json must not be empty")
	}

	normalized, err := json.Marshal(segments)
	if err != nil {
		return nil, apperr.BadRequest( "segments_json must be valid JSON")
	}

	return normalized, nil
}
