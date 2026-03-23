package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/yoshioka0101/voiceblog/backend/internal/apperr"
	entity "github.com/yoshioka0101/voiceblog/backend/internal/entity/transcription"
)

var (
	ErrNotFound  = apperr.New(http.StatusNotFound, "transcription not found")
	ErrForbidden = apperr.New(http.StatusForbidden, "forbidden")
)

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
		return nil, apperr.New(http.StatusBadRequest, "full_text is required")
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

func (uc *UseCase) ListTranscriptionsByUserID(ctx context.Context, userID int64) ([]*entity.Transcription, error) {
	return uc.repo.ListByUserID(ctx, userID)
}

func (uc *UseCase) GetTranscription(ctx context.Context, userID, transcriptionID int64) (*entity.Transcription, error) {
	transcription, err := uc.repo.FindByID(ctx, transcriptionID)
	if err != nil {
		return nil, err
	}
	if transcription.UserID != userID {
		return nil, ErrForbidden
	}
	return transcription, nil
}

func normalizeSegmentsJSON(raw json.RawMessage) (json.RawMessage, error) {
	if len(bytes.TrimSpace(raw)) == 0 {
		return nil, apperr.New(http.StatusBadRequest, "segments_json must not be empty")
	}

	var segments []map[string]any
	if err := json.Unmarshal(raw, &segments); err != nil {
		return nil, apperr.New(http.StatusBadRequest, "segments_json must be valid JSON")
	}
	if len(segments) == 0 {
		return nil, apperr.New(http.StatusBadRequest, "segments_json must not be empty")
	}

	normalized, err := json.Marshal(segments)
	if err != nil {
		return nil, apperr.New(http.StatusBadRequest, "segments_json must be valid JSON")
	}

	return normalized, nil
}
