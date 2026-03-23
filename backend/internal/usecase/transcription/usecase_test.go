package usecase_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	entity "github.com/yoshioka0101/voiceblog/backend/internal/entity/transcription"
	transcriptionusecase "github.com/yoshioka0101/voiceblog/backend/internal/usecase/transcription"
)

func TestCreateTranscription(t *testing.T) {
	repo := &transcriptionRepositoryStub{
		createFunc: func(_ context.Context, value *entity.Transcription) (*entity.Transcription, error) {
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

func TestListTranscriptionsByUserID(t *testing.T) {
	expected := []*entity.Transcription{{ID: 1, UserID: 7, FullText: "visible"}}
	repo := &transcriptionRepositoryStub{
		listByUserIDFunc: func(_ context.Context, userID int64) ([]*entity.Transcription, error) {
			if userID != 7 {
				t.Fatalf("userID = %d, want 7", userID)
			}
			return expected, nil
		},
	}

	uc := transcriptionusecase.NewUseCase(repo)
	values, err := uc.ListTranscriptionsByUserID(context.Background(), 7)
	if err != nil {
		t.Fatalf("ListTranscriptionsByUserID failed: %v", err)
	}
	if len(values) != 1 || values[0].ID != expected[0].ID {
		t.Fatalf("transcriptions = %#v", values)
	}
}

func TestGetTranscription_ForbiddenForOtherUser(t *testing.T) {
	repo := &transcriptionRepositoryStub{
		findByIDFunc: func(_ context.Context, id int64) (*entity.Transcription, error) {
			return &entity.Transcription{ID: id, UserID: 99}, nil
		},
	}

	uc := transcriptionusecase.NewUseCase(repo)
	_, err := uc.GetTranscription(context.Background(), 7, 4)
	if !errors.Is(err, transcriptionusecase.ErrForbidden) {
		t.Fatalf("err = %v, want ErrForbidden", err)
	}
}

type transcriptionRepositoryStub struct {
	createFunc       func(ctx context.Context, transcription *entity.Transcription) (*entity.Transcription, error)
	listByUserIDFunc func(ctx context.Context, userID int64) ([]*entity.Transcription, error)
	findByIDFunc     func(ctx context.Context, id int64) (*entity.Transcription, error)
}

func (s *transcriptionRepositoryStub) Create(ctx context.Context, transcription *entity.Transcription) (*entity.Transcription, error) {
	return s.createFunc(ctx, transcription)
}

func (s *transcriptionRepositoryStub) ListByUserID(ctx context.Context, userID int64) ([]*entity.Transcription, error) {
	return s.listByUserIDFunc(ctx, userID)
}

func (s *transcriptionRepositoryStub) FindByID(ctx context.Context, id int64) (*entity.Transcription, error) {
	return s.findByIDFunc(ctx, id)
}
