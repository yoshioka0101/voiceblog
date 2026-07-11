package validation

import (
	"encoding/json"
	"strings"

	"github.com/yoshioka0101/voiceblog/backend/internal/api"
	"github.com/yoshioka0101/voiceblog/backend/internal/apperr"
	usecase "github.com/yoshioka0101/voiceblog/backend/internal/usecase/transcription"
)

func CreateTranscriptionInput(userID int64, req api.CreateTranscriptionRequest) (usecase.CreateTranscriptionInput, error) {
	if strings.TrimSpace(req.FullText) == "" {
		return usecase.CreateTranscriptionInput{}, apperr.BadRequestWithCode("full_text_required", "full_text is required")
	}
	if len(req.SegmentsJson) == 0 {
		return usecase.CreateTranscriptionInput{}, apperr.BadRequestWithCode("segments_required", "segments_json must not be empty")
	}

	segmentsJSON, err := json.Marshal(req.SegmentsJson)
	if err != nil {
		return usecase.CreateTranscriptionInput{}, apperr.BadRequestWithCode("invalid_segments", "segments_json must be valid JSON")
	}

	return usecase.CreateTranscriptionInput{
		UserID:       userID,
		FullText:     req.FullText,
		SegmentsJSON: segmentsJSON,
	}, nil
}
