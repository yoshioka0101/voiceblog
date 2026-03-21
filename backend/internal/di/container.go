package di

import (
	"database/sql"

	authfeature "github.com/yoshioka0101/voiceblog/backend/internal/feature/auth"
	userfeature "github.com/yoshioka0101/voiceblog/backend/internal/feature/user"
)

type Container struct {
	Auth *authfeature.Feature
	User *userfeature.Feature
}

func New(db *sql.DB, googleClientID string) *Container {
	userFeature := userfeature.New(db)
	authFeature := authfeature.New(googleClientID, userFeature.Repository)

	return &Container{
		Auth: authFeature,
		User: userFeature,
	}
}
