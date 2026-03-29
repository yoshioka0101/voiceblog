package entity

import "context"

type Repository interface {
	ListVisiblePromptsByUserID(ctx context.Context, userID int64) ([]*Prompt, error)
	CreatePrompt(ctx context.Context, prompt *Prompt) (*Prompt, error)
	FindPromptByID(ctx context.Context, id int64) (*Prompt, error)
	UpdatePrompt(ctx context.Context, prompt *Prompt) (*Prompt, error)
	DeletePrompt(ctx context.Context, id int64) error
}
