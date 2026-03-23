package entity

import (
	"encoding/json"
	"time"
)

type Transcription struct {
	ID           int64
	UserID       int64
	FullText     string
	SegmentsJSON json.RawMessage
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
