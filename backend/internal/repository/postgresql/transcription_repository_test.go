package postgresql_test

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
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

	value, err := repo.CreateTranscription(context.Background(), &entity.Transcription{
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
	assertJSONEqual(t, value.SegmentsJSON, json.RawMessage(`[{"text":"hello world"}]`))
}

func TestFindTranscriptionByID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	userID := testutil.SeedUser(t, db, "google", "sub-find", "find@example.com", "Find User")
	repo := postgresql.NewTranscriptionRepository(db)
	ctx := context.Background()

	created, err := repo.CreateTranscription(ctx, &entity.Transcription{
		UserID:       userID,
		FullText:     "find me",
		SegmentsJSON: json.RawMessage(`[{"text":"find me"}]`),
	})
	if err != nil {
		t.Fatalf("CreateTranscription failed: %v", err)
	}

	found, err := repo.FindTranscriptionByID(ctx, created.ID)
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

	_, err := repo.FindTranscriptionByID(context.Background(), 99999)
	if !errors.Is(err, transcriptionusecase.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func assertJSONEqual(t *testing.T, got json.RawMessage, want json.RawMessage) {
	t.Helper()

	var gotValue any
	if err := json.Unmarshal(got, &gotValue); err != nil {
		t.Fatalf("failed to unmarshal got JSON: %v", err)
	}

	var wantValue any
	if err := json.Unmarshal(want, &wantValue); err != nil {
		t.Fatalf("failed to unmarshal want JSON: %v", err)
	}

	if !reflect.DeepEqual(gotValue, wantValue) {
		t.Fatalf("SegmentsJSON = %s, want %s", string(got), string(want))
	}
}
