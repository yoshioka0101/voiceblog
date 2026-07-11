package validation

import (
	"github.com/yoshioka0101/voiceblog/backend/internal/api"
	"github.com/yoshioka0101/voiceblog/backend/internal/apperr"
	usecase "github.com/yoshioka0101/voiceblog/backend/internal/usecase/promptrunjob"
)

func CreatePromptRunJobInput(userID int64, req api.CreatePromptRunJobRequest) (usecase.CreatePromptRunJobInput, error) {
	if req.TranscriptionId <= 0 {
		return usecase.CreatePromptRunJobInput{}, apperr.BadRequestWithCode("invalid_id", "transcription_id must be positive")
	}
	if req.PromptId <= 0 {
		return usecase.CreatePromptRunJobInput{}, apperr.BadRequestWithCode("invalid_id", "prompt_id must be positive")
	}

	return usecase.CreatePromptRunJobInput{
		UserID:          userID,
		TranscriptionID: req.TranscriptionId,
		PromptID:        req.PromptId,
	}, nil
}
