package postgresql_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	entity "github.com/yoshioka0101/voiceblog/backend/internal/entity/transcription"
	"github.com/yoshioka0101/voiceblog/backend/internal/repository/postgresql"
	"github.com/yoshioka0101/voiceblog/backend/internal/testutil"
	transcriptionusecase "github.com/yoshioka0101/voiceblog/backend/internal/usecase/transcription"
)

func TestCreateTranscription(t *testing.T) {
	db := testutil.SetupTestDB(t)
	userID := testutil.SeedUser(t, db, "google", "sub-create", "create@example.com", "Create User")
	repo := postgresql.NewTranscriptionRepository(db)

	value, err := repo.Create(context.Background(), &entity.Transcription{
		UserID:       userID,
		FullText:     "hello world",
		SegmentsJSON: json.RawMessage(`[{"text":"hello world"}]`),
	})
	if err != nil {
		t.Fatalf("CreateTranscription failed: %v", err)
	}

	if value.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
	if value.UserID != userID {
		t.Fatalf("UserID = %d, want %d", value.UserID, userID)
	}
	if value.FullText != "hello world" {
		t.Fatalf("FullText = %q", value.FullText)
	}
	if string(value.SegmentsJSON) != `[{"text":"hello world"}]` {
		t.Fatalf("SegmentsJSON = %s", string(value.SegmentsJSON))
	}
}

func TestFindTranscriptionByID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	userID := testutil.SeedUser(t, db, "google", "sub-find", "find@example.com", "Find User")
	repo := postgresql.NewTranscriptionRepository(db)
	ctx := context.Background()

	created, err := repo.Create(ctx, &entity.Transcription{
		UserID:       userID,
		FullText:     "find me",
		SegmentsJSON: json.RawMessage(`[{"text":"find me"}]`),
	})
	if err != nil {
		t.Fatalf("CreateTranscription failed: %v", err)
	}

	found, err := repo.FindByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}

	if found.ID != created.ID {
		t.Fatalf("ID = %d, want %d", found.ID, created.ID)
	}
}

func TestFindTranscriptionByID_NotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := postgresql.NewTranscriptionRepository(db)

	_, err := repo.FindByID(context.Background(), 99999)
	if !errors.Is(err, transcriptionusecase.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}
