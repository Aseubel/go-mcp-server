package search

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// --- Bing Provider ---

type BingProvider struct {
	APIKey string
}

type BingSearchResponse struct {
	WebPages struct {
		Value []struct {
			Name    string `json:"name"`
			Url     string `json:"url"`
			Snippet string `json:"snippet"`
		} `json:"value"`
	} `json:"webPages"`
}

func (p *BingProvider) Search(ctx context.Context, query string) (string, error) {
	if p.APIKey == "" {
		return "", fmt.Errorf("search configuration (api_key) is missing")
	}

	baseURL := "https://api.bing.microsoft.com/v7.0/search"
	params := url.Values{}
	params.Add("q", query)

	reqURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	httpReq, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	
	httpReq.Header.Set("Ocp-Apim-Subscription-Key", p.APIKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("search request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("search API returned error status: %d", resp.StatusCode)
	}

	var searchResp BingSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(searchResp.WebPages.Value) == 0 {
		return "No results found.", nil
	}

	var resultText string
	resultText = fmt.Sprintf("Found %d results for '%s':\n\n", len(searchResp.WebPages.Value), query)
	for i, item := range searchResp.WebPages.Value {
		if i >= 5 {
			break
		}
		resultText += fmt.Sprintf("%d. %s\n   URL: %s\n   %s\n\n", i+1, item.Name, item.Url, item.Snippet)
	}
	return resultText, nil
}
