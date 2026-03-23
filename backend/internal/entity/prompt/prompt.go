package entity

import "time"

type Prompt struct {
	ID        int64
	UserID    *int64
	Name      string
	Body      string
	IsActive  bool
	IsDefault bool
	CreatedAt time.Time
	UpdatedAt time.Time
}
