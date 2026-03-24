package presenter

import (
	"github.com/yoshioka0101/voiceblog/backend/internal/api"
	entity "github.com/yoshioka0101/voiceblog/backend/internal/entity/promptrunjob"
)

func PromptRunJob(value *entity.Job) api.PromptRunJobResponse {
	response := api.PromptRunJobResponse{
		Id:               value.ID,
		TranscriptionId:  value.TranscriptionID,
		PromptId:         value.PromptID,
		Status:           value.Status,
		AttemptCount:     value.AttemptCount,
		NextRunAt:        value.NextRunAt,
		ErrorMessage:     value.ErrorMessage,
		GeneratedTitle:   value.GeneratedTitle,
		GeneratedContent: value.GeneratedContent,
		CreatedAt:        value.CreatedAt,
	}
	return response
}
