package entity

import "context"

// Repository is the minimal contract for storing a transcription and reading it back by ID.
type Repository interface {
	CreateTranscription(ctx context.Context, transcription *Transcription) (*Transcription, error)
	FindTranscriptionByID(ctx context.Context, id int64) (*Transcription, error)
}
