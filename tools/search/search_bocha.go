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

func (p *BochaProvider) Search(ctx context.Context, query string, options *SearchOptions) ([]SearchResultItem, error) {
	if p.APIKey == "" {
		return nil, fmt.Errorf("search configuration (api_key) is missing")
	}

	url := "https://api.bochaai.com/v1/web-search"

	count := 10
	if options != nil && options.MaxResults > 0 {
		count = options.MaxResults
	}

	// Request body: {"query": "...", "freshness": "noLimit", "summary": true, "count": 20}
	payloadStr := fmt.Sprintf(`{"query":%q, "freshness": "noLimit", "summary": true, "count": %d}`, query, count)
	payload := strings.NewReader(payloadStr)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Add("Authorization", "Bearer "+p.APIKey)
	httpReq.Header.Add("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search API returned error status: %d", resp.StatusCode)
	}

	var searchResp BochaSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(searchResp.Data.WebPages.Value) == 0 {
		return []SearchResultItem{}, nil
	}

	results := make([]SearchResultItem, 0, len(searchResp.Data.WebPages.Value))
	for _, item := range searchResp.Data.WebPages.Value {
		// Use summary if available, otherwise snippet
		content := item.Summary
		if content == "" {
			content = item.Snippet
		}

		results = append(results, SearchResultItem{
			Title:   item.Name,
			URL:     item.Url,
			Content: content,
		})
	}
	return results, nil
}
