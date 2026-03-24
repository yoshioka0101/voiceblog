package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	promptrunjobusecase "github.com/yoshioka0101/voiceblog/backend/internal/usecase/promptrunjob"
)

const defaultModel = "gemini-2.0-flash"

type Client struct {
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:  strings.TrimSpace(apiKey),
		model:   defaultModel,
		baseURL: "https://generativelanguage.googleapis.com/v1beta/models",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) GenerateArticle(ctx context.Context, input promptrunjobusecase.GenerateArticleInput) (*promptrunjobusecase.GeneratedArticle, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("gemini api key is required")
	}

	requestBody, err := json.Marshal(buildGenerateRequest(input))
	if err != nil {
		return nil, fmt.Errorf("marshal gemini request: %w", err)
	}

	url := fmt.Sprintf("%s/%s:generateContent?key=%s", c.baseURL, c.model, c.apiKey)
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("build gemini request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("request gemini: %w", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read gemini response: %w", err)
	}

	if response.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("gemini request failed: status=%d body=%s", response.StatusCode, strings.TrimSpace(string(body)))
	}

	var geminiResponse generateContentResponse
	if err := json.Unmarshal(body, &geminiResponse); err != nil {
		return nil, fmt.Errorf("decode gemini response: %w", err)
	}

	text := geminiResponse.firstText()
	if text == "" {
		return nil, fmt.Errorf("gemini response did not include article content")
	}

	article, err := parseGeneratedArticle(text)
	if err != nil {
		return nil, err
	}

	return article, nil
}

type generateContentRequest struct {
	Contents         []content        `json:"contents"`
	GenerationConfig generationConfig `json:"generationConfig"`
}

type content struct {
	Parts []part `json:"parts"`
}

type part struct {
	Text string `json:"text"`
}

type generationConfig struct {
	ResponseMIMEType string `json:"responseMimeType"`
}

type generateContentResponse struct {
	Candidates []candidate `json:"candidates"`
}

type candidate struct {
	Content content `json:"content"`
}

type generatedArticlePayload struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

func buildGenerateRequest(input promptrunjobusecase.GenerateArticleInput) generateContentRequest {
	prompt := fmt.Sprintf(`あなたは voiceblog の記事作成アシスタントです。
以下の指示と文字起こしをもとに、日本語の記事を JSON で返してください。

返却形式:
{"title":"記事タイトル","content":"記事本文"}

制約:
- title は 40 文字以内
- content は Markdown で返す
- 見出し、段落、箇条書きを必要に応じて使う
- 読みやすい段落構成にする
- 事実の追加創作はしない

prompt 名:
%s

prompt 本文:
%s

文字起こし:
%s`, input.PromptName, input.PromptBody, input.FullText)

	return generateContentRequest{
		Contents: []content{
			{
				Parts: []part{
					{Text: prompt},
				},
			},
		},
		GenerationConfig: generationConfig{
			ResponseMIMEType: "application/json",
		},
	}
}

func (r generateContentResponse) firstText() string {
	for _, candidate := range r.Candidates {
		for _, candidatePart := range candidate.Content.Parts {
			if text := strings.TrimSpace(candidatePart.Text); text != "" {
				return text
			}
		}
	}
	return ""
}

func parseGeneratedArticle(raw string) (*promptrunjobusecase.GeneratedArticle, error) {
	trimmed := strings.TrimSpace(raw)
	trimmed = strings.TrimPrefix(trimmed, "```json")
	trimmed = strings.TrimPrefix(trimmed, "```")
	trimmed = strings.TrimSuffix(trimmed, "```")
	trimmed = strings.TrimSpace(trimmed)

	var payload generatedArticlePayload
	if err := json.Unmarshal([]byte(trimmed), &payload); err != nil {
		return nil, fmt.Errorf("decode generated article: %w", err)
	}

	payload.Title = strings.TrimSpace(payload.Title)
	payload.Content = strings.TrimSpace(payload.Content)
	if payload.Title == "" || payload.Content == "" {
		return nil, fmt.Errorf("generated article must include title and content")
	}

	return &promptrunjobusecase.GeneratedArticle{
		Title:   payload.Title,
		Content: payload.Content,
	}, nil
}
