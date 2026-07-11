package validation

import (
	"errors"
	"testing"

	"github.com/yoshioka0101/voiceblog/backend/internal/api"
	"github.com/yoshioka0101/voiceblog/backend/internal/apperr"
)

func assertBadRequestCode(t *testing.T, err error, wantCode string) {
	t.Helper()

	var appErr *apperr.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("want *apperr.AppError, got %T (%v)", err, err)
	}
	if appErr.Status != 400 {
		t.Errorf("want status 400, got %d", appErr.Status)
	}
	if appErr.Code != wantCode {
		t.Errorf("want code %q, got %q", wantCode, appErr.Code)
	}
}

func TestIDParam(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    int64
		wantErr bool
	}{
		{name: "valid", raw: "42", want: 42},
		{name: "zero", raw: "0", wantErr: true},
		{name: "negative", raw: "-1", wantErr: true},
		{name: "not a number", raw: "abc", wantErr: true},
		{name: "empty", raw: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IDParam(tt.raw, "test id")
			if tt.wantErr {
				assertBadRequestCode(t, err, "invalid_id")
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("want %d, got %d", tt.want, got)
			}
		})
	}
}

func TestCreateArticleInput(t *testing.T) {
	valid := api.CreateArticleRequest{Title: "t", Content: "c"}

	input, err := CreateArticleInput(1, valid)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if input.UserID != 1 || input.Title != "t" || input.Content != "c" {
		t.Errorf("unexpected input: %+v", input)
	}

	_, err = CreateArticleInput(1, api.CreateArticleRequest{Title: "  ", Content: "c"})
	assertBadRequestCode(t, err, "title_required")

	_, err = CreateArticleInput(1, api.CreateArticleRequest{Title: "t", Content: ""})
	assertBadRequestCode(t, err, "content_required")

	badID := int64(0)
	_, err = CreateArticleInput(1, api.CreateArticleRequest{Title: "t", Content: "c", PromptRunJobId: &badID})
	assertBadRequestCode(t, err, "invalid_id")
}

func TestUpdateArticleInput(t *testing.T) {
	title := "t"
	blank := " "

	if _, err := UpdateArticleInput(1, 2, api.UpdateArticleRequest{Title: &title}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err := UpdateArticleInput(1, 2, api.UpdateArticleRequest{})
	assertBadRequestCode(t, err, "field_required")

	_, err = UpdateArticleInput(1, 2, api.UpdateArticleRequest{Title: &blank})
	assertBadRequestCode(t, err, "title_required")

	_, err = UpdateArticleInput(1, 2, api.UpdateArticleRequest{Content: &blank})
	assertBadRequestCode(t, err, "content_required")
}

func TestGenerateArticleInput(t *testing.T) {
	if _, err := GenerateArticleInput(1, api.GenerateArticleRequest{TranscriptionId: 1, PromptId: 2}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err := GenerateArticleInput(1, api.GenerateArticleRequest{TranscriptionId: 0, PromptId: 2})
	assertBadRequestCode(t, err, "invalid_id")

	_, err = GenerateArticleInput(1, api.GenerateArticleRequest{TranscriptionId: 1, PromptId: -1})
	assertBadRequestCode(t, err, "invalid_id")
}

func TestCreatePromptInput(t *testing.T) {
	if _, err := CreatePromptInput(1, api.CreatePromptRequest{Name: "n", Body: "b"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err := CreatePromptInput(1, api.CreatePromptRequest{Name: "", Body: "b"})
	assertBadRequestCode(t, err, "name_required")

	_, err = CreatePromptInput(1, api.CreatePromptRequest{Name: "n", Body: "\t"})
	assertBadRequestCode(t, err, "body_required")
}

func TestUpdatePromptInput(t *testing.T) {
	active := true
	blank := ""

	if _, err := UpdatePromptInput(1, 2, api.UpdatePromptRequest{IsActive: &active}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err := UpdatePromptInput(1, 2, api.UpdatePromptRequest{})
	assertBadRequestCode(t, err, "field_required")

	_, err = UpdatePromptInput(1, 2, api.UpdatePromptRequest{Name: &blank})
	assertBadRequestCode(t, err, "name_required")
}

func TestCreateTranscriptionInput(t *testing.T) {
	segments := []map[string]interface{}{{"text": "hello", "start_ms": 0, "end_ms": 100}}

	input, err := CreateTranscriptionInput(1, api.CreateTranscriptionRequest{FullText: "hello", SegmentsJson: segments})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(input.SegmentsJSON) == 0 {
		t.Error("SegmentsJSON should not be empty")
	}

	_, err = CreateTranscriptionInput(1, api.CreateTranscriptionRequest{FullText: " ", SegmentsJson: segments})
	assertBadRequestCode(t, err, "full_text_required")

	_, err = CreateTranscriptionInput(1, api.CreateTranscriptionRequest{FullText: "hello"})
	assertBadRequestCode(t, err, "segments_required")
}

func TestCreatePromptRunJobInput(t *testing.T) {
	if _, err := CreatePromptRunJobInput(1, api.CreatePromptRunJobRequest{TranscriptionId: 1, PromptId: 2}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err := CreatePromptRunJobInput(1, api.CreatePromptRunJobRequest{TranscriptionId: 0, PromptId: 2})
	assertBadRequestCode(t, err, "invalid_id")
}
