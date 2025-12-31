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

func (p *GoogleProvider) Search(ctx context.Context, query string) (string, error) {
	if p.APIKey == "" || p.CX == "" {
		return "", fmt.Errorf("search configuration (api_key or cx) is missing")
	}

	baseURL := "https://www.googleapis.com/customsearch/v1"
	params := url.Values{}
	params.Add("key", p.APIKey)
	params.Add("cx", p.CX)
	params.Add("q", query)

	reqURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	httpReq, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("search request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("search API returned error status: %d", resp.StatusCode)
	}

	var searchResp GoogleSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(searchResp.Items) == 0 {
		return "No results found.", nil
	}

	var resultText string
	resultText = fmt.Sprintf("Found %d results for '%s':\n\n", len(searchResp.Items), query)
	for i, item := range searchResp.Items {
		if i >= 5 {
			break
		}
		resultText += fmt.Sprintf("%d. %s\n   URL: %s\n   %s\n\n", i+1, item.Title, item.Link, item.Snippet)
	}
	return resultText, nil
}
