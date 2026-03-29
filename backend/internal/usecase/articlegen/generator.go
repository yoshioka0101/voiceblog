package articlegen

import (
	"context"
	"encoding/json"
)

type Generator interface {
	GenerateArticle(ctx context.Context, input Input) (*GeneratedArticle, error)
}

type Input struct {
	PromptName string
	PromptBody string
	FullText   string
	// SegmentsJSON is passed through as raw JSON because transcription segment fields depend on the speech provider.
	SegmentsJSON json.RawMessage
}

type GeneratedArticle struct {
	Title   string
	Content string
}
