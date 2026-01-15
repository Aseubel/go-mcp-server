package search

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// --- Google Provider ---

type GoogleProvider struct {
	APIKey string
	CX     string
}

type GoogleSearchResponse struct {
	Items []struct {
		Title   string `json:"title"`
		Link    string `json:"link"`
		Snippet string `json:"snippet"`
	} `json:"items"`
}

func (p *GoogleProvider) Search(ctx context.Context, query string, options *SearchOptions) ([]SearchResultItem, error) {
	if p.APIKey == "" || p.CX == "" {
		return nil, fmt.Errorf("search configuration (api_key or cx) is missing")
	}

	baseURL := "https://www.googleapis.com/customsearch/v1"
	params := url.Values{}
	params.Add("key", p.APIKey)
	params.Add("cx", p.CX)
	params.Add("q", query)

	if options != nil {
		if options.MaxResults > 0 {
			if options.MaxResults > 10 {
				options.MaxResults = 10 // Google API max is 10
			}
			params.Add("num", fmt.Sprintf("%d", options.MaxResults))
		}
	}

	reqURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	httpReq, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search API returned error status: %d", resp.StatusCode)
	}

	var searchResp GoogleSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(searchResp.Items) == 0 {
		return []SearchResultItem{}, nil
	}

	results := make([]SearchResultItem, 0, len(searchResp.Items))
	for _, item := range searchResp.Items {

		results = append(results, SearchResultItem{
			Title:   item.Title,
			URL:     item.Link,
			Content: item.Snippet,
		})
	}
	return results, nil
}
