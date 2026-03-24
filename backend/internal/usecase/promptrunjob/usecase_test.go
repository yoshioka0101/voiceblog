package usecase_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	promptEntity "github.com/yoshioka0101/voiceblog/backend/internal/entity/prompt"
	entity "github.com/yoshioka0101/voiceblog/backend/internal/entity/promptrunjob"
	transcriptionEntity "github.com/yoshioka0101/voiceblog/backend/internal/entity/transcription"
	promptrunjobusecase "github.com/yoshioka0101/voiceblog/backend/internal/usecase/promptrunjob"
)

func TestCreatePromptRunJob_CompletesAndStoresGeneratedContent(t *testing.T) {
	userID := int64(7)
	repo := &promptRunJobRepositoryStub{
		createPromptRunJobFunc: func(_ context.Context, value *entity.Job) (*entity.Job, error) {
			return &entity.Job{
				ID:              13,
				TranscriptionID: value.TranscriptionID,
				PromptID:        value.PromptID,
				Status:          value.Status,
				AttemptCount:    value.AttemptCount,
				NextRunAt:       value.NextRunAt,
				CreatedAt:       time.Now(),
			}, nil
		},
		updatePromptRunJobFunc: func(_ context.Context, value *entity.Job) (*entity.Job, error) {
			return value, nil
		},
	}
	transcriptionRepo := &transcriptionRepositoryStub{
		findByIDFunc: func(_ context.Context, id int64) (*transcriptionEntity.Transcription, error) {
			return &transcriptionEntity.Transcription{
				ID:           id,
				UserID:       userID,
				FullText:     "full text",
				SegmentsJSON: json.RawMessage(`[{"text":"full text"}]`),
			}, nil
		},
	}
	promptRepo := &promptRepositoryStub{
		findByIDFunc: func(_ context.Context, id int64) (*promptEntity.Prompt, error) {
			return &promptEntity.Prompt{ID: id, Name: "blog", Body: "write", IsActive: true}, nil
		},
	}
	generator := &generatorStub{
		generateArticleFunc: func(_ context.Context, input promptrunjobusecase.GenerateArticleInput) (*promptrunjobusecase.GeneratedArticle, error) {
			if input.FullText != "full text" {
				t.Fatalf("FullText = %q", input.FullText)
			}
			return &promptrunjobusecase.GeneratedArticle{
				Title:   "generated title",
				Content: "generated content",
			}, nil
		},
	}

	uc := promptrunjobusecase.NewUseCase(repo, transcriptionRepo, promptRepo, generator)
	job, err := uc.CreatePromptRunJob(context.Background(), promptrunjobusecase.CreatePromptRunJobInput{
		UserID:          userID,
		TranscriptionID: 2,
		PromptID:        3,
	})
	if err != nil {
		t.Fatalf("CreatePromptRunJob failed: %v", err)
	}
	if job.Status != entity.StatusCompleted {
		t.Fatalf("Status = %q, want completed", job.Status)
	}
	if job.GeneratedTitle == nil || *job.GeneratedTitle != "generated title" {
		t.Fatalf("GeneratedTitle = %#v", job.GeneratedTitle)
	}
	if job.GeneratedContent == nil || *job.GeneratedContent != "generated content" {
		t.Fatalf("GeneratedContent = %#v", job.GeneratedContent)
	}
}

func TestCreatePromptRunJob_FailsWhenGeneratorFails(t *testing.T) {
	userID := int64(5)
	repo := &promptRunJobRepositoryStub{
		createPromptRunJobFunc: func(_ context.Context, value *entity.Job) (*entity.Job, error) {
			return &entity.Job{
				ID:              8,
				TranscriptionID: value.TranscriptionID,
				PromptID:        value.PromptID,
				Status:          value.Status,
				AttemptCount:    value.AttemptCount,
				NextRunAt:       value.NextRunAt,
				CreatedAt:       time.Now(),
			}, nil
		},
		updatePromptRunJobFunc: func(_ context.Context, value *entity.Job) (*entity.Job, error) {
			return value, nil
		},
	}
	transcriptionRepo := &transcriptionRepositoryStub{
		findByIDFunc: func(_ context.Context, id int64) (*transcriptionEntity.Transcription, error) {
			return &transcriptionEntity.Transcription{ID: id, UserID: userID, FullText: "text", SegmentsJSON: json.RawMessage(`[{}]`)}, nil
		},
	}
	promptRepo := &promptRepositoryStub{
		findByIDFunc: func(_ context.Context, id int64) (*promptEntity.Prompt, error) {
			return &promptEntity.Prompt{ID: id, Name: "blog", Body: "write", IsActive: true}, nil
		},
	}
	generator := &generatorStub{
		generateArticleFunc: func(_ context.Context, _ promptrunjobusecase.GenerateArticleInput) (*promptrunjobusecase.GeneratedArticle, error) {
			return nil, errors.New("gemini api key is required")
		},
	}

	uc := promptrunjobusecase.NewUseCase(repo, transcriptionRepo, promptRepo, generator)
	job, err := uc.CreatePromptRunJob(context.Background(), promptrunjobusecase.CreatePromptRunJobInput{
		UserID:          userID,
		TranscriptionID: 1,
		PromptID:        1,
	})
	if err != nil {
		t.Fatalf("CreatePromptRunJob failed: %v", err)
	}
	if job.Status != entity.StatusFailed {
		t.Fatalf("Status = %q, want failed", job.Status)
	}
	if job.ErrorMessage == nil || *job.ErrorMessage == "" {
		t.Fatalf("ErrorMessage = %#v", job.ErrorMessage)
	}
}

func TestGetPromptRunJob_ForbiddenForOtherUser(t *testing.T) {
	repo := &promptRunJobRepositoryStub{
		findPromptRunJobByIDFunc: func(_ context.Context, id int64) (*entity.Job, error) {
			return &entity.Job{ID: id, TranscriptionID: 55}, nil
		},
	}
	transcriptionRepo := &transcriptionRepositoryStub{
		findByIDFunc: func(_ context.Context, id int64) (*transcriptionEntity.Transcription, error) {
			return &transcriptionEntity.Transcription{ID: id, UserID: 99}, nil
		},
	}
	uc := promptrunjobusecase.NewUseCase(repo, transcriptionRepo, &promptRepositoryStub{}, &generatorStub{})

	_, err := uc.GetPromptRunJob(context.Background(), 7, 3)
	if !errors.Is(err, promptrunjobusecase.ErrForbidden) {
		t.Fatalf("err = %v, want ErrForbidden", err)
	}
}

type promptRunJobRepositoryStub struct {
	createPromptRunJobFunc   func(ctx context.Context, value *entity.Job) (*entity.Job, error)
	findPromptRunJobByIDFunc func(ctx context.Context, id int64) (*entity.Job, error)
	updatePromptRunJobFunc   func(ctx context.Context, value *entity.Job) (*entity.Job, error)
}

func (s *promptRunJobRepositoryStub) CreatePromptRunJob(ctx context.Context, value *entity.Job) (*entity.Job, error) {
	return s.createPromptRunJobFunc(ctx, value)
}

func (s *promptRunJobRepositoryStub) FindPromptRunJobByID(ctx context.Context, id int64) (*entity.Job, error) {
	return s.findPromptRunJobByIDFunc(ctx, id)
}

func (s *promptRunJobRepositoryStub) UpdatePromptRunJob(ctx context.Context, value *entity.Job) (*entity.Job, error) {
	return s.updatePromptRunJobFunc(ctx, value)
}

type generatorStub struct {
	generateArticleFunc func(ctx context.Context, input promptrunjobusecase.GenerateArticleInput) (*promptrunjobusecase.GeneratedArticle, error)
}

func (s *generatorStub) GenerateArticle(ctx context.Context, input promptrunjobusecase.GenerateArticleInput) (*promptrunjobusecase.GeneratedArticle, error) {
	if s.generateArticleFunc == nil {
		return nil, errors.New("generateArticleFunc is nil")
	}
	return s.generateArticleFunc(ctx, input)
}

type transcriptionRepositoryStub struct {
	createFunc       func(ctx context.Context, transcription *transcriptionEntity.Transcription) (*transcriptionEntity.Transcription, error)
	listByUserIDFunc func(ctx context.Context, userID int64) ([]*transcriptionEntity.Transcription, error)
	findByIDFunc     func(ctx context.Context, id int64) (*transcriptionEntity.Transcription, error)
}

func (s *transcriptionRepositoryStub) Create(ctx context.Context, transcription *transcriptionEntity.Transcription) (*transcriptionEntity.Transcription, error) {
	if s.createFunc == nil {
		return nil, errors.New("createFunc is nil")
	}
	return s.createFunc(ctx, transcription)
}

func (s *transcriptionRepositoryStub) ListByUserID(ctx context.Context, userID int64) ([]*transcriptionEntity.Transcription, error) {
	if s.listByUserIDFunc == nil {
		return nil, errors.New("listByUserIDFunc is nil")
	}
	return s.listByUserIDFunc(ctx, userID)
}

func (s *transcriptionRepositoryStub) FindByID(ctx context.Context, id int64) (*transcriptionEntity.Transcription, error) {
	if s.findByIDFunc == nil {
		return nil, errors.New("findByIDFunc is nil")
	}
	return s.findByIDFunc(ctx, id)
}

type promptRepositoryStub struct {
	listVisibleByUserIDFunc func(ctx context.Context, userID int64) ([]*promptEntity.Prompt, error)
	createFunc              func(ctx context.Context, prompt *promptEntity.Prompt) (*promptEntity.Prompt, error)
	findByIDFunc            func(ctx context.Context, id int64) (*promptEntity.Prompt, error)
	updateFunc              func(ctx context.Context, prompt *promptEntity.Prompt) (*promptEntity.Prompt, error)
	deleteFunc              func(ctx context.Context, id int64) error
}

func (s *promptRepositoryStub) ListVisibleByUserID(ctx context.Context, userID int64) ([]*promptEntity.Prompt, error) {
	if s.listVisibleByUserIDFunc == nil {
		return nil, errors.New("listVisibleByUserIDFunc is nil")
	}
	return s.listVisibleByUserIDFunc(ctx, userID)
}

func (s *promptRepositoryStub) Create(ctx context.Context, prompt *promptEntity.Prompt) (*promptEntity.Prompt, error) {
	if s.createFunc == nil {
		return nil, errors.New("createFunc is nil")
	}
	return s.createFunc(ctx, prompt)
}

func (s *promptRepositoryStub) FindByID(ctx context.Context, id int64) (*promptEntity.Prompt, error) {
	if s.findByIDFunc == nil {
		return nil, errors.New("findByIDFunc is nil")
	}
	return s.findByIDFunc(ctx, id)
}

func (s *promptRepositoryStub) Update(ctx context.Context, prompt *promptEntity.Prompt) (*promptEntity.Prompt, error) {
	if s.updateFunc == nil {
		return nil, errors.New("updateFunc is nil")
	}
	return s.updateFunc(ctx, prompt)
}

func (s *promptRepositoryStub) Delete(ctx context.Context, id int64) error {
	if s.deleteFunc == nil {
		return errors.New("deleteFunc is nil")
	}
	return s.deleteFunc(ctx, id)
}
