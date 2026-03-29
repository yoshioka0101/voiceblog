package db

import (
	"context"
	"database/sql"

	"github.com/stephenafamo/bob"
)

type TxRunner interface {
	RunInTx(ctx context.Context, fn func(context.Context) error) error
}

type txRunner struct {
	db bob.DB
}

type executorKey struct{}

func NewTxRunner(db *sql.DB) TxRunner {
	return &txRunner{db: bob.NewDB(db)}
}

func (r *txRunner) RunInTx(ctx context.Context, fn func(context.Context) error) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, exec bob.Executor) error {
		return fn(WithExecutor(ctx, exec))
	})
}

func WithExecutor(ctx context.Context, exec bob.Executor) context.Context {
	return context.WithValue(ctx, executorKey{}, exec)
}

func ExecutorFromContext(ctx context.Context, fallback bob.Executor) bob.Executor {
	if exec, ok := ctx.Value(executorKey{}).(bob.Executor); ok && exec != nil {
		return exec
	}

	return fallback
}

func InTx(ctx context.Context) bool {
	_, ok := ctx.Value(executorKey{}).(bob.Executor)
	return ok
}
