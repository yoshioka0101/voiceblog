package testutil

import (
	"context"
	"database/sql"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// SetupTestDB は testcontainers-go で PostgreSQL 16 コンテナを起動し、
// マイグレーション SQL を適用して *sql.DB を返す。
// テスト終了時にコンテナは自動破棄される。
func SetupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	ctx := context.Background()

	migrationPath := filepath.Join(projectRoot(), "migrations", "migrations", "20260320_create_users.sql")

	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("voiceblog_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		postgres.WithInitScripts(migrationPath),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	t.Cleanup(func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate postgres container: %v", err)
		}
	})

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	// コンテナ起動直後はポートフォワーディングが安定しない場合があるためリトライ
	for i := range 10 {
		if err := db.PingContext(ctx); err == nil {
			break
		}
		if i == 9 {
			t.Fatalf("failed to ping database after retries: connStr=%s", connStr)
		}
		time.Sleep(500 * time.Millisecond)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

// projectRoot はプロジェクトルート（voiceblog/）のパスを返す。
func projectRoot() string {
	_, filename, _, _ := runtime.Caller(0)
	// filename = .../voiceblog/backend/internal/testutil/db.go
	// backend の親がプロジェクトルート
	return filepath.Join(filepath.Dir(filename), "..", "..", "..")
}
