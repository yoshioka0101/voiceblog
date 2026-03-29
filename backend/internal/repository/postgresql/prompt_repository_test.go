package postgresql_test

import (
	"context"
	"errors"
	"testing"

	entity "github.com/yoshioka0101/voiceblog/backend/internal/entity/prompt"
	"github.com/yoshioka0101/voiceblog/backend/internal/repository/postgresql"
	"github.com/yoshioka0101/voiceblog/backend/internal/testutil"
	promptusecase "github.com/yoshioka0101/voiceblog/backend/internal/usecase/prompt"
)

func TestCreatePrompt(t *testing.T) {
	db := testutil.SetupTestDB(t)
	userID := testutil.SeedUser(t, db, "google", "sub-create-prompt", "create-prompt@example.com", "Create Prompt User")
	repo := postgresql.NewPromptRepository(db)

	promptValue, err := repo.CreatePrompt(context.Background(), &entity.Prompt{
		UserID:   &userID,
		Name:     "My Prompt",
		Body:     "Summarize this transcript.",
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("CreatePrompt failed: %v", err)
	}

	if promptValue.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
	if promptValue.UserID == nil || *promptValue.UserID != userID {
		t.Fatalf("UserID = %v, want %d", promptValue.UserID, userID)
	}
	if promptValue.Name != "My Prompt" {
		t.Fatalf("Name = %q", promptValue.Name)
	}
	if promptValue.Body != "Summarize this transcript." {
		t.Fatalf("Body = %q", promptValue.Body)
	}
	if !promptValue.IsActive {
		t.Fatal("expected IsActive = true")
	}
	if promptValue.IsDefault {
		t.Fatal("expected IsDefault = false")
	}
}

func TestListVisiblePromptsByUserID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	targetUserID := testutil.SeedUser(t, db, "google", "sub-prompt-list-target", "prompt-target@example.com", "Prompt Target User")
	otherUserID := testutil.SeedUser(t, db, "google", "sub-prompt-list-other", "prompt-other@example.com", "Prompt Other User")
	repo := postgresql.NewPromptRepository(db)

	testutil.SeedPrompt(t, db, &targetUserID, "target-active", "visible", true, false)
	testutil.SeedPrompt(t, db, &targetUserID, "target-inactive", "hidden", false, false)
	testutil.SeedPrompt(t, db, &otherUserID, "other-active", "hidden", true, false)

	prompts, err := repo.ListVisiblePromptsByUserID(context.Background(), targetUserID)
	if err != nil {
		t.Fatalf("ListVisibleByUserID failed: %v", err)
	}

	var foundSystem bool
	var foundTarget bool
	for _, value := range prompts {
		switch value.Name {
		case "ブログ記事作成（標準）":
			foundSystem = true
			if value.UserID != nil {
				t.Fatalf("system prompt UserID = %v, want nil", value.UserID)
			}
		case "target-active":
			foundTarget = true
			if value.UserID == nil || *value.UserID != targetUserID {
				t.Fatalf("target-active UserID = %v, want %d", value.UserID, targetUserID)
			}
		case "target-inactive", "other-active":
			t.Fatalf("unexpected prompt in result: %q", value.Name)
		}
	}

	if !foundSystem {
		t.Fatal("expected seeded system prompt in result")
	}
	if !foundTarget {
		t.Fatal("expected active user prompt in result")
	}
}

func TestUpdatePrompt(t *testing.T) {
	db := testutil.SetupTestDB(t)
	userID := testutil.SeedUser(t, db, "google", "sub-prompt-update", "prompt-update@example.com", "Prompt Update User")
	repo := postgresql.NewPromptRepository(db)

	promptID := testutil.SeedPrompt(t, db, &userID, "old", "old body", true, false)

	updated, err := repo.UpdatePrompt(context.Background(), &entity.Prompt{
		ID:       promptID,
		UserID:   &userID,
		Name:     "new",
		Body:     "new body",
		IsActive: false,
	})
	if err != nil {
		t.Fatalf("UpdatePrompt failed: %v", err)
	}

	if updated.Name != "new" {
		t.Fatalf("Name = %q", updated.Name)
	}
	if updated.Body != "new body" {
		t.Fatalf("Body = %q", updated.Body)
	}
	if updated.IsActive {
		t.Fatal("expected IsActive = false")
	}
}

func TestDeletePrompt(t *testing.T) {
	db := testutil.SetupTestDB(t)
	userID := testutil.SeedUser(t, db, "google", "sub-prompt-delete", "prompt-delete@example.com", "Prompt Delete User")
	repo := postgresql.NewPromptRepository(db)

	promptID := testutil.SeedPrompt(t, db, &userID, "delete-me", "body", true, false)

	if err := repo.DeletePrompt(context.Background(), promptID); err != nil {
		t.Fatalf("DeletePrompt failed: %v", err)
	}

	_, err := repo.FindPromptByID(context.Background(), promptID)
	if !errors.Is(err, promptusecase.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestFindPromptByID_NotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := postgresql.NewPromptRepository(db)

	_, err := repo.FindPromptByID(context.Background(), 99999)
	if !errors.Is(err, promptusecase.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}
