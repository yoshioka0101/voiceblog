package entity

import (
	"encoding/json"
	"time"
)

type Transcription struct {
	ID       int64
	UserID   int64
	FullText string
	// SegmentsJSON keeps the speech-to-text segment payload as raw JSON because provider-specific fields can vary.
	SegmentsJSON json.RawMessage
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
