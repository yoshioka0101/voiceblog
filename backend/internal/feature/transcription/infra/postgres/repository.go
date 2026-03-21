package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	transcriptiondomain "github.com/yoshioka0101/voiceblog/backend/internal/feature/transcription/domain"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, params transcriptiondomain.CreateParams) (*transcriptiondomain.Transcription, error) {
	const query = `
		INSERT INTO transcriptions (user_id, full_text, segments_json)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, full_text, segments_json, created_at, updated_at
	`

	row := r.db.QueryRowContext(ctx, query, params.UserID, params.FullText, []byte(params.SegmentsJSON))

	transcription, err := scanTranscription(row.Scan)
	if err != nil {
		return nil, fmt.Errorf("insert transcription: %w", err)
	}

	return transcription, nil
}

func (r *Repository) ListByUserID(ctx context.Context, userID int64) ([]*transcriptiondomain.Transcription, error) {
	const query = `
		SELECT id, user_id, full_text, segments_json, created_at, updated_at
		FROM transcriptions
		WHERE user_id = $1
		ORDER BY created_at DESC, id DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list transcriptions: %w", err)
	}
	defer rows.Close()

	var transcriptions []*transcriptiondomain.Transcription
	for rows.Next() {
		transcription, err := scanTranscription(rows.Scan)
		if err != nil {
			return nil, fmt.Errorf("scan transcription: %w", err)
		}
		transcriptions = append(transcriptions, transcription)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate transcriptions: %w", err)
	}

	return transcriptions, nil
}

func (r *Repository) FindByID(ctx context.Context, id int64) (*transcriptiondomain.Transcription, error) {
	const query = `
		SELECT id, user_id, full_text, segments_json, created_at, updated_at
		FROM transcriptions
		WHERE id = $1
	`

	row := r.db.QueryRowContext(ctx, query, id)

	transcription, err := scanTranscription(row.Scan)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, transcriptiondomain.ErrNotFound
		}
		return nil, fmt.Errorf("find transcription: %w", err)
	}

	return transcription, nil
}

func scanTranscription(scan func(dest ...any) error) (*transcriptiondomain.Transcription, error) {
	var transcription transcriptiondomain.Transcription
	var segments []byte

	if err := scan(
		&transcription.ID,
		&transcription.UserID,
		&transcription.FullText,
		&segments,
		&transcription.CreatedAt,
		&transcription.UpdatedAt,
	); err != nil {
		return nil, err
	}

	transcription.SegmentsJSON = json.RawMessage(append([]byte(nil), segments...))

	return &transcription, nil
}
