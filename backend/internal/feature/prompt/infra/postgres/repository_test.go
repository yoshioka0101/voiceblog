package postgres_test

import (
	"context"
	"errors"
	"testing"

	promptdomain "github.com/yoshioka0101/voiceblog/backend/internal/feature/prompt/domain"
	pg "github.com/yoshioka0101/voiceblog/backend/internal/feature/prompt/infra/postgres"
	"github.com/yoshioka0101/voiceblog/backend/internal/testutil"
)

func TestCreate(t *testing.T) {
	db := testutil.SetupTestDB(t)
	userID := testutil.SeedUser(t, db, "google", "sub-create-prompt", "create-prompt@example.com", "Create Prompt User")
	repo := pg.NewRepository(db)

	prompt, err := repo.Create(context.Background(), promptdomain.CreateParams{
		UserID:   userID,
		Name:     "My Prompt",
		Body:     "Summarize this transcript.",
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if prompt.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
	if prompt.UserID == nil || *prompt.UserID != userID {
		t.Fatalf("UserID = %v, want %d", prompt.UserID, userID)
	}
	if prompt.Name != "My Prompt" {
		t.Fatalf("Name = %q", prompt.Name)
	}
	if prompt.Body != "Summarize this transcript." {
		t.Fatalf("Body = %q", prompt.Body)
	}
	if !prompt.IsActive {
		t.Fatal("expected IsActive = true")
	}
	if prompt.IsDefault {
		t.Fatal("expected IsDefault = false")
	}
}

func TestListVisibleByUserID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	targetUserID := testutil.SeedUser(t, db, "google", "sub-prompt-list-target", "prompt-target@example.com", "Prompt Target User")
	otherUserID := testutil.SeedUser(t, db, "google", "sub-prompt-list-other", "prompt-other@example.com", "Prompt Other User")
	repo := pg.NewRepository(db)

	testutil.SeedPrompt(t, db, &targetUserID, "target-active", "visible", true, false)
	testutil.SeedPrompt(t, db, &targetUserID, "target-inactive", "hidden", false, false)
	testutil.SeedPrompt(t, db, &otherUserID, "other-active", "hidden", true, false)

	prompts, err := repo.ListVisibleByUserID(context.Background(), targetUserID)
	if err != nil {
		t.Fatalf("ListVisibleByUserID failed: %v", err)
	}

	var foundSystem bool
	var foundTarget bool
	for _, prompt := range prompts {
		switch prompt.Name {
		case "ブログ記事作成（標準）":
			foundSystem = true
			if prompt.UserID != nil {
				t.Fatalf("system prompt UserID = %v, want nil", prompt.UserID)
			}
		case "target-active":
			foundTarget = true
			if prompt.UserID == nil || *prompt.UserID != targetUserID {
				t.Fatalf("target-active UserID = %v, want %d", prompt.UserID, targetUserID)
			}
		case "target-inactive", "other-active":
			t.Fatalf("unexpected prompt in result: %q", prompt.Name)
		}
	}

	if !foundSystem {
		t.Fatal("expected seeded system prompt in result")
	}
	if !foundTarget {
		t.Fatal("expected active user prompt in result")
	}
}

func TestUpdate(t *testing.T) {
	db := testutil.SetupTestDB(t)
	userID := testutil.SeedUser(t, db, "google", "sub-prompt-update", "prompt-update@example.com", "Prompt Update User")
	repo := pg.NewRepository(db)

	promptID := testutil.SeedPrompt(t, db, &userID, "old", "old body", true, false)

	updated, err := repo.Update(context.Background(), promptdomain.UpdateParams{
		ID:       promptID,
		Name:     "new",
		Body:     "new body",
		IsActive: false,
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
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

func TestDelete(t *testing.T) {
	db := testutil.SetupTestDB(t)
	userID := testutil.SeedUser(t, db, "google", "sub-prompt-delete", "prompt-delete@example.com", "Prompt Delete User")
	repo := pg.NewRepository(db)

	promptID := testutil.SeedPrompt(t, db, &userID, "delete-me", "body", true, false)

	if err := repo.Delete(context.Background(), promptID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err := repo.FindByID(context.Background(), promptID)
	if !errors.Is(err, promptdomain.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestFindByID_NotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := pg.NewRepository(db)

	_, err := repo.FindByID(context.Background(), 99999)
	if !errors.Is(err, promptdomain.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}
