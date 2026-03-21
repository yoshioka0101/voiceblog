package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	promptdomain "github.com/yoshioka0101/voiceblog/backend/internal/feature/prompt/domain"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListVisibleByUserID(ctx context.Context, userID int64) ([]*promptdomain.Prompt, error) {
	const query = `
		SELECT id, user_id, name, body, is_active, is_default, created_at, updated_at
		FROM prompts
		WHERE is_active = TRUE
		  AND (user_id IS NULL OR user_id = $1)
		ORDER BY is_default DESC, CASE WHEN user_id IS NULL THEN 0 ELSE 1 END, created_at DESC, id DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list prompts: %w", err)
	}
	defer rows.Close()

	var prompts []*promptdomain.Prompt
	for rows.Next() {
		prompt, err := scanPrompt(rows.Scan)
		if err != nil {
			return nil, fmt.Errorf("scan prompt: %w", err)
		}
		prompts = append(prompts, prompt)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate prompts: %w", err)
	}

	return prompts, nil
}

func (r *Repository) Create(ctx context.Context, params promptdomain.CreateParams) (*promptdomain.Prompt, error) {
	const query = `
		INSERT INTO prompts (user_id, name, body, is_active, is_default)
		VALUES ($1, $2, $3, $4, FALSE)
		RETURNING id, user_id, name, body, is_active, is_default, created_at, updated_at
	`

	row := r.db.QueryRowContext(ctx, query, params.UserID, params.Name, params.Body, params.IsActive)

	prompt, err := scanPrompt(row.Scan)
	if err != nil {
		return nil, fmt.Errorf("insert prompt: %w", err)
	}

	return prompt, nil
}

func (r *Repository) FindByID(ctx context.Context, id int64) (*promptdomain.Prompt, error) {
	const query = `
		SELECT id, user_id, name, body, is_active, is_default, created_at, updated_at
		FROM prompts
		WHERE id = $1
	`

	row := r.db.QueryRowContext(ctx, query, id)

	prompt, err := scanPrompt(row.Scan)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, promptdomain.ErrNotFound
		}
		return nil, fmt.Errorf("find prompt: %w", err)
	}

	return prompt, nil
}

func (r *Repository) Update(ctx context.Context, params promptdomain.UpdateParams) (*promptdomain.Prompt, error) {
	const query = `
		UPDATE prompts
		SET name = $2,
		    body = $3,
		    is_active = $4,
		    updated_at = NOW()
		WHERE id = $1
		RETURNING id, user_id, name, body, is_active, is_default, created_at, updated_at
	`

	row := r.db.QueryRowContext(ctx, query, params.ID, params.Name, params.Body, params.IsActive)

	prompt, err := scanPrompt(row.Scan)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, promptdomain.ErrNotFound
		}
		return nil, fmt.Errorf("update prompt: %w", err)
	}

	return prompt, nil
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	const query = `
		DELETE FROM prompts
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete prompt: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete prompt rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return promptdomain.ErrNotFound
	}

	return nil
}

func scanPrompt(scan func(dest ...any) error) (*promptdomain.Prompt, error) {
	var prompt promptdomain.Prompt
	var userID sql.NullInt64

	if err := scan(
		&prompt.ID,
		&userID,
		&prompt.Name,
		&prompt.Body,
		&prompt.IsActive,
		&prompt.IsDefault,
		&prompt.CreatedAt,
		&prompt.UpdatedAt,
	); err != nil {
		return nil, err
	}

	if userID.Valid {
		prompt.UserID = &userID.Int64
	}

	return &prompt, nil
}
