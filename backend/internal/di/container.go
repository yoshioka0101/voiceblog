package di

import (
	"database/sql"

	authfeature "github.com/yoshioka0101/voiceblog/backend/internal/feature/auth"
	transcriptionfeature "github.com/yoshioka0101/voiceblog/backend/internal/feature/transcription"
	userfeature "github.com/yoshioka0101/voiceblog/backend/internal/feature/user"
)

type Container struct {
	Auth          *authfeature.Feature
	User          *userfeature.Feature
	Transcription *transcriptionfeature.Feature
}

func New(db *sql.DB, googleClientID string) *Container {
	userFeature := userfeature.New(db)
	authFeature := authfeature.New(googleClientID, userFeature.Repository)
	transcriptionFeature := transcriptionfeature.New(db)

	return &Container{
		Auth:          authFeature,
		User:          userFeature,
		Transcription: transcriptionFeature,
	}
}
