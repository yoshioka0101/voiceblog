package usecase_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/yoshioka0101/voiceblog/backend/internal/apperr"
	entity "github.com/yoshioka0101/voiceblog/backend/internal/entity/article"
	promptEntity "github.com/yoshioka0101/voiceblog/backend/internal/entity/prompt"
	promptrunjobEntity "github.com/yoshioka0101/voiceblog/backend/internal/entity/promptrunjob"
	transcriptionEntity "github.com/yoshioka0101/voiceblog/backend/internal/entity/transcription"
	articleusecase "github.com/yoshioka0101/voiceblog/backend/internal/usecase/article"
	"github.com/yoshioka0101/voiceblog/backend/internal/usecase/articlegen"
)

func TestCreateArticle(t *testing.T) {
	repo := &articleRepositoryStub{
		createArticleFunc: func(_ context.Context, value *entity.Article) (*entity.Article, error) {
			if value.UserID != 9 {
				t.Fatalf("UserID = %d, want 9", value.UserID)
			}
			if value.Title != "title" {
				t.Fatalf("Title = %q", value.Title)
			}
			return &entity.Article{
				ID:      11,
				UserID:  value.UserID,
				Title:   value.Title,
				Content: value.Content,
			}, nil
		},
	}

	uc := articleusecase.NewUseCase(repo, &promptRunJobRepositoryStub{}, &transcriptionRepositoryStub{})
	value, err := uc.CreateArticle(context.Background(), articleusecase.CreateArticleInput{
		UserID:  9,
		Title:   "title",
		Content: "content",
	})
	if err != nil {
		t.Fatalf("CreateArticle failed: %v", err)
	}
	if value.ID != 11 {
		t.Fatalf("ID = %d, want 11", value.ID)
	}
}

func TestGetArticle_Forbidden(t *testing.T) {
	repo := &articleRepositoryStub{
		findArticleByIDFunc: func(_ context.Context, id int64) (*entity.Article, error) {
			return &entity.Article{ID: id, UserID: 99}, nil
		},
	}

	uc := articleusecase.NewUseCase(repo, &promptRunJobRepositoryStub{}, &transcriptionRepositoryStub{})
	_, err := uc.GetArticle(context.Background(), 7, 3)
	if !errors.Is(err, articleusecase.ErrForbidden) {
		t.Fatalf("err = %v, want ErrForbidden", err)
	}
}

func TestCreateArticle_WithPromptRunJobID(t *testing.T) {
	jobID := int64(21)
	repo := &articleRepositoryStub{
		upsertArticleByPromptRunJobIDFunc: func(_ context.Context, value *entity.Article) (*entity.Article, error) {
			if value.PromptRunJobID == nil || *value.PromptRunJobID != jobID {
				t.Fatalf("PromptRunJobID = %#v", value.PromptRunJobID)
			}
			return &entity.Article{
				ID:             12,
				UserID:         value.UserID,
				PromptRunJobID: value.PromptRunJobID,
				Title:          value.Title,
				Content:        value.Content,
			}, nil
		},
	}
	promptRunJobRepo := &promptRunJobRepositoryStub{
		findPromptRunJobByIDFunc: func(_ context.Context, id int64) (*promptrunjobEntity.Job, error) {
			return &promptrunjobEntity.Job{ID: id, TranscriptionID: 33}, nil
		},
	}
	transcriptionRepo := &transcriptionRepositoryStub{
		findTranscriptionByIDFunc: func(_ context.Context, id int64) (*transcriptionEntity.Transcription, error) {
			return &transcriptionEntity.Transcription{ID: id, UserID: 9}, nil
		},
	}

	uc := articleusecase.NewUseCase(repo, promptRunJobRepo, transcriptionRepo)
	value, err := uc.CreateArticle(context.Background(), articleusecase.CreateArticleInput{
		UserID:         9,
		PromptRunJobID: &jobID,
		Title:          "title",
		Content:        "content",
	})
	if err != nil {
		t.Fatalf("CreateArticle failed: %v", err)
	}
	if value.PromptRunJobID == nil || *value.PromptRunJobID != jobID {
		t.Fatalf("PromptRunJobID = %#v", value.PromptRunJobID)
	}
}

func TestDeleteArticle(t *testing.T) {
	deletedID := int64(0)
	repo := &articleRepositoryStub{
		findArticleByIDFunc: func(_ context.Context, id int64) (*entity.Article, error) {
			return &entity.Article{ID: id, UserID: 5}, nil
		},
		deleteArticleFunc: func(_ context.Context, id int64) error {
			deletedID = id
			return nil
		},
	}

	uc := articleusecase.NewUseCase(repo, &promptRunJobRepositoryStub{}, &transcriptionRepositoryStub{})
	if err := uc.DeleteArticle(context.Background(), 5, 12); err != nil {
		t.Fatalf("DeleteArticle failed: %v", err)
	}
	if deletedID != 12 {
		t.Fatalf("deletedID = %d, want 12", deletedID)
	}
}

func TestGenerateArticle(t *testing.T) {
	repo := &articleRepositoryStub{
		createArticleFunc: func(_ context.Context, value *entity.Article) (*entity.Article, error) {
			if value.UserID != 9 {
				t.Fatalf("UserID = %d, want 9", value.UserID)
			}
			if value.Title != "generated title" {
				t.Fatalf("Title = %q", value.Title)
			}
			if value.Content != "generated content" {
				t.Fatalf("Content = %q", value.Content)
			}
			return &entity.Article{
				ID:      31,
				UserID:  value.UserID,
				Title:   value.Title,
				Content: value.Content,
			}, nil
		},
	}
	promptRepo := &promptRepositoryStub{
		findPromptByIDFunc: func(_ context.Context, id int64) (*promptEntity.Prompt, error) {
			return &promptEntity.Prompt{ID: id, Name: "blog", Body: "write", IsActive: true}, nil
		},
	}
	transcriptionRepo := &transcriptionRepositoryStub{
		findTranscriptionByIDFunc: func(_ context.Context, id int64) (*transcriptionEntity.Transcription, error) {
			return &transcriptionEntity.Transcription{
				ID:           id,
				UserID:       9,
				FullText:     "full text",
				SegmentsJSON: json.RawMessage(`[{"text":"full text"}]`),
			}, nil
		},
	}
	generator := &generatorStub{
		generateArticleFunc: func(_ context.Context, input articlegen.Input) (*articlegen.GeneratedArticle, error) {
			if input.PromptName != "blog" {
				t.Fatalf("PromptName = %q", input.PromptName)
			}
			if input.FullText != "full text" {
				t.Fatalf("FullText = %q", input.FullText)
			}
			return &articlegen.GeneratedArticle{
				Title:   "  generated title  ",
				Content: "  generated content  ",
			}, nil
		},
	}

	uc := articleusecase.NewUseCase(repo, &promptRunJobRepositoryStub{}, transcriptionRepo).WithGenerator(promptRepo, generator)
	value, err := uc.GenerateArticle(context.Background(), articleusecase.GenerateArticleInput{
		UserID:          9,
		TranscriptionID: 5,
		PromptID:        6,
	})
	if err != nil {
		t.Fatalf("GenerateArticle failed: %v", err)
	}
	if value.ID != 31 {
		t.Fatalf("ID = %d, want 31", value.ID)
	}
}

func TestGenerateArticle_ForbiddenForOtherUsersTranscription(t *testing.T) {
	transcriptionRepo := &transcriptionRepositoryStub{
		findTranscriptionByIDFunc: func(_ context.Context, id int64) (*transcriptionEntity.Transcription, error) {
			return &transcriptionEntity.Transcription{ID: id, UserID: 99}, nil
		},
	}

	uc := articleusecase.NewUseCase(&articleRepositoryStub{}, &promptRunJobRepositoryStub{}, transcriptionRepo).WithGenerator(&promptRepositoryStub{}, &generatorStub{})
	_, err := uc.GenerateArticle(context.Background(), articleusecase.GenerateArticleInput{
		UserID:          9,
		TranscriptionID: 5,
		PromptID:        6,
	})
	if !errors.Is(err, articleusecase.ErrForbidden) {
		t.Fatalf("err = %v, want ErrForbidden", err)
	}
}

func TestGenerateArticle_ForbiddenForInvisiblePrompt(t *testing.T) {
	ownerID := int64(88)
	transcriptionRepo := &transcriptionRepositoryStub{
		findTranscriptionByIDFunc: func(_ context.Context, id int64) (*transcriptionEntity.Transcription, error) {
			return &transcriptionEntity.Transcription{ID: id, UserID: 9}, nil
		},
	}
	promptRepo := &promptRepositoryStub{
		findPromptByIDFunc: func(_ context.Context, id int64) (*promptEntity.Prompt, error) {
			return &promptEntity.Prompt{ID: id, UserID: &ownerID, Name: "private", Body: "body", IsActive: true}, nil
		},
	}

	uc := articleusecase.NewUseCase(&articleRepositoryStub{}, &promptRunJobRepositoryStub{}, transcriptionRepo).WithGenerator(promptRepo, &generatorStub{})
	_, err := uc.GenerateArticle(context.Background(), articleusecase.GenerateArticleInput{
		UserID:          9,
		TranscriptionID: 5,
		PromptID:        6,
	})
	if !errors.Is(err, articleusecase.ErrForbidden) {
		t.Fatalf("err = %v, want ErrForbidden", err)
	}
}

func TestGenerateArticle_BlankGeneratedArticle(t *testing.T) {
	repo := &articleRepositoryStub{
		createArticleFunc: func(_ context.Context, value *entity.Article) (*entity.Article, error) {
			t.Fatalf("CreateArticle called unexpectedly: %#v", value)
			return nil, nil
		},
	}
	transcriptionRepo := &transcriptionRepositoryStub{
		findTranscriptionByIDFunc: func(_ context.Context, id int64) (*transcriptionEntity.Transcription, error) {
			return &transcriptionEntity.Transcription{ID: id, UserID: 9, FullText: "text"}, nil
		},
	}
	promptRepo := &promptRepositoryStub{
		findPromptByIDFunc: func(_ context.Context, id int64) (*promptEntity.Prompt, error) {
			return &promptEntity.Prompt{ID: id, Name: "blog", Body: "write", IsActive: true}, nil
		},
	}
	generator := &generatorStub{
		generateArticleFunc: func(_ context.Context, _ articlegen.Input) (*articlegen.GeneratedArticle, error) {
			return &articlegen.GeneratedArticle{Title: " ", Content: "generated"}, nil
		},
	}

	uc := articleusecase.NewUseCase(repo, &promptRunJobRepositoryStub{}, transcriptionRepo).WithGenerator(promptRepo, generator)
	_, err := uc.GenerateArticle(context.Background(), articleusecase.GenerateArticleInput{
		UserID:          9,
		TranscriptionID: 5,
		PromptID:        6,
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGenerateArticle_RateLimited(t *testing.T) {
	repo := &articleRepositoryStub{
		createArticleFunc: func(_ context.Context, value *entity.Article) (*entity.Article, error) {
			t.Fatalf("CreateArticle called unexpectedly: %#v", value)
			return nil, nil
		},
	}
	transcriptionRepo := &transcriptionRepositoryStub{
		findTranscriptionByIDFunc: func(_ context.Context, id int64) (*transcriptionEntity.Transcription, error) {
			return &transcriptionEntity.Transcription{ID: id, UserID: 9, FullText: "text"}, nil
		},
	}
	promptRepo := &promptRepositoryStub{
		findPromptByIDFunc: func(_ context.Context, id int64) (*promptEntity.Prompt, error) {
			return &promptEntity.Prompt{ID: id, Name: "blog", Body: "write", IsActive: true}, nil
		},
	}
	generator := &generatorStub{
		generateArticleFunc: func(_ context.Context, _ articlegen.Input) (*articlegen.GeneratedArticle, error) {
			return nil, articlegen.NewRateLimitError(35*time.Second, errors.New("quota exceeded"))
		},
	}

	uc := articleusecase.NewUseCase(repo, &promptRunJobRepositoryStub{}, transcriptionRepo).WithGenerator(promptRepo, generator)
	_, err := uc.GenerateArticle(context.Background(), articleusecase.GenerateArticleInput{
		UserID:          9,
		TranscriptionID: 5,
		PromptID:        6,
	})
	if err == nil {
		t.Fatal("expected error")
	}

	var appErr *apperr.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("err = %T, want *apperr.AppError", err)
	}
	if appErr.Status != 429 {
		t.Fatalf("status = %d, want 429", appErr.Status)
	}
	if appErr.Code != "rate_limited" {
		t.Fatalf("code = %q, want rate_limited", appErr.Code)
	}
	if appErr.Message != "article generation is temporarily rate limited; retry in 35s" {
		t.Fatalf("message = %q", appErr.Message)
	}
}

func TestCreateArticle_WithPromptRunJobID_UsesTransactionRunner(t *testing.T) {
	jobID := int64(21)
	txRunner := &txRunnerStub{}
	repo := &articleRepositoryStub{
		upsertArticleByPromptRunJobIDFunc: func(_ context.Context, value *entity.Article) (*entity.Article, error) {
			return value, nil
		},
	}
	promptRunJobRepo := &promptRunJobRepositoryStub{
		findPromptRunJobByIDFunc: func(_ context.Context, id int64) (*promptrunjobEntity.Job, error) {
			return &promptrunjobEntity.Job{ID: id, TranscriptionID: 33}, nil
		},
	}
	transcriptionRepo := &transcriptionRepositoryStub{
		findTranscriptionByIDFunc: func(_ context.Context, id int64) (*transcriptionEntity.Transcription, error) {
			return &transcriptionEntity.Transcription{ID: id, UserID: 9}, nil
		},
	}

	uc := articleusecase.NewUseCase(repo, promptRunJobRepo, transcriptionRepo, txRunner)
	_, err := uc.CreateArticle(context.Background(), articleusecase.CreateArticleInput{
		UserID:         9,
		PromptRunJobID: &jobID,
		Title:          "title",
		Content:        "content",
	})
	if err != nil {
		t.Fatalf("CreateArticle failed: %v", err)
	}
	if !txRunner.called {
		t.Fatal("transaction runner was not used")
	}
}

type articleRepositoryStub struct {
	createArticleFunc                 func(ctx context.Context, value *entity.Article) (*entity.Article, error)
	findArticleByIDFunc               func(ctx context.Context, id int64) (*entity.Article, error)
	findArticleByPromptRunJobIDFunc   func(ctx context.Context, promptRunJobID int64) (*entity.Article, error)
	listArticlesByUserIDFunc          func(ctx context.Context, userID int64) ([]*entity.Article, error)
	updateArticleFunc                 func(ctx context.Context, value *entity.Article) (*entity.Article, error)
	deleteArticleFunc                 func(ctx context.Context, id int64) error
	upsertArticleByPromptRunJobIDFunc func(ctx context.Context, value *entity.Article) (*entity.Article, error)
}

func (s *articleRepositoryStub) CreateArticle(ctx context.Context, value *entity.Article) (*entity.Article, error) {
	return s.createArticleFunc(ctx, value)
}

func (s *articleRepositoryStub) FindArticleByID(ctx context.Context, id int64) (*entity.Article, error) {
	return s.findArticleByIDFunc(ctx, id)
}

func (s *articleRepositoryStub) FindArticleByPromptRunJobID(ctx context.Context, promptRunJobID int64) (*entity.Article, error) {
	if s.findArticleByPromptRunJobIDFunc == nil {
		return nil, nil
	}
	return s.findArticleByPromptRunJobIDFunc(ctx, promptRunJobID)
}

func (s *articleRepositoryStub) ListArticlesByUserID(ctx context.Context, userID int64) ([]*entity.Article, error) {
	return s.listArticlesByUserIDFunc(ctx, userID)
}

func (s *articleRepositoryStub) UpdateArticle(ctx context.Context, value *entity.Article) (*entity.Article, error) {
	return s.updateArticleFunc(ctx, value)
}

func (s *articleRepositoryStub) DeleteArticle(ctx context.Context, id int64) error {
	return s.deleteArticleFunc(ctx, id)
}

func (s *articleRepositoryStub) UpsertArticleByPromptRunJobID(ctx context.Context, value *entity.Article) (*entity.Article, error) {
	return s.upsertArticleByPromptRunJobIDFunc(ctx, value)
}

type promptRunJobRepositoryStub struct {
	findPromptRunJobByIDFunc func(ctx context.Context, id int64) (*promptrunjobEntity.Job, error)
}

func (s *promptRunJobRepositoryStub) CreatePromptRunJob(_ context.Context, _ *promptrunjobEntity.Job) (*promptrunjobEntity.Job, error) {
	return nil, errors.New("not implemented")
}

func (s *promptRunJobRepositoryStub) FindPromptRunJobByID(ctx context.Context, id int64) (*promptrunjobEntity.Job, error) {
	if s.findPromptRunJobByIDFunc == nil {
		return nil, errors.New("findPromptRunJobByIDFunc is nil")
	}
	return s.findPromptRunJobByIDFunc(ctx, id)
}

func (s *promptRunJobRepositoryStub) UpdatePromptRunJob(_ context.Context, _ *promptrunjobEntity.Job) (*promptrunjobEntity.Job, error) {
	return nil, errors.New("not implemented")
}

type transcriptionRepositoryStub struct {
	findTranscriptionByIDFunc func(ctx context.Context, id int64) (*transcriptionEntity.Transcription, error)
}

func (s *transcriptionRepositoryStub) CreateTranscription(_ context.Context, _ *transcriptionEntity.Transcription) (*transcriptionEntity.Transcription, error) {
	return nil, errors.New("not implemented")
}

func (s *transcriptionRepositoryStub) ListByUserID(_ context.Context, _ int64) ([]*transcriptionEntity.Transcription, error) {
	return nil, errors.New("not implemented")
}

func (s *transcriptionRepositoryStub) FindTranscriptionByID(ctx context.Context, id int64) (*transcriptionEntity.Transcription, error) {
	if s.findTranscriptionByIDFunc == nil {
		return nil, errors.New("findTranscriptionByIDFunc is nil")
	}
	return s.findTranscriptionByIDFunc(ctx, id)
}

type txRunnerStub struct {
	called bool
}

func (s *txRunnerStub) RunInTx(ctx context.Context, fn func(context.Context) error) error {
	s.called = true
	return fn(ctx)
}

type promptRepositoryStub struct {
	findPromptByIDFunc func(ctx context.Context, id int64) (*promptEntity.Prompt, error)
}

func (s *promptRepositoryStub) ListVisiblePromptsByUserID(_ context.Context, _ int64) ([]*promptEntity.Prompt, error) {
	return nil, errors.New("not implemented")
}

func (s *promptRepositoryStub) CreatePrompt(_ context.Context, _ *promptEntity.Prompt) (*promptEntity.Prompt, error) {
	return nil, errors.New("not implemented")
}

func (s *promptRepositoryStub) FindPromptByID(ctx context.Context, id int64) (*promptEntity.Prompt, error) {
	if s.findPromptByIDFunc == nil {
		return nil, errors.New("findPromptByIDFunc is nil")
	}
	return s.findPromptByIDFunc(ctx, id)
}

func (s *promptRepositoryStub) UpdatePrompt(_ context.Context, _ *promptEntity.Prompt) (*promptEntity.Prompt, error) {
	return nil, errors.New("not implemented")
}

func (s *promptRepositoryStub) DeletePrompt(_ context.Context, _ int64) error {
	return errors.New("not implemented")
}

type generatorStub struct {
	generateArticleFunc func(ctx context.Context, input articlegen.Input) (*articlegen.GeneratedArticle, error)
}

func (s *generatorStub) GenerateArticle(ctx context.Context, input articlegen.Input) (*articlegen.GeneratedArticle, error) {
	if s.generateArticleFunc == nil {
		return nil, errors.New("generateArticleFunc is nil")
	}
	return s.generateArticleFunc(ctx, input)
}
