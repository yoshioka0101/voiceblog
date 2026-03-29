package postgresql_test

import (
	"context"
	"errors"
	"testing"
	"time"

	entity "github.com/yoshioka0101/voiceblog/backend/internal/entity/promptrunjob"
	"github.com/yoshioka0101/voiceblog/backend/internal/repository/postgresql"
	"github.com/yoshioka0101/voiceblog/backend/internal/testutil"
	promptrunjobusecase "github.com/yoshioka0101/voiceblog/backend/internal/usecase/promptrunjob"
)

func TestCreatePromptRunJob(t *testing.T) {
	db := testutil.SetupTestDB(t)
	userID := testutil.SeedUser(t, db, "google", "sub-job-create", "job-create@example.com", "Job Create User")
	transcriptionID := testutil.SeedTranscription(t, db, userID, "text", `[{"text":"text"}]`)
	promptID := testutil.SeedPrompt(t, db, nil, "system", "body", true, true)
	repo := postgresql.NewPromptRunJobRepository(db)

	value, err := repo.CreatePromptRunJob(context.Background(), &entity.Job{
		TranscriptionID:  transcriptionID,
		PromptID:         promptID,
		Status:           entity.StatusPending,
		AttemptCount:     0,
		NextRunAt:        time.Now(),
		GeneratedTitle:   pointer("preview title"),
		GeneratedContent: pointer("preview content"),
	})
	if err != nil {
		t.Fatalf("CreatePromptRunJob failed: %v", err)
	}

	if value.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
	if value.Status != entity.StatusPending {
		t.Fatalf("Status = %q", value.Status)
	}
	if value.GeneratedTitle == nil || *value.GeneratedTitle != "preview title" {
		t.Fatalf("GeneratedTitle = %#v", value.GeneratedTitle)
	}
}

func TestUpdatePromptRunJob(t *testing.T) {
	db := testutil.SetupTestDB(t)
	userID := testutil.SeedUser(t, db, "google", "sub-job-update", "job-update@example.com", "Job Update User")
	transcriptionID := testutil.SeedTranscription(t, db, userID, "text", `[{"text":"text"}]`)
	promptID := testutil.SeedPrompt(t, db, nil, "system", "body", true, true)
	jobID := testutil.SeedPromptRunJob(t, db, transcriptionID, promptID, entity.StatusPending, 0, nil, nil, nil)
	repo := postgresql.NewPromptRunJobRepository(db)

	value, err := repo.UpdatePromptRunJob(context.Background(), &entity.Job{
		ID:               jobID,
		TranscriptionID:  transcriptionID,
		PromptID:         promptID,
		Status:           entity.StatusCompleted,
		AttemptCount:     1,
		NextRunAt:        time.Now(),
		GeneratedTitle:   pointer("preview title"),
		GeneratedContent: pointer("preview content"),
	})
	if err != nil {
		t.Fatalf("UpdatePromptRunJob failed: %v", err)
	}

	if value.Status != entity.StatusCompleted {
		t.Fatalf("Status = %q", value.Status)
	}
	if value.AttemptCount != 1 {
		t.Fatalf("AttemptCount = %d", value.AttemptCount)
	}
	if value.GeneratedContent == nil || *value.GeneratedContent != "preview content" {
		t.Fatalf("GeneratedContent = %#v", value.GeneratedContent)
	}
}

func TestFindPromptRunJobByID_NotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := postgresql.NewPromptRunJobRepository(db)

	_, err := repo.FindPromptRunJobByID(context.Background(), 99999)
	if !errors.Is(err, promptrunjobusecase.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func pointer(value string) *string {
	return &value
}
