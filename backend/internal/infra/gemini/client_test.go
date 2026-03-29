package gemini

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/yoshioka0101/voiceblog/backend/internal/usecase/articlegen"
)

func TestGenerateArticle_RateLimited(t *testing.T) {
	client := NewClient("test-api-key", "")
	client.baseURL = "https://example.test/v1beta/models"
	client.httpClient = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		return &http.Response{
			StatusCode: http.StatusTooManyRequests,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body: io.NopCloser(strings.NewReader(`{
			"error": {
				"code": 429,
				"message": "quota exceeded",
				"status": "RESOURCE_EXHAUSTED",
				"details": [
					{
						"@type": "type.googleapis.com/google.rpc.RetryInfo",
						"retryDelay": "35s"
					}
				]
			}
		}`)),
		}, nil
	})}

	_, err := client.GenerateArticle(context.Background(), articlegen.Input{
		PromptName: "blog",
		PromptBody: "write",
		FullText:   "full text",
	})
	if err == nil {
		t.Fatal("expected error")
	}

	var rateLimitErr *articlegen.RateLimitError
	if !errors.As(err, &rateLimitErr) {
		t.Fatalf("err = %T, want *articlegen.RateLimitError", err)
	}
	if rateLimitErr.RetryAfter != 35*time.Second {
		t.Fatalf("RetryAfter = %s, want 35s", rateLimitErr.RetryAfter)
	}
}

func TestGenerateArticle_UsesConfiguredModel(t *testing.T) {
	client := NewClient("test-api-key", "custom-model")
	client.baseURL = "https://example.test/v1beta/models"
	client.httpClient = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if !strings.Contains(r.URL.Path, "/models/custom-model:generateContent") {
			t.Fatalf("path = %s, want custom-model generateContent path", r.URL.Path)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body: io.NopCloser(strings.NewReader(`{
			"candidates": [
				{
					"content": {
						"parts": [
							{
								"text": "{\"title\":\"title\",\"content\":\"content\"}"
							}
						]
					}
				}
			]
		}`)),
		}, nil
	})}

	article, err := client.GenerateArticle(context.Background(), articlegen.Input{
		PromptName: "blog",
		PromptBody: "write",
		FullText:   "full text",
	})
	if err != nil {
		t.Fatalf("GenerateArticle failed: %v", err)
	}
	if article.Title != "title" || article.Content != "content" {
		t.Fatalf("article = %#v", article)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return fn(r)
}
