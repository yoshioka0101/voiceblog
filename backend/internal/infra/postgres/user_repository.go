package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/im"

	domain "github.com/yoshioka0101/voiceblog/backend/internal/domain/user"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindOrCreate(ctx context.Context, provider, subject, email, name string) (*domain.User, error) {
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

	u := &domain.User{}
	row := r.db.QueryRowContext(ctx, queryStr, args...)
	if err := row.Scan(&u.ID, &u.AuthProvider, &u.AuthSubject, &u.Email, &u.Name); err != nil {
		return nil, fmt.Errorf("upsert user: %w", err)
	}

	return u, nil
}
