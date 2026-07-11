package usecase

import (
	"context"
	"github.com/yoshioka0101/voiceblog/backend/internal/entity/repository"
	"strings"
	"time"

	"github.com/yoshioka0101/voiceblog/backend/internal/apperr"
	dbtx "github.com/yoshioka0101/voiceblog/backend/internal/db"
	promptEntity "github.com/yoshioka0101/voiceblog/backend/internal/entity/prompt"
	entity "github.com/yoshioka0101/voiceblog/backend/internal/entity/promptrunjob"
	transcriptionEntity "github.com/yoshioka0101/voiceblog/backend/internal/entity/transcription"
	"github.com/yoshioka0101/voiceblog/backend/internal/usecase/articlegen"
)

var (
	ErrNotFound  = apperr.NotFound("prompt run job")
	ErrForbidden = apperr.ErrForbidden
)

type Generator interface {
	GenerateArticle(ctx context.Context, input GenerateArticleInput) (*GeneratedArticle, error)
}

type GenerateArticleInput = articlegen.Input

type GeneratedArticle = articlegen.GeneratedArticle

type CreatePromptRunJobInput struct {
	UserID          int64
	TranscriptionID int64
	PromptID        int64
}

type UseCase struct {
	repo              repository.PromptRunJobRepository
	transcriptionRepo repository.TranscriptionRepository
	promptRepo        repository.PromptRepository
	generator         articlegen.Generator
	txRunner          dbtx.TxRunner
}

func NewUseCase(
	repo repository.PromptRunJobRepository,
	transcriptionRepo repository.TranscriptionRepository,
	promptRepo repository.PromptRepository,
	generator articlegen.Generator,
	txRunners ...dbtx.TxRunner,
) *UseCase {
	var txRunner dbtx.TxRunner
	if len(txRunners) > 0 {
		txRunner = txRunners[0]
	}

	return &UseCase{
		repo:              repo,
		transcriptionRepo: transcriptionRepo,
		promptRepo:        promptRepo,
		generator:         generator,
		txRunner:          txRunner,
	}
}

func (uc *UseCase) CreatePromptRunJob(ctx context.Context, input CreatePromptRunJobInput) (*entity.Job, error) {
	transcriptionValue, promptValue, job, err := uc.startPromptRunJob(ctx, input)
	if err != nil {
		return nil, err
	}

	generated, err := uc.generator.GenerateArticle(ctx, articlegen.Input{
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

func (uc *UseCase) startPromptRunJob(ctx context.Context, input CreatePromptRunJobInput) (*transcriptionEntity.Transcription, *promptEntity.Prompt, *entity.Job, error) {
	var (
		transcriptionValue *transcriptionEntity.Transcription
		promptValue        *promptEntity.Prompt
		job                *entity.Job
	)

	run := func(runCtx context.Context) error {
		var err error

		transcriptionValue, err = uc.transcriptionRepo.FindTranscriptionByID(runCtx, input.TranscriptionID)
		if err != nil {
			return err
		}
		if transcriptionValue.UserID != input.UserID {
			return ErrForbidden
		}

		promptValue, err = uc.promptRepo.FindPromptByID(runCtx, input.PromptID)
		if err != nil {
			return err
		}
		if !visibleToUser(promptValue, input.UserID) {
			return ErrForbidden
		}

		now := time.Now()
		job, err = uc.repo.CreatePromptRunJob(runCtx, &entity.Job{
			TranscriptionID: input.TranscriptionID,
			PromptID:        input.PromptID,
			Status:          entity.StatusPending,
			AttemptCount:    0,
			NextRunAt:       now,
		})
		if err != nil {
			return err
		}

		job.Status = entity.StatusRunning
		job.AttemptCount = 1
		job.NextRunAt = time.Now()
		job, err = uc.repo.UpdatePromptRunJob(runCtx, job)
		return err
	}

	if uc.txRunner == nil {
		if err := run(ctx); err != nil {
			return nil, nil, nil, err
		}
		return transcriptionValue, promptValue, job, nil
	}

	if err := uc.txRunner.RunInTx(ctx, run); err != nil {
		return nil, nil, nil, err
	}

	return transcriptionValue, promptValue, job, nil
}

func (uc *UseCase) GetPromptRunJob(ctx context.Context, userID, jobID int64) (*entity.Job, error) {
	job, err := uc.repo.FindPromptRunJobByID(ctx, jobID)
	if err != nil {
		return nil, err
	}

	transcriptionValue, err := uc.transcriptionRepo.FindTranscriptionByID(ctx, job.TranscriptionID)
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
