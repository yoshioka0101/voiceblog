package entity

import "context"

type Repository interface {
	Create(ctx context.Context, transcription *Transcription) (*Transcription, error)
	ListByUserID(ctx context.Context, userID int64) ([]*Transcription, error)
	FindByID(ctx context.Context, id int64) (*Transcription, error)
}
