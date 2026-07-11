package postgresql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/im"

	dbctx "github.com/yoshioka0101/voiceblog/backend/internal/db"
	entity "github.com/yoshioka0101/voiceblog/backend/internal/entity/user"
)

type UserRepository struct {
	exec bob.Executor
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{exec: bob.NewDB(db)}
}

func NewUserRepositoryWithExecutor(exec bob.Executor) *UserRepository {
	return &UserRepository{exec: exec}
}

func (r *UserRepository) FindOrCreate(ctx context.Context, provider, subject, email, name string) (*entity.User, error) {
	exec := dbctx.ExecutorFromContext(ctx, r.exec)
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

	u := &entity.User{}
	rows, err := exec.QueryContext(ctx, queryStr, args...)
	if err != nil {
		return nil, fmt.Errorf("upsert user: %w", err)
	}
	defer func() { _ = rows.Close() }()

	if !rows.Next() {
		return nil, fmt.Errorf("upsert user: no rows returned")
	}
	if err := rows.Scan(&u.ID, &u.AuthProvider, &u.AuthSubject, &u.Email, &u.Name); err != nil {
		return nil, fmt.Errorf("upsert user: %w", err)
	}

	return u, nil
}
