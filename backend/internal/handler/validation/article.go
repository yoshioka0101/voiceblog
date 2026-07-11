package validation

import (
	"strings"

	"github.com/yoshioka0101/voiceblog/backend/internal/api"
	"github.com/yoshioka0101/voiceblog/backend/internal/apperr"
	usecase "github.com/yoshioka0101/voiceblog/backend/internal/usecase/article"
)

func CreateArticleInput(userID int64, req api.CreateArticleRequest) (usecase.CreateArticleInput, error) {
	if strings.TrimSpace(req.Title) == "" {
		return usecase.CreateArticleInput{}, apperr.BadRequestWithCode("title_required", "title is required")
	}
	if strings.TrimSpace(req.Content) == "" {
		return usecase.CreateArticleInput{}, apperr.BadRequestWithCode("content_required", "content is required")
	}
	if req.PromptRunJobId != nil && *req.PromptRunJobId <= 0 {
		return usecase.CreateArticleInput{}, apperr.BadRequestWithCode("invalid_id", "prompt_run_job_id must be positive")
	}

	return usecase.CreateArticleInput{
		UserID:         userID,
		PromptRunJobID: req.PromptRunJobId,
		Title:          req.Title,
		Content:        req.Content,
	}, nil
}

func UpdateArticleInput(userID, articleID int64, req api.UpdateArticleRequest) (usecase.UpdateArticleInput, error) {
	if req.Title == nil && req.Content == nil {
		return usecase.UpdateArticleInput{}, apperr.BadRequestWithCode("field_required", "at least one field is required")
	}
	if req.Title != nil && strings.TrimSpace(*req.Title) == "" {
		return usecase.UpdateArticleInput{}, apperr.BadRequestWithCode("title_required", "title must not be blank")
	}
	if req.Content != nil && strings.TrimSpace(*req.Content) == "" {
		return usecase.UpdateArticleInput{}, apperr.BadRequestWithCode("content_required", "content must not be blank")
	}

	return usecase.UpdateArticleInput{
		UserID:    userID,
		ArticleID: articleID,
		Title:     req.Title,
		Content:   req.Content,
	}, nil
}

func GenerateArticleInput(userID int64, req api.GenerateArticleRequest) (usecase.GenerateArticleInput, error) {
	if req.TranscriptionId <= 0 {
		return usecase.GenerateArticleInput{}, apperr.BadRequestWithCode("invalid_id", "transcription_id must be positive")
	}
	if req.PromptId <= 0 {
		return usecase.GenerateArticleInput{}, apperr.BadRequestWithCode("invalid_id", "prompt_id must be positive")
	}

	return usecase.GenerateArticleInput{
		UserID:          userID,
		TranscriptionID: req.TranscriptionId,
		PromptID:        req.PromptId,
	}, nil
}
