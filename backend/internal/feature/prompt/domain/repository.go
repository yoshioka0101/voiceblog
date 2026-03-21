package domain

import "context"

type Repository interface {
	ListVisibleByUserID(ctx context.Context, userID int64) ([]*Prompt, error)
	Create(ctx context.Context, params CreateParams) (*Prompt, error)
	FindByID(ctx context.Context, id int64) (*Prompt, error)
	Update(ctx context.Context, params UpdateParams) (*Prompt, error)
	Delete(ctx context.Context, id int64) error
}
