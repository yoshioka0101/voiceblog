package usecase

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/yoshioka0101/voiceblog/backend/internal/apperr"
	promptEntity "github.com/yoshioka0101/voiceblog/backend/internal/entity/prompt"
	entity "github.com/yoshioka0101/voiceblog/backend/internal/entity/promptrunjob"
	transcriptionEntity "github.com/yoshioka0101/voiceblog/backend/internal/entity/transcription"
)

var (
	ErrNotFound  = apperr.NotFound("prompt run job")
	ErrForbidden = apperr.ErrForbidden
)

type Generator interface {
	GenerateArticle(ctx context.Context, input GenerateArticleInput) (*GeneratedArticle, error)
}

type GenerateArticleInput struct {
	PromptName   string
	PromptBody   string
	FullText     string
	SegmentsJSON json.RawMessage
}

type GeneratedArticle struct {
	Title   string
	Content string
}

type CreatePromptRunJobInput struct {
	UserID          int64
	TranscriptionID int64
	PromptID        int64
}

type UseCase struct {
	repo              entity.Repository
	transcriptionRepo transcriptionEntity.Repository
	promptRepo        promptEntity.Repository
	generator         Generator
}

func NewUseCase(
	repo entity.Repository,
	transcriptionRepo transcriptionEntity.Repository,
	promptRepo promptEntity.Repository,
	generator Generator,
) *UseCase {
	return &UseCase{
		repo:              repo,
		transcriptionRepo: transcriptionRepo,
		promptRepo:        promptRepo,
		generator:         generator,
	}
}

func (uc *UseCase) CreatePromptRunJob(ctx context.Context, input CreatePromptRunJobInput) (*entity.Job, error) {
	transcriptionValue, err := uc.transcriptionRepo.FindByID(ctx, input.TranscriptionID)
	if err != nil {
		return nil, err
	}
	if transcriptionValue.UserID != input.UserID {
		return nil, ErrForbidden
	}

	promptValue, err := uc.promptRepo.FindByID(ctx, input.PromptID)
	if err != nil {
		return nil, err
	}
	if !visibleToUser(promptValue, input.UserID) {
		return nil, ErrForbidden
	}

	now := time.Now()
	job, err := uc.repo.CreatePromptRunJob(ctx, &entity.Job{
		TranscriptionID: input.TranscriptionID,
		PromptID:        input.PromptID,
		Status:          entity.StatusPending,
		AttemptCount:    0,
		NextRunAt:       now,
	})
	if err != nil {
		return nil, err
	}

	job.Status = entity.StatusRunning
	job.AttemptCount = 1
	job.NextRunAt = time.Now()
	job, err = uc.repo.UpdatePromptRunJob(ctx, job)
	if err != nil {
		return nil, err
	}

	generated, err := uc.generator.GenerateArticle(ctx, GenerateArticleInput{
		PromptName:   promptValue.Name,
		PromptBody:   promptValue.Body,
		FullText:     transcriptionValue.FullText,
		SegmentsJSON: transcriptionValue.SegmentsJSON,
	})
	if err != nil {
		return uc.failJob(ctx, job, err)
	}

	generatedTitle := strings.TrimSpace(generated.Title)
	generatedContent := strings.TrimSpace(generated.Content)
	if generatedTitle == "" || generatedContent == "" {
		return uc.failJob(ctx, job, apperr.InternalError("generated article is empty"))
	}

	job.Status = entity.StatusCompleted
	job.ErrorMessage = nil
	job.GeneratedTitle = &generatedTitle
	job.GeneratedContent = &generatedContent
	job.NextRunAt = time.Now()
	job, err = uc.repo.UpdatePromptRunJob(ctx, job)
	if err != nil {
		return nil, err
	}

	return job, nil
}

func (uc *UseCase) GetPromptRunJob(ctx context.Context, userID, jobID int64) (*entity.Job, error) {
	job, err := uc.repo.FindPromptRunJobByID(ctx, jobID)
	if err != nil {
		return nil, err
	}

	transcriptionValue, err := uc.transcriptionRepo.FindByID(ctx, job.TranscriptionID)
	if err != nil {
		return nil, err
	}
	if transcriptionValue.UserID != userID {
		return nil, ErrForbidden
	}

	return job, nil
}

func (uc *UseCase) failJob(ctx context.Context, job *entity.Job, source error) (*entity.Job, error) {
	message := strings.TrimSpace(source.Error())
	if message == "" {
		message = "failed to generate article"
	}

	job.Status = entity.StatusFailed
	job.ErrorMessage = &message
	job.GeneratedTitle = nil
	job.GeneratedContent = nil
	job.NextRunAt = time.Now()

	return uc.repo.UpdatePromptRunJob(ctx, job)
}

func visibleToUser(value *promptEntity.Prompt, userID int64) bool {
	if !value.IsActive {
		return false
	}
	if value.UserID == nil {
		return true
	}
	return *value.UserID == userID
}
