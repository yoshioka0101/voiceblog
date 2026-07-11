package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"runtime"
	"sort"
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
	testcontainers.SkipIfProviderIsNotHealthy(t)

	ctx := context.Background()

	migrationPaths, err := filepath.Glob(filepath.Join(projectRoot(), "migrations", "migrations", "*.sql"))
	if err != nil {
		t.Fatalf("failed to list migrations: %v", err)
	}
	if len(migrationPaths) == 0 {
		t.Fatal("no migrations found")
	}
	sort.Strings(migrationPaths)

	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("voiceblog_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		postgres.WithInitScripts(migrationPaths...),
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
		_ = db.Close()
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

func SeedUser(t *testing.T, db *sql.DB, provider, subject, email, name string) int64 {
	t.Helper()

	const query = `
		INSERT INTO users (auth_provider, auth_subject, email, name)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	var id int64
	if err := db.QueryRowContext(context.Background(), query, provider, subject, email, name).Scan(&id); err != nil {
		t.Fatalf("failed to seed user: %v", err)
	}

	return id
}

func SeedPrompt(t *testing.T, db *sql.DB, userID *int64, name, body string, isActive, isDefault bool) int64 {
	t.Helper()

	const query = `
		INSERT INTO prompts (user_id, name, body, is_active, is_default)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	var id int64
	var nullableUserID any
	if userID != nil {
		nullableUserID = *userID
	}

	if err := db.QueryRowContext(context.Background(), query, nullableUserID, name, body, isActive, isDefault).Scan(&id); err != nil {
		t.Fatalf("failed to seed prompt: %v", err)
	}

	return id
}

func SeedTranscription(t *testing.T, db *sql.DB, userID int64, fullText, segmentsJSON string) int64 {
	t.Helper()

	const query = `
		INSERT INTO transcriptions (user_id, full_text, segments_json)
		VALUES ($1, $2, $3::jsonb)
		RETURNING id
	`

	var id int64
	if err := db.QueryRowContext(context.Background(), query, userID, fullText, segmentsJSON).Scan(&id); err != nil {
		t.Fatalf("failed to seed transcription: %v", err)
	}

	return id
}

func SeedArticle(t *testing.T, db *sql.DB, userID int64, title, content string) int64 {
	t.Helper()

	const query = `
		INSERT INTO articles (user_id, title, content)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	var id int64
	if err := db.QueryRowContext(context.Background(), query, userID, title, content).Scan(&id); err != nil {
		t.Fatalf("failed to seed article: %v", err)
	}

	return id
}

func MustFindMigrationPath(name string) string {
	return filepath.Join(projectRoot(), "migrations", "migrations", name)
}

func FormatDSN(host string, port int, dbName string) string {
	return fmt.Sprintf("postgres://test:test@%s:%d/%s?sslmode=disable", host, port, dbName)
}
