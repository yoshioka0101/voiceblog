package entity

import "context"

type Repository interface {
	CreatePromptRunJob(ctx context.Context, value *Job) (*Job, error)
	FindPromptRunJobByID(ctx context.Context, id int64) (*Job, error)
	UpdatePromptRunJob(ctx context.Context, value *Job) (*Job, error)
}
