package postgres_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	transcriptiondomain "github.com/yoshioka0101/voiceblog/backend/internal/feature/transcription/domain"
	pg "github.com/yoshioka0101/voiceblog/backend/internal/feature/transcription/infra/postgres"
	"github.com/yoshioka0101/voiceblog/backend/internal/testutil"
)

func TestCreate(t *testing.T) {
	db := testutil.SetupTestDB(t)
	userID := testutil.SeedUser(t, db, "google", "sub-create", "create@example.com", "Create User")
	repo := pg.NewRepository(db)

	transcription, err := repo.Create(context.Background(), transcriptiondomain.CreateParams{
		UserID:       userID,
		FullText:     "hello world",
		SegmentsJSON: json.RawMessage(`[{"text":"hello world"}]`),
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if transcription.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
	if transcription.UserID != userID {
		t.Fatalf("UserID = %d, want %d", transcription.UserID, userID)
	}
	if transcription.FullText != "hello world" {
		t.Fatalf("FullText = %q", transcription.FullText)
	}
	if string(transcription.SegmentsJSON) != `[{"text":"hello world"}]` {
		t.Fatalf("SegmentsJSON = %s", string(transcription.SegmentsJSON))
	}
}

func TestListByUserID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	targetUserID := testutil.SeedUser(t, db, "google", "sub-list-target", "target@example.com", "Target User")
	otherUserID := testutil.SeedUser(t, db, "google", "sub-list-other", "other@example.com", "Other User")
	repo := pg.NewRepository(db)
	ctx := context.Background()

	if _, err := repo.Create(ctx, transcriptiondomain.CreateParams{
		UserID:       targetUserID,
		FullText:     "target-1",
		SegmentsJSON: json.RawMessage(`[{"text":"target-1"}]`),
	}); err != nil {
		t.Fatalf("Create target-1 failed: %v", err)
	}
	if _, err := repo.Create(ctx, transcriptiondomain.CreateParams{
		UserID:       targetUserID,
		FullText:     "target-2",
		SegmentsJSON: json.RawMessage(`[{"text":"target-2"}]`),
	}); err != nil {
		t.Fatalf("Create target-2 failed: %v", err)
	}
	if _, err := repo.Create(ctx, transcriptiondomain.CreateParams{
		UserID:       otherUserID,
		FullText:     "other",
		SegmentsJSON: json.RawMessage(`[{"text":"other"}]`),
	}); err != nil {
		t.Fatalf("Create other failed: %v", err)
	}

	transcriptions, err := repo.ListByUserID(ctx, targetUserID)
	if err != nil {
		t.Fatalf("ListByUserID failed: %v", err)
	}

	if len(transcriptions) != 2 {
		t.Fatalf("len(transcriptions) = %d, want 2", len(transcriptions))
	}
	for _, transcription := range transcriptions {
		if transcription.UserID != targetUserID {
			t.Fatalf("unexpected UserID = %d", transcription.UserID)
		}
	}
}

func TestFindByID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	userID := testutil.SeedUser(t, db, "google", "sub-find", "find@example.com", "Find User")
	repo := pg.NewRepository(db)
	ctx := context.Background()

	created, err := repo.Create(ctx, transcriptiondomain.CreateParams{
		UserID:       userID,
		FullText:     "find me",
		SegmentsJSON: json.RawMessage(`[{"text":"find me"}]`),
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	found, err := repo.FindByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}

	if found.ID != created.ID {
		t.Fatalf("ID = %d, want %d", found.ID, created.ID)
	}
}

func TestFindByID_NotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := pg.NewRepository(db)

	_, err := repo.FindByID(context.Background(), 99999)
	if !errors.Is(err, transcriptiondomain.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}
