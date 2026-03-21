package domain

import (
	"encoding/json"
	"errors"
	"time"
)

var (
	ErrNotFound  = errors.New("transcription not found")
	ErrForbidden = errors.New("transcription forbidden")
)

type Transcription struct {
	ID           int64
	UserID       int64
	FullText     string
	SegmentsJSON json.RawMessage
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type CreateParams struct {
	UserID       int64
	FullText     string
	SegmentsJSON json.RawMessage
}
