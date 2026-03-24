package entity

import "time"

const (
	StatusPending   = "pending"
	StatusRunning   = "running"
	StatusCompleted = "completed"
	StatusFailed    = "failed"
)

type Job struct {
	ID               int64
	TranscriptionID  int64
	PromptID         int64
	Status           string
	AttemptCount     int
	NextRunAt        time.Time
	ErrorMessage     *string
	GeneratedTitle   *string
	GeneratedContent *string
	CreatedAt        time.Time
}
