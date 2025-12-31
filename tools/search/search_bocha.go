package search

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// --- Bocha Provider ---

type BochaProvider struct {
	APIKey string
}

type BochaSearchResponse struct {
	Data struct {
		WebPages struct {
			Value []struct {
				Name    string `json:"name"`
				Url     string `json:"url"`
				Snippet string `json:"snippet"`
				Summary string `json:"summary"`
			} `json:"value"`
		} `json:"webPages"`
	} `json:"data"`
}

func (p *BochaProvider) Search(ctx context.Context, query string) (string, error) {
	if p.APIKey == "" {
		return "", fmt.Errorf("search configuration (api_key) is missing")
	}

	url := "https://api.bochaai.com/v1/web-search"
	// Request body: {"query": "...", "freshness": "noLimit", "summary": true, "count": 10}
	payloadStr := fmt.Sprintf(`{"query":%q, "freshness": "noLimit", "summary": true, "count": 1}`, query)
	payload := strings.NewReader(payloadStr)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, payload)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Add("Authorization", "Bearer "+p.APIKey)
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

	var searchResp BochaSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(searchResp.Data.WebPages.Value) == 0 {
		return "No results found.", nil
	}

	var resultText string
	resultText = fmt.Sprintf("Found %d results for '%s':\n\n", len(searchResp.Data.WebPages.Value), query)
	for i, item := range searchResp.Data.WebPages.Value {
		if i >= 5 {
			break
		}
		// Use summary if available, otherwise snippet
		content := item.Summary
		if content == "" {
			content = item.Snippet
		}
		resultText += fmt.Sprintf("%d. %s\n   URL: %s\n   %s\n\n", i+1, item.Name, item.Url, content)
	}
	return resultText, nil
}
