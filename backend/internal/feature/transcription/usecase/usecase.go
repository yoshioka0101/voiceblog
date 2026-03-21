package usecase

import (
	"context"

	transcriptiondomain "github.com/yoshioka0101/voiceblog/backend/internal/feature/transcription/domain"
)

type UseCase struct {
	repo transcriptiondomain.Repository
}

func New(repo transcriptiondomain.Repository) *UseCase {
	return &UseCase{repo: repo}
}

func (uc *UseCase) Create(ctx context.Context, params transcriptiondomain.CreateParams) (*transcriptiondomain.Transcription, error) {
	return uc.repo.Create(ctx, params)
}

func (uc *UseCase) ListByUserID(ctx context.Context, userID int64) ([]*transcriptiondomain.Transcription, error) {
	return uc.repo.ListByUserID(ctx, userID)
}

func (uc *UseCase) FindByID(ctx context.Context, userID, transcriptionID int64) (*transcriptiondomain.Transcription, error) {
	transcription, err := uc.repo.FindByID(ctx, transcriptionID)
	if err != nil {
		return nil, err
	}
	if transcription.UserID != userID {
		return nil, transcriptiondomain.ErrForbidden
	}
	return transcription, nil
}
