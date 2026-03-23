package entity

import "context"

type Repository interface {
	ListVisibleByUserID(ctx context.Context, userID int64) ([]*Prompt, error)
	Create(ctx context.Context, prompt *Prompt) (*Prompt, error)
	FindByID(ctx context.Context, id int64) (*Prompt, error)
	Update(ctx context.Context, prompt *Prompt) (*Prompt, error)
	Delete(ctx context.Context, id int64) error
}
