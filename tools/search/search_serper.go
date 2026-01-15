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

func (p *SerperProvider) Search(ctx context.Context, query string, options *SearchOptions) ([]SearchResultItem, error) {
	if p.APIKey == "" {
		return nil, fmt.Errorf("search configuration (api_key) is missing")
	}

	url := "https://google.serper.dev/search"

	num := 10
	if options != nil && options.MaxResults > 0 {
		num = options.MaxResults
	}

	payloadStr := fmt.Sprintf(`{"q":"%s", "num": %d}`, query, num)
	payload := strings.NewReader(payloadStr)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Add("X-API-KEY", p.APIKey)
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

	var searchResp SerperSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(searchResp.Organic) == 0 {
		return []SearchResultItem{}, nil
	}

	results := make([]SearchResultItem, 0, len(searchResp.Organic))
	for _, item := range searchResp.Organic {

		results = append(results, SearchResultItem{
			Title:   item.Title,
			URL:     item.Link,
			Content: item.Snippet,
		})
	}
	return results, nil
}
