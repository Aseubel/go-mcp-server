package search

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// --- Serper Provider ---

type SerperProvider struct {
	APIKey string
}

type SerperSearchResponse struct {
	Organic []struct {
		Title   string `json:"title"`
		Link    string `json:"link"`
		Snippet string `json:"snippet"`
	} `json:"organic"`
}

func (p *SerperProvider) Search(ctx context.Context, query string) (string, error) {
	if p.APIKey == "" {
		return "", fmt.Errorf("search configuration (api_key) is missing")
	}

	url := "https://google.serper.dev/search"
	payload := strings.NewReader(fmt.Sprintf(`{"q":"%s"}`, query))

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, payload)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Add("X-API-KEY", p.APIKey)
	httpReq.Header.Add("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("search request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("search API returned error status: %d", resp.StatusCode)
	}

	var searchResp SerperSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(searchResp.Organic) == 0 {
		return "No results found.", nil
	}

	var resultText string
	resultText = fmt.Sprintf("Found %d results for '%s':\n\n", len(searchResp.Organic), query)
	for i, item := range searchResp.Organic {
		if i >= 5 {
			break
		}
		resultText += fmt.Sprintf("%d. %s\n   URL: %s\n   %s\n\n", i+1, item.Title, item.Link, item.Snippet)
	}
	return resultText, nil
}
