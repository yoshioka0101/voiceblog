package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/im"

	userdomain "github.com/yoshioka0101/voiceblog/backend/internal/feature/user/domain"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindOrCreate(ctx context.Context, provider, subject, email, name string) (*userdomain.User, error) {
	q := psql.Insert(
		im.Into("users", "auth_provider", "auth_subject", "email", "name", "updated_at"),
		im.Values(psql.Arg(provider, subject, email, name), psql.Raw("NOW()")),
		im.OnConflict("auth_provider", "auth_subject").DoUpdate(
			im.SetExcluded("email", "name", "updated_at"),
		),
		im.Returning("id", "auth_provider", "auth_subject", "email", "name"),
	)

	queryStr, args, err := q.Build(ctx)
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	u := &userdomain.User{}
	row := r.db.QueryRowContext(ctx, queryStr, args...)
	if err := row.Scan(&u.ID, &u.AuthProvider, &u.AuthSubject, &u.Email, &u.Name); err != nil {
		return nil, fmt.Errorf("upsert user: %w", err)
	}

	return u, nil
}
