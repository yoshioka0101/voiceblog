package validation

import (
	"strings"

	"github.com/yoshioka0101/voiceblog/backend/internal/api"
	"github.com/yoshioka0101/voiceblog/backend/internal/apperr"
	usecase "github.com/yoshioka0101/voiceblog/backend/internal/usecase/prompt"
)

func CreatePromptInput(userID int64, req api.CreatePromptRequest) (usecase.CreatePromptInput, error) {
	if strings.TrimSpace(req.Name) == "" {
		return usecase.CreatePromptInput{}, apperr.BadRequestWithCode("name_required", "name is required")
	}
	if strings.TrimSpace(req.Body) == "" {
		return usecase.CreatePromptInput{}, apperr.BadRequestWithCode("body_required", "body is required")
	}

	return usecase.CreatePromptInput{
		UserID:   userID,
		Name:     req.Name,
		Body:     req.Body,
		IsActive: req.IsActive,
	}, nil
}

func UpdatePromptInput(userID, promptID int64, req api.UpdatePromptRequest) (usecase.UpdatePromptInput, error) {
	if req.Name == nil && req.Body == nil && req.IsActive == nil {
		return usecase.UpdatePromptInput{}, apperr.BadRequestWithCode("field_required", "at least one field is required")
	}
	if req.Name != nil && strings.TrimSpace(*req.Name) == "" {
		return usecase.UpdatePromptInput{}, apperr.BadRequestWithCode("name_required", "name must not be blank")
	}
	if req.Body != nil && strings.TrimSpace(*req.Body) == "" {
		return usecase.UpdatePromptInput{}, apperr.BadRequestWithCode("body_required", "body must not be blank")
	}

	return usecase.UpdatePromptInput{
		UserID:   userID,
		PromptID: promptID,
		Name:     req.Name,
		Body:     req.Body,
		IsActive: req.IsActive,
	}, nil
}
