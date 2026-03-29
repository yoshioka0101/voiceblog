package usecase_test

import (
	"context"
	"encoding/json"
	"testing"

	entity "github.com/yoshioka0101/voiceblog/backend/internal/entity/transcription"
	transcriptionusecase "github.com/yoshioka0101/voiceblog/backend/internal/usecase/transcription"
)

func TestCreateTranscription(t *testing.T) {
	repo := &transcriptionRepositoryStub{
		createTranscriptionFunc: func(_ context.Context, value *entity.Transcription) (*entity.Transcription, error) {
			if value.UserID != 3 {
				t.Fatalf("UserID = %d, want 3", value.UserID)
			}
			if string(value.SegmentsJSON) != `[{"text":"hello"}]` {
				t.Fatalf("SegmentsJSON = %s", string(value.SegmentsJSON))
			}
			return &entity.Transcription{
				ID:           8,
				UserID:       value.UserID,
				FullText:     value.FullText,
				SegmentsJSON: value.SegmentsJSON,
			}, nil
		},
	}

	uc := transcriptionusecase.NewUseCase(repo)
	created, err := uc.CreateTranscription(context.Background(), transcriptionusecase.CreateTranscriptionInput{
		UserID:       3,
		FullText:     "hello",
		SegmentsJSON: json.RawMessage(`[{"text":"hello"}]`),
	})
	if err != nil {
		t.Fatalf("CreateTranscription failed: %v", err)
	}
	if created.ID != 8 {
		t.Fatalf("ID = %d, want 8", created.ID)
	}
}

type transcriptionRepositoryStub struct {
	createTranscriptionFunc   func(ctx context.Context, transcription *entity.Transcription) (*entity.Transcription, error)
	findTranscriptionByIDFunc func(ctx context.Context, id int64) (*entity.Transcription, error)
}

func (s *transcriptionRepositoryStub) CreateTranscription(ctx context.Context, transcription *entity.Transcription) (*entity.Transcription, error) {
	return s.createTranscriptionFunc(ctx, transcription)
}

func (s *transcriptionRepositoryStub) FindTranscriptionByID(ctx context.Context, id int64) (*entity.Transcription, error) {
	return s.findTranscriptionByIDFunc(ctx, id)
}
