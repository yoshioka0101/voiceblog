package entity

import "time"

// These values are persisted to the DB and returned by the API, so they stay as stable string constants.
const (
	// StatusPending means the job record exists but generation has not started yet.
	StatusPending = "pending"
	// StatusRunning means the job is currently generating article content.
	StatusRunning = "running"
	// StatusCompleted means article generation succeeded and generated content is available.
	StatusCompleted = "completed"
	// StatusFailed means generation stopped with an error message for the caller.
	StatusFailed = "failed"
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
