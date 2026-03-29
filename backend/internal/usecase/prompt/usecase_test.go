package usecase_test

import (
	"context"
	"errors"
	"testing"

	entity "github.com/yoshioka0101/voiceblog/backend/internal/entity/prompt"
	promptusecase "github.com/yoshioka0101/voiceblog/backend/internal/usecase/prompt"
)

func TestCreatePrompt(t *testing.T) {
	repo := &promptRepositoryStub{
		createPromptFunc: func(_ context.Context, prompt *entity.Prompt) (*entity.Prompt, error) {
			if prompt.UserID == nil || *prompt.UserID != 11 {
				t.Fatalf("UserID = %v, want 11", prompt.UserID)
			}
			if prompt.Name != "draft" {
				t.Fatalf("Name = %q", prompt.Name)
			}
			return &entity.Prompt{
				ID:       21,
				UserID:   prompt.UserID,
				Name:     prompt.Name,
				Body:     prompt.Body,
				IsActive: prompt.IsActive,
			}, nil
		},
	}

	uc := promptusecase.NewUseCase(repo)
	isActive := true
	created, err := uc.CreatePrompt(context.Background(), promptusecase.CreatePromptInput{
		UserID:   11,
		Name:     "draft",
		Body:     "body",
		IsActive: &isActive,
	})
	if err != nil {
		t.Fatalf("CreatePrompt failed: %v", err)
	}
	if created.ID != 21 {
		t.Fatalf("ID = %d, want 21", created.ID)
	}
}

func TestListVisiblePromptsByUserID(t *testing.T) {
	expected := []*entity.Prompt{{ID: 1, Name: "visible"}}
	repo := &promptRepositoryStub{
		listVisiblePromptsByUserIDFunc: func(_ context.Context, userID int64) ([]*entity.Prompt, error) {
			if userID != 42 {
				t.Fatalf("userID = %d, want 42", userID)
			}
			return expected, nil
		},
	}

	uc := promptusecase.NewUseCase(repo)
	prompts, err := uc.ListVisiblePromptsByUserID(context.Background(), 42)
	if err != nil {
		t.Fatalf("ListVisiblePromptsByUserID failed: %v", err)
	}
	if len(prompts) != 1 || prompts[0].ID != expected[0].ID {
		t.Fatalf("prompts = %#v", prompts)
	}
}

func TestUpdatePrompt_ForbiddenForSystemPrompt(t *testing.T) {
	repo := &promptRepositoryStub{
		findPromptByIDFunc: func(_ context.Context, id int64) (*entity.Prompt, error) {
			if id != 9 {
				t.Fatalf("id = %d, want 9", id)
			}
			return &entity.Prompt{ID: id, UserID: nil, Name: "system"}, nil
		},
	}

	uc := promptusecase.NewUseCase(repo)
	updatedName := "updated"
	_, err := uc.UpdatePrompt(context.Background(), promptusecase.UpdatePromptInput{
		UserID:   5,
		PromptID: 9,
		Name:     &updatedName,
	})
	if !errors.Is(err, promptusecase.ErrForbidden) {
		t.Fatalf("err = %v, want ErrForbidden", err)
	}
}

func TestDeletePrompt_DeletesOwnedPrompt(t *testing.T) {
	userID := int64(7)
	deletedID := int64(0)
	repo := &promptRepositoryStub{
		findPromptByIDFunc: func(_ context.Context, id int64) (*entity.Prompt, error) {
			return &entity.Prompt{ID: id, UserID: &userID}, nil
		},
		deletePromptFunc: func(_ context.Context, id int64) error {
			deletedID = id
			return nil
		},
	}

	uc := promptusecase.NewUseCase(repo)
	if err := uc.DeletePrompt(context.Background(), userID, 15); err != nil {
		t.Fatalf("DeletePrompt failed: %v", err)
	}
	if deletedID != 15 {
		t.Fatalf("deletedID = %d, want 15", deletedID)
	}
}

func TestUpdatePrompt_UsesTransactionRunner(t *testing.T) {
	userID := int64(5)
	updated := false
	repo := &promptRepositoryStub{
		findPromptByIDFunc: func(_ context.Context, id int64) (*entity.Prompt, error) {
			return &entity.Prompt{ID: id, UserID: &userID, Name: "before", Body: "body", IsActive: true}, nil
		},
		updatePromptFunc: func(_ context.Context, prompt *entity.Prompt) (*entity.Prompt, error) {
			updated = true
			return prompt, nil
		},
	}
	txRunner := &txRunnerStub{}

	uc := promptusecase.NewUseCase(repo, txRunner)
	updatedName := "after"
	_, err := uc.UpdatePrompt(context.Background(), promptusecase.UpdatePromptInput{
		UserID:   userID,
		PromptID: 9,
		Name:     &updatedName,
	})
	if err != nil {
		t.Fatalf("UpdatePrompt failed: %v", err)
	}
	if !txRunner.called {
		t.Fatal("transaction runner was not used")
	}
	if !updated {
		t.Fatal("update was not called")
	}
}

type promptRepositoryStub struct {
	listVisiblePromptsByUserIDFunc func(ctx context.Context, userID int64) ([]*entity.Prompt, error)
	createPromptFunc               func(ctx context.Context, prompt *entity.Prompt) (*entity.Prompt, error)
	findPromptByIDFunc             func(ctx context.Context, id int64) (*entity.Prompt, error)
	updatePromptFunc               func(ctx context.Context, prompt *entity.Prompt) (*entity.Prompt, error)
	deletePromptFunc               func(ctx context.Context, id int64) error
}

func (s *promptRepositoryStub) ListVisiblePromptsByUserID(ctx context.Context, userID int64) ([]*entity.Prompt, error) {
	return s.listVisiblePromptsByUserIDFunc(ctx, userID)
}

func (s *promptRepositoryStub) CreatePrompt(ctx context.Context, prompt *entity.Prompt) (*entity.Prompt, error) {
	return s.createPromptFunc(ctx, prompt)
}

func (s *promptRepositoryStub) FindPromptByID(ctx context.Context, id int64) (*entity.Prompt, error) {
	return s.findPromptByIDFunc(ctx, id)
}

func (s *promptRepositoryStub) UpdatePrompt(ctx context.Context, prompt *entity.Prompt) (*entity.Prompt, error) {
	return s.updatePromptFunc(ctx, prompt)
}

func (s *promptRepositoryStub) DeletePrompt(ctx context.Context, id int64) error {
	return s.deletePromptFunc(ctx, id)
}

type txRunnerStub struct {
	called bool
}

func (s *txRunnerStub) RunInTx(ctx context.Context, fn func(context.Context) error) error {
	s.called = true
	return fn(ctx)
}
