package entity

import "time"

type Article struct {
	ID             int64
	UserID         int64
	PromptRunJobID *int64
	Title          string
	Content        string
	DeletedAt      *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
