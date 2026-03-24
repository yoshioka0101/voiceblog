package postgresql_test

import (
	"context"
	"errors"
	"testing"

	entity "github.com/yoshioka0101/voiceblog/backend/internal/entity/article"
	"github.com/yoshioka0101/voiceblog/backend/internal/repository/postgresql"
	"github.com/yoshioka0101/voiceblog/backend/internal/testutil"
	articleusecase "github.com/yoshioka0101/voiceblog/backend/internal/usecase/article"
)

func TestCreateArticle(t *testing.T) {
	db := testutil.SetupTestDB(t)
	userID := testutil.SeedUser(t, db, "google", "sub-article-create", "article-create@example.com", "Article Create User")
	repo := postgresql.NewArticleRepository(db)

	value, err := repo.CreateArticle(context.Background(), &entity.Article{
		UserID:  userID,
		Title:   "title",
		Content: "content",
	})
	if err != nil {
		t.Fatalf("CreateArticle failed: %v", err)
	}

	if value.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
	if value.UserID != userID {
		t.Fatalf("UserID = %d, want %d", value.UserID, userID)
	}
}

func TestListArticlesByUserID_ExcludesDeleted(t *testing.T) {
	db := testutil.SetupTestDB(t)
	userID := testutil.SeedUser(t, db, "google", "sub-article-list", "article-list@example.com", "Article List User")
	repo := postgresql.NewArticleRepository(db)

	visibleID := testutil.SeedArticle(t, db, userID, nil, "visible", "visible content")
	deletedID := testutil.SeedArticle(t, db, userID, nil, "deleted", "deleted content")

	if err := repo.DeleteArticle(context.Background(), deletedID); err != nil {
		t.Fatalf("DeleteArticle failed: %v", err)
	}

	values, err := repo.ListArticlesByUserID(context.Background(), userID)
	if err != nil {
		t.Fatalf("ListArticlesByUserID failed: %v", err)
	}

	if len(values) != 1 || values[0].ID != visibleID {
		t.Fatalf("articles = %#v", values)
	}
}

func TestUpsertArticleByPromptRunJobID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	userID := testutil.SeedUser(t, db, "google", "sub-article-upsert", "article-upsert@example.com", "Article Upsert User")
	transcriptionID := testutil.SeedTranscription(t, db, userID, "text", `[{"text":"text"}]`)
	promptID := testutil.SeedPrompt(t, db, nil, "system", "body", true, true)
	jobID := testutil.SeedPromptRunJob(t, db, transcriptionID, promptID, "completed", 1, nil, nil, nil)
	repo := postgresql.NewArticleRepository(db)

	first, err := repo.UpsertArticleByPromptRunJobID(context.Background(), &entity.Article{
		UserID:         userID,
		PromptRunJobID: &jobID,
		Title:          "first",
		Content:        "first content",
	})
	if err != nil {
		t.Fatalf("first upsert failed: %v", err)
	}

	second, err := repo.UpsertArticleByPromptRunJobID(context.Background(), &entity.Article{
		UserID:         userID,
		PromptRunJobID: &jobID,
		Title:          "second",
		Content:        "second content",
	})
	if err != nil {
		t.Fatalf("second upsert failed: %v", err)
	}

	if first.ID != second.ID {
		t.Fatalf("IDs = %d and %d, want same", first.ID, second.ID)
	}
	if second.Title != "second" {
		t.Fatalf("Title = %q", second.Title)
	}
}

func TestFindArticleByID_NotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := postgresql.NewArticleRepository(db)

	_, err := repo.FindArticleByID(context.Background(), 99999)
	if !errors.Is(err, articleusecase.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}
