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
		findTranscriptionByIDFunc: func(_ context.Context, id int64) (*transcriptionEntity.Transcription, error) {
			return &transcriptionEntity.Transcription{
				ID:           id,
				UserID:       userID,
				FullText:     "full text",
				SegmentsJSON: json.RawMessage(`[{"text":"full text"}]`),
			}, nil
		},
	}
	promptRepo := &promptRepositoryStub{
		findPromptByIDFunc: func(_ context.Context, id int64) (*promptEntity.Prompt, error) {
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
		findTranscriptionByIDFunc: func(_ context.Context, id int64) (*transcriptionEntity.Transcription, error) {
			return &transcriptionEntity.Transcription{ID: id, UserID: userID, FullText: "text", SegmentsJSON: json.RawMessage(`[{}]`)}, nil
		},
	}
	promptRepo := &promptRepositoryStub{
		findPromptByIDFunc: func(_ context.Context, id int64) (*promptEntity.Prompt, error) {
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
		findTranscriptionByIDFunc: func(_ context.Context, id int64) (*transcriptionEntity.Transcription, error) {
			return &transcriptionEntity.Transcription{ID: id, UserID: 99}, nil
		},
	}
	uc := promptrunjobusecase.NewUseCase(repo, transcriptionRepo, &promptRepositoryStub{}, &generatorStub{})

	_, err := uc.GetPromptRunJob(context.Background(), 7, 3)
	if !errors.Is(err, promptrunjobusecase.ErrForbidden) {
		t.Fatalf("err = %v, want ErrForbidden", err)
	}
}

func TestCreatePromptRunJob_UsesTransactionRunner(t *testing.T) {
	userID := int64(7)
	txRunner := &txRunnerStub{}
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
		findTranscriptionByIDFunc: func(_ context.Context, id int64) (*transcriptionEntity.Transcription, error) {
			return &transcriptionEntity.Transcription{
				ID:           id,
				UserID:       userID,
				FullText:     "full text",
				SegmentsJSON: json.RawMessage(`[{"text":"full text"}]`),
			}, nil
		},
	}
	promptRepo := &promptRepositoryStub{
		findPromptByIDFunc: func(_ context.Context, id int64) (*promptEntity.Prompt, error) {
			return &promptEntity.Prompt{ID: id, Name: "blog", Body: "write", IsActive: true}, nil
		},
	}
	generator := &generatorStub{
		generateArticleFunc: func(_ context.Context, _ promptrunjobusecase.GenerateArticleInput) (*promptrunjobusecase.GeneratedArticle, error) {
			return &promptrunjobusecase.GeneratedArticle{
				Title:   "generated title",
				Content: "generated content",
			}, nil
		},
	}

	uc := promptrunjobusecase.NewUseCase(repo, transcriptionRepo, promptRepo, generator, txRunner)
	_, err := uc.CreatePromptRunJob(context.Background(), promptrunjobusecase.CreatePromptRunJobInput{
		UserID:          userID,
		TranscriptionID: 2,
		PromptID:        3,
	})
	if err != nil {
		t.Fatalf("CreatePromptRunJob failed: %v", err)
	}
	if !txRunner.called {
		t.Fatal("transaction runner was not used")
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
	createTranscriptionFunc   func(ctx context.Context, transcription *transcriptionEntity.Transcription) (*transcriptionEntity.Transcription, error)
	listByUserIDFunc          func(ctx context.Context, userID int64) ([]*transcriptionEntity.Transcription, error)
	findTranscriptionByIDFunc func(ctx context.Context, id int64) (*transcriptionEntity.Transcription, error)
}

func (s *transcriptionRepositoryStub) CreateTranscription(ctx context.Context, transcription *transcriptionEntity.Transcription) (*transcriptionEntity.Transcription, error) {
	if s.createTranscriptionFunc == nil {
		return nil, errors.New("createTranscriptionFunc is nil")
	}
	return s.createTranscriptionFunc(ctx, transcription)
}

func (s *transcriptionRepositoryStub) ListByUserID(ctx context.Context, userID int64) ([]*transcriptionEntity.Transcription, error) {
	if s.listByUserIDFunc == nil {
		return nil, errors.New("listByUserIDFunc is nil")
	}
	return s.listByUserIDFunc(ctx, userID)
}

func (s *transcriptionRepositoryStub) FindTranscriptionByID(ctx context.Context, id int64) (*transcriptionEntity.Transcription, error) {
	if s.findTranscriptionByIDFunc == nil {
		return nil, errors.New("findTranscriptionByIDFunc is nil")
	}
	return s.findTranscriptionByIDFunc(ctx, id)
}

type promptRepositoryStub struct {
	listVisiblePromptsByUserIDFunc func(ctx context.Context, userID int64) ([]*promptEntity.Prompt, error)
	createPromptFunc               func(ctx context.Context, prompt *promptEntity.Prompt) (*promptEntity.Prompt, error)
	findPromptByIDFunc             func(ctx context.Context, id int64) (*promptEntity.Prompt, error)
	updatePromptFunc               func(ctx context.Context, prompt *promptEntity.Prompt) (*promptEntity.Prompt, error)
	deletePromptFunc               func(ctx context.Context, id int64) error
}

func (s *promptRepositoryStub) ListVisiblePromptsByUserID(ctx context.Context, userID int64) ([]*promptEntity.Prompt, error) {
	if s.listVisiblePromptsByUserIDFunc == nil {
		return nil, errors.New("listVisiblePromptsByUserIDFunc is nil")
	}
	return s.listVisiblePromptsByUserIDFunc(ctx, userID)
}

func (s *promptRepositoryStub) CreatePrompt(ctx context.Context, prompt *promptEntity.Prompt) (*promptEntity.Prompt, error) {
	if s.createPromptFunc == nil {
		return nil, errors.New("createPromptFunc is nil")
	}
	return s.createPromptFunc(ctx, prompt)
}

func (s *promptRepositoryStub) FindPromptByID(ctx context.Context, id int64) (*promptEntity.Prompt, error) {
	if s.findPromptByIDFunc == nil {
		return nil, errors.New("findPromptByIDFunc is nil")
	}
	return s.findPromptByIDFunc(ctx, id)
}

func (s *promptRepositoryStub) UpdatePrompt(ctx context.Context, prompt *promptEntity.Prompt) (*promptEntity.Prompt, error) {
	if s.updatePromptFunc == nil {
		return nil, errors.New("updatePromptFunc is nil")
	}
	return s.updatePromptFunc(ctx, prompt)
}

func (s *promptRepositoryStub) DeletePrompt(ctx context.Context, id int64) error {
	if s.deletePromptFunc == nil {
		return errors.New("deletePromptFunc is nil")
	}
	return s.deletePromptFunc(ctx, id)
}

type txRunnerStub struct {
	called bool
}

func (s *txRunnerStub) RunInTx(ctx context.Context, fn func(context.Context) error) error {
	s.called = true
	return fn(ctx)
}
