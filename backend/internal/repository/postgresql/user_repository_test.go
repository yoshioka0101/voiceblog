package postgresql_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/yoshioka0101/voiceblog/backend/internal/repository/postgresql"
	"github.com/yoshioka0101/voiceblog/backend/internal/testutil"
)

func TestFindOrCreate_NewUser(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := postgresql.NewUserRepository(db)
	ctx := context.Background()

	user, err := repo.FindOrCreate(ctx, "google", "sub-001", "alice@example.com", "Alice")
	if err != nil {
		t.Fatalf("FindOrCreate failed: %v", err)
	}

	if user.ID == 0 {
		t.Error("expected non-zero ID")
	}
	if user.AuthProvider != "google" {
		t.Errorf("AuthProvider = %q, want %q", user.AuthProvider, "google")
	}
	if user.AuthSubject != "sub-001" {
		t.Errorf("AuthSubject = %q, want %q", user.AuthSubject, "sub-001")
	}
	if user.Email != "alice@example.com" {
		t.Errorf("Email = %q, want %q", user.Email, "alice@example.com")
	}
	if user.Name != "Alice" {
		t.Errorf("Name = %q, want %q", user.Name, "Alice")
	}
}

func TestFindOrCreate_ExistingUser_Updates(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := postgresql.NewUserRepository(db)
	ctx := context.Background()

	// 1回目: 新規作成
	first, err := repo.FindOrCreate(ctx, "google", "sub-002", "bob@example.com", "Bob")
	if err != nil {
		t.Fatalf("first FindOrCreate failed: %v", err)
	}

	// 2回目: 同じ provider+subject で email と name を変更
	second, err := repo.FindOrCreate(ctx, "google", "sub-002", "bob-new@example.com", "Bob Updated")
	if err != nil {
		t.Fatalf("second FindOrCreate failed: %v", err)
	}

	// ID は同じであること（UPDATE されたのであって INSERT ではない）
	if first.ID != second.ID {
		t.Errorf("ID changed: first=%d, second=%d", first.ID, second.ID)
	}
	if second.Email != "bob-new@example.com" {
		t.Errorf("Email = %q, want %q", second.Email, "bob-new@example.com")
	}
	if second.Name != "Bob Updated" {
		t.Errorf("Name = %q, want %q", second.Name, "Bob Updated")
	}
}

func TestFindOrCreate_DifferentProviders(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := postgresql.NewUserRepository(db)
	ctx := context.Background()

	// 同じ subject でも provider が異なれば別ユーザー
	google, err := repo.FindOrCreate(ctx, "google", "sub-003", "carol@example.com", "Carol G")
	if err != nil {
		t.Fatalf("google FindOrCreate failed: %v", err)
	}

	apple, err := repo.FindOrCreate(ctx, "apple", "sub-003", "carol@example.com", "Carol A")
	if err != nil {
		t.Fatalf("apple FindOrCreate failed: %v", err)
	}

	if google.ID == apple.ID {
		t.Error("expected different IDs for different providers")
	}
	if google.AuthProvider != "google" {
		t.Errorf("google.AuthProvider = %q", google.AuthProvider)
	}
	if apple.AuthProvider != "apple" {
		t.Errorf("apple.AuthProvider = %q", apple.AuthProvider)
	}
}

func TestDeleteWithRelatedData_DeletesOnlyTargetUser(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := postgresql.NewUserRepository(db)
	ctx := context.Background()

	target, err := repo.FindOrCreate(ctx, "device", "subject-delete", "", "")
	if err != nil {
		t.Fatalf("create target user: %v", err)
	}
	other, err := repo.FindOrCreate(ctx, "device", "subject-keep", "", "")
	if err != nil {
		t.Fatalf("create other user: %v", err)
	}

	for _, uid := range []int64{target.ID, other.ID} {
		testutil.SeedTranscription(t, db, uid, "text", "[]")
		testutil.SeedPrompt(t, db, &uid, "p", "body", true, false)
		articleID := testutil.SeedArticle(t, db, uid, "t", "c")
		mustExec(t, db, `INSERT INTO external_tokens (user_id, provider, encrypted_token, nonce) VALUES ($1, 'qiita', '\x00', '\x00')`, uid)
		mustExec(t, db, `INSERT INTO article_share_targets (article_id, provider) VALUES ($1, 'qiita')`, articleID)
	}

	if err := repo.DeleteWithRelatedData(ctx, target.ID); err != nil {
		t.Fatalf("DeleteWithRelatedData failed: %v", err)
	}

	counts := map[string]string{
		"users":                 `SELECT COUNT(*) FROM users WHERE id = $1`,
		"transcriptions":        `SELECT COUNT(*) FROM transcriptions WHERE user_id = $1`,
		"prompts":               `SELECT COUNT(*) FROM prompts WHERE user_id = $1`,
		"external_tokens":       `SELECT COUNT(*) FROM external_tokens WHERE user_id = $1`,
		"articles":              `SELECT COUNT(*) FROM articles WHERE user_id = $1`,
		"article_share_targets": `SELECT COUNT(*) FROM article_share_targets s JOIN articles a ON a.id = s.article_id WHERE a.user_id = $1`,
	}
	for table, query := range counts {
		if got := countRows(t, db, query, target.ID); got != 0 {
			t.Errorf("%s: target user rows = %d, want 0", table, got)
		}
		if got := countRows(t, db, query, other.ID); got != 1 {
			t.Errorf("%s: other user rows = %d, want 1", table, got)
		}
	}

	if got := countRows(t, db, `SELECT COUNT(*) FROM prompts WHERE user_id IS NULL`); got == 0 {
		t.Error("shared prompts (user_id IS NULL) must not be deleted")
	}
}

func mustExec(t *testing.T, db *sql.DB, query string, args ...any) {
	t.Helper()
	if _, err := db.ExecContext(context.Background(), query, args...); err != nil {
		t.Fatalf("exec %q: %v", query, err)
	}
}

func countRows(t *testing.T, db *sql.DB, query string, args ...any) int {
	t.Helper()
	var n int
	if err := db.QueryRowContext(context.Background(), query, args...).Scan(&n); err != nil {
		t.Fatalf("count %q: %v", query, err)
	}
	return n
}
