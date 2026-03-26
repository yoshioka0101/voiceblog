package hatena

import (
	"context"
	"crypto/sha1"
	"crypto/rand"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
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
		baseURL: "https://blog.hatena.ne.jp",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Verify checks if the token is valid by fetching the AtomPub service document.
// token format: "hatenaID:blogID:apiKey"
func (c *Client) Verify(ctx context.Context, token string) error {
	parts := strings.SplitN(token, ":", 3)
	if len(parts) != 3 {
		return fmt.Errorf("hatena token must be in format hatenaID:blogID:apiKey")
	}
	hatenaID, blogID, apiKey := parts[0], parts[1], parts[2]

	url := fmt.Sprintf("%s/%s/%s/atom", c.baseURL, hatenaID, blogID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("build hatena verify request: %w", err)
	}
	req.Header.Set("X-WSSE", buildWSSEHeader(hatenaID, apiKey))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request hatena verify: %w", err)
	}
	defer resp.Body.Close()
	io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("invalid token: hatena returned status %d", resp.StatusCode)
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("hatena verify failed: status %d", resp.StatusCode)
	}

	return nil
}

// token format: "hatenaID:blogID:apiKey"
func (c *Client) Publish(ctx context.Context, token, title, content string) (*PublishResult, error) {
	parts := strings.SplitN(token, ":", 3)
	if len(parts) != 3 {
		return nil, fmt.Errorf("hatena token must be in format hatenaID:blogID:apiKey")
	}
	hatenaID, blogID, apiKey := parts[0], parts[1], parts[2]

	entry := atomEntry{
		XMLName: xml.Name{Local: "entry"},
		Xmlns:   "http://www.w3.org/2005/Atom",
		Title:   title,
		Content: atomContent{
			Type: "text/plain",
			Body: content,
		},
	}

	body, err := xml.Marshal(entry)
	if err != nil {
		return nil, fmt.Errorf("marshal hatena request: %w", err)
	}

	url := fmt.Sprintf("%s/%s/%s/atom/entry", c.baseURL, hatenaID, blogID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(xml.Header+string(body)))
	if err != nil {
		return nil, fmt.Errorf("build hatena request: %w", err)
	}
	req.Header.Set("Content-Type", "application/xml")
	req.Header.Set("X-WSSE", buildWSSEHeader(hatenaID, apiKey))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request hatena: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read hatena response: %w", err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("hatena request failed: status=%d body=%s", resp.StatusCode, string(respBody))
	}

	var respEntry atomEntryResponse
	if err := xml.Unmarshal(respBody, &respEntry); err != nil {
		return nil, fmt.Errorf("decode hatena response: %w", err)
	}

	entryURL := ""
	for _, link := range respEntry.Links {
		if link.Rel == "alternate" {
			entryURL = link.Href
			break
		}
	}

	return &PublishResult{
		ID:  respEntry.ID,
		URL: entryURL,
	}, nil
}

type atomEntry struct {
	XMLName xml.Name    `xml:"entry"`
	Xmlns   string      `xml:"xmlns,attr"`
	Title   string      `xml:"title"`
	Content atomContent `xml:"content"`
}

type atomContent struct {
	Type string `xml:"type,attr"`
	Body string `xml:",chardata"`
}

type atomEntryResponse struct {
	XMLName xml.Name   `xml:"entry"`
	ID      string     `xml:"id"`
	Links   []atomLink `xml:"link"`
}

type atomLink struct {
	Rel  string `xml:"rel,attr"`
	Href string `xml:"href,attr"`
}

func buildWSSEHeader(username, password string) string {
	nonce := make([]byte, 20)
	rand.Read(nonce)
	nonceBase64 := base64.StdEncoding.EncodeToString(nonce)
	created := time.Now().UTC().Format(time.RFC3339)

	digest := sha1.Sum(append(append(nonce, []byte(created)...), []byte(password)...))
	digestBase64 := base64.StdEncoding.EncodeToString(digest[:])

	return fmt.Sprintf(
		`UsernameToken Username="%s", PasswordDigest="%s", Nonce="%s", Created="%s"`,
		username, digestBase64, nonceBase64, created,
	)
}
