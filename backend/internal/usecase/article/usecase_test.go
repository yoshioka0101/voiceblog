package usecase_test

import (
	"context"
	"errors"
	"testing"

	entity "github.com/yoshioka0101/voiceblog/backend/internal/entity/article"
	promptrunjobEntity "github.com/yoshioka0101/voiceblog/backend/internal/entity/promptrunjob"
	transcriptionEntity "github.com/yoshioka0101/voiceblog/backend/internal/entity/transcription"
	articleusecase "github.com/yoshioka0101/voiceblog/backend/internal/usecase/article"
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
		findByIDFunc: func(_ context.Context, id int64) (*transcriptionEntity.Transcription, error) {
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
	findByIDFunc func(ctx context.Context, id int64) (*transcriptionEntity.Transcription, error)
}

func (s *transcriptionRepositoryStub) Create(_ context.Context, _ *transcriptionEntity.Transcription) (*transcriptionEntity.Transcription, error) {
	return nil, errors.New("not implemented")
}

func (s *transcriptionRepositoryStub) ListByUserID(_ context.Context, _ int64) ([]*transcriptionEntity.Transcription, error) {
	return nil, errors.New("not implemented")
}

func (s *transcriptionRepositoryStub) FindByID(ctx context.Context, id int64) (*transcriptionEntity.Transcription, error) {
	if s.findByIDFunc == nil {
		return nil, errors.New("findByIDFunc is nil")
	}
	return s.findByIDFunc(ctx, id)
}
