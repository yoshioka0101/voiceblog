package di

import (
	"database/sql"

	authfeature "github.com/yoshioka0101/voiceblog/backend/internal/feature/auth"
	promptfeature "github.com/yoshioka0101/voiceblog/backend/internal/feature/prompt"
	transcriptionfeature "github.com/yoshioka0101/voiceblog/backend/internal/feature/transcription"
	userfeature "github.com/yoshioka0101/voiceblog/backend/internal/feature/user"
)

type Container struct {
	Auth          *authfeature.Feature
	Prompt        *promptfeature.Feature
	User          *userfeature.Feature
	Transcription *transcriptionfeature.Feature
}

func New(db *sql.DB, googleClientID string) *Container {
	userFeature := userfeature.New(db)
	authFeature := authfeature.New(googleClientID, userFeature.Repository)
	promptFeature := promptfeature.New(db)
	transcriptionFeature := transcriptionfeature.New(db)

	return &Container{
		Auth:          authFeature,
		Prompt:        promptFeature,
		User:          userFeature,
		Transcription: transcriptionFeature,
	}
}
