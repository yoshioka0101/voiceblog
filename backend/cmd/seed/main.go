package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Fatal("DB_DSN is required")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	ctx := context.Background()
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	log.Println("seeding database...")

	userID := seedUser(ctx, db, "google", "seed-user-001", "seed@example.com", "Seed User")
	log.Printf("  created user: id=%d", userID)

	promptID := seedPrompt(ctx, db, nil, "デフォルトプロンプト", "以下の文字起こしをもとに技術ブログ記事を生成してください。\n\n{{transcription}}", true, true)
	log.Printf("  created system prompt: id=%d", promptID)

	userPromptID := seedPrompt(ctx, db, &userID, "カジュアル記事", "以下の文字起こしをカジュアルなブログ記事にしてください。\n\n{{transcription}}", true, false)
	log.Printf("  created user prompt: id=%d", userPromptID)

	transcriptionID := seedTranscription(ctx, db, userID,
		"今日はGoのクリーンアーキテクチャについて話します。まずエンティティ層から説明しましょう。",
		`[{"text":"今日はGoのクリーンアーキテクチャについて話します。","start_ms":0,"end_ms":3500},{"text":"まずエンティティ層から説明しましょう。","start_ms":3500,"end_ms":6000}]`,
	)
	log.Printf("  created transcription: id=%d", transcriptionID)

	articleID := seedArticle(ctx, db, userID, "手動で作成した記事", "これは手動で作成したテスト記事です。")
	log.Printf("  created article: id=%d", articleID)

	log.Println("seed completed successfully")
}

func seedUser(ctx context.Context, db *sql.DB, provider, subject, email, name string) int64 {
	var id int64
	err := db.QueryRowContext(ctx,
		`INSERT INTO users (auth_provider, auth_subject, email, name) VALUES ($1, $2, $3, $4) ON CONFLICT (auth_provider, auth_subject) DO UPDATE SET name = EXCLUDED.name RETURNING id`,
		provider, subject, email, name,
	).Scan(&id)
	if err != nil {
		log.Fatalf("failed to seed user: %v", err)
	}
	return id
}

func seedPrompt(ctx context.Context, db *sql.DB, userID *int64, name, body string, isActive, isDefault bool) int64 {
	var id int64
	var nullableUserID any
	if userID != nil {
		nullableUserID = *userID
	}
	err := db.QueryRowContext(ctx,
		`INSERT INTO prompts (user_id, name, body, is_active, is_default) VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		nullableUserID, name, body, isActive, isDefault,
	).Scan(&id)
	if err != nil {
		log.Fatalf("failed to seed prompt: %v", err)
	}
	return id
}

func seedTranscription(ctx context.Context, db *sql.DB, userID int64, fullText, segmentsJSON string) int64 {
	var id int64
	err := db.QueryRowContext(ctx,
		`INSERT INTO transcriptions (user_id, full_text, segments_json) VALUES ($1, $2, $3::jsonb) RETURNING id`,
		userID, fullText, segmentsJSON,
	).Scan(&id)
	if err != nil {
		log.Fatalf("failed to seed transcription: %v", err)
	}
	return id
}

func seedArticle(ctx context.Context, db *sql.DB, userID int64, title, content string) int64 {
	var id int64
	err := db.QueryRowContext(ctx,
		`INSERT INTO articles (user_id, title, content) VALUES ($1, $2, $3) RETURNING id`,
		userID, title, content,
	).Scan(&id)
	if err != nil {
		log.Fatalf("failed to seed article: %v", err)
	}
	return id
}

func init() {
	fmt.Println("VoiceBlog Database Seeder")
	fmt.Println("========================")
}
