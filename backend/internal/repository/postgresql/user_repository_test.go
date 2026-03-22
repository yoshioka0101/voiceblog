package postgresql_test

import (
	"context"
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
