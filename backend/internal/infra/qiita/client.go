package qiita

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type PublishResult struct {
	ID  string
	URL string
}

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		baseURL: "https://qiita.com/api/v2",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type createItemRequest struct {
	Title   string   `json:"title"`
	Body    string   `json:"body"`
	Private bool     `json:"private"`
	Tags    []tag    `json:"tags"`
	Tweet   bool     `json:"tweet"`
}

type tag struct {
	Name string `json:"name"`
}

type createItemResponse struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

// Verify checks if the token is valid by calling GET /authenticated_user.
func (c *Client) Verify(ctx context.Context, token string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/authenticated_user", nil)
	if err != nil {
		return fmt.Errorf("build qiita verify request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request qiita verify: %w", err)
	}
	defer resp.Body.Close()
	io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("invalid token: qiita returned status %d", resp.StatusCode)
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("qiita verify failed: status %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) Publish(ctx context.Context, token, title, content string) (*PublishResult, error) {
	reqBody := createItemRequest{
		Title:   title,
		Body:    content,
		Private: false,
		Tags:    []tag{{Name: "voiceblog"}},
		Tweet:   false,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal qiita request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/items", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("build qiita request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request qiita: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read qiita response: %w", err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("qiita request failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	var result createItemResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("decode qiita response: %w", err)
	}

	return &PublishResult{
		ID:  result.ID,
		URL: result.URL,
	}, nil
}
