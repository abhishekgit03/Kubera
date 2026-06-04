package gemini

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

const apiBase = "https://generativelanguage.googleapis.com/v1beta/models"

type Client struct {
	apiKey    string
	modelName string
	http      *http.Client
}

func New() (*Client, error) {
	modelName := os.Getenv("GEMINI_MODEL")
	if modelName == "" {
		modelName = "gemini-2.5-flash"
	}
	return &Client{
		apiKey:    os.Getenv("GEMINI_API_KEY"),
		modelName: modelName,
		http:      &http.Client{},
	}, nil
}

// StreamNarrative streams response tokens to ch, then closes it.
func (c *Client) StreamNarrative(ctx context.Context, prompt string, ch chan<- string) error {
	defer close(ch)

	if c.apiKey == "" {
		return fmt.Errorf("GEMINI_API_KEY is not set")
	}

	body, _ := json.Marshal(map[string]any{
		"contents": []map[string]any{
			{"role": "user", "parts": []map[string]any{{"text": prompt}}},
		},
		"generationConfig": map[string]any{
			"temperature":     0.3,
			"maxOutputTokens": 8192,
			// Disable thinking — all tokens go to the response, not internal reasoning
			"thinkingConfig": map[string]any{
				"thinkingBudget": 0,
			},
		},
	})

	url := fmt.Sprintf("%s/%s:streamGenerateContent?key=%s&alt=sse", apiBase, c.modelName, c.apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("gemini API %d: %s", resp.StatusCode, string(b))
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var chunk struct {
			Candidates []struct {
				Content struct {
					Parts []struct {
						Text    string `json:"text"`
						Thought bool   `json:"thought"` // thinking tokens — skip these
					} `json:"parts"`
				} `json:"content"`
			} `json:"candidates"`
		}
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue // skip malformed chunks
		}

		for _, cand := range chunk.Candidates {
			for _, part := range cand.Content.Parts {
				if part.Text == "" || part.Thought {
					continue
				}
				select {
				case ch <- part.Text:
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		}
	}

	return scanner.Err()
}
