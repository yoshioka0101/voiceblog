package domain

import (
	"errors"
	"time"
)

var (
	ErrNotFound  = errors.New("prompt not found")
	ErrForbidden = errors.New("prompt forbidden")
)

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

func (p *Prompt) OwnedBy(userID int64) bool {
	return p.UserID != nil && *p.UserID == userID
}

type CreateParams struct {
	UserID   int64
	Name     string
	Body     string
	IsActive bool
}

type PatchParams struct {
	Name     *string
	Body     *string
	IsActive *bool
}

type UpdateParams struct {
	ID       int64
	Name     string
	Body     string
	IsActive bool
}
