package presenter

import (
	"github.com/yoshioka0101/voiceblog/backend/internal/api"
	entity "github.com/yoshioka0101/voiceblog/backend/internal/entity/prompt"
)

func Prompt(value *entity.Prompt) api.PromptResponse {
	return api.PromptResponse{
		Id:        value.ID,
		UserId:    value.UserID,
		Name:      value.Name,
		Body:      value.Body,
		IsActive:  value.IsActive,
		IsDefault: value.IsDefault,
		CreatedAt: value.CreatedAt,
		UpdatedAt: value.UpdatedAt,
	}
}

func Prompts(values []*entity.Prompt) api.PromptsResponse {
	responses := make(api.PromptsResponse, 0, len(values))
	for _, value := range values {
		responses = append(responses, Prompt(value))
	}
	return responses
}
