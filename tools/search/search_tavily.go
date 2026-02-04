package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// --- Tavily Provider ---

type TavilyProvider struct {
	APIKey string
}

type TavilySearchRequest struct {
	Query                    string   `json:"query"`
	SearchDepth              string   `json:"search_depth,omitempty"`
	IncludeAnswer            bool     `json:"include_answer,omitempty"`
	IncludeRawContent        bool     `json:"include_raw_content,omitempty"`
	MaxResults               int      `json:"max_results,omitempty"`
	IncludeImages            bool     `json:"include_images,omitempty"`
	IncludeImageDescriptions bool     `json:"include_image_descriptions,omitempty"`
	IncludeFavicon           bool     `json:"include_favicon,omitempty"`
	IncludeDomains           []string `json:"include_domains,omitempty"`
	ExcludeDomains           []string `json:"exclude_domains,omitempty"`
	Topic                    string   `json:"topic,omitempty"`
	TimeRange                string   `json:"time_range,omitempty"`
	ChunksPerSource          int      `json:"chunks_per_source,omitempty"`
	Country                  string   `json:"country,omitempty"`
}

type TavilySearchResponse struct {
	Query   string   `json:"query"`
	Answer  string   `json:"answer"`
	Images  []string `json:"images"`
	Results []struct {
		Title      string  `json:"title"`
		URL        string  `json:"url"`
		Content    string  `json:"content"`
		RawContent string  `json:"raw_content"`
		Score      float64 `json:"score"`
		Favicon    string  `json:"favicon"`
	} `json:"results"`
}

func (p *TavilyProvider) Search(ctx context.Context, query string, options *SearchOptions) ([]SearchResultItem, error) {
	if p.APIKey == "" {
		return nil, fmt.Errorf("search configuration (api_key) is missing")
	}

	url := "https://api.tavily.com/search"

	reqBody := TavilySearchRequest{
		Query: query,
	}

	if options != nil {
		reqBody.SearchDepth = options.SearchDepth
		reqBody.IncludeAnswer = true
		reqBody.IncludeRawContent = options.IncludeRawContent
		reqBody.MaxResults = options.MaxResults
		reqBody.IncludeImages = options.IncludeImages
		reqBody.IncludeImageDescriptions = options.IncludeImageDescriptions
		reqBody.IncludeFavicon = options.IncludeFavicon
		reqBody.IncludeDomains = options.IncludeDomains
		reqBody.ExcludeDomains = options.ExcludeDomains
		reqBody.Topic = options.Topic
		reqBody.TimeRange = options.TimeRange
		reqBody.ChunksPerSource = options.ChunksPerSource
		reqBody.Country = options.Country
	}

	// Defaults if not set
	if reqBody.MaxResults == 0 {
		reqBody.MaxResults = 3
	}
	if reqBody.SearchDepth == "" {
		reqBody.SearchDepth = "basic"
	}

	jsonValue, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonValue))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Add("Authorization", "Bearer "+p.APIKey)
	httpReq.Header.Add("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second} // Tavily can be slow with advanced depth
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search API returned error status: %d", resp.StatusCode)
	}

	var searchResp TavilySearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var results []SearchResultItem
	// If we have an AI answer, we might want to return it as a special item or attached to the first item,
	// but the interface returns a list of items.
	// We can treat the answer as a separate logic or just attach it to the result items if possible.
	// For now, let's map the results.

	// If options asked for answer, we should probably include it.
	// The SearchResultItem struct was updated to include Answer and Images.

	// Since Answer and Images are global to the search response in Tavily (mostly),
	// we can attach them to the first item or duplicate them?
	// Attaching to the first item is a common pattern if the list is just a list.
	// Or we create a dummy item if there are no results but an answer?

	if len(searchResp.Results) == 0 {
		if searchResp.Answer != "" {
			return []SearchResultItem{{
				Content: searchResp.Answer,
				Answer:  searchResp.Answer,
				Images:  searchResp.Images,
			}}, nil
		}
		return []SearchResultItem{}, nil
	}

	for i, item := range searchResp.Results {
		resItem := SearchResultItem{
			Title:   item.Title,
			URL:     item.URL,
			Content: item.Content,
		}

		// If raw content was requested and returned, we could use it,
		// but SearchResultItem only has Content. existing consumers might expect snippet/summary.
		// item.Content is usually the snippet in Tavily.

		// Attach global answer and images to the FIRST item only to avoid duplication?
		// Or maybe the user wants them accessible easily.
		if i == 0 {
			resItem.Answer = searchResp.Answer
			resItem.Images = searchResp.Images
		}

		results = append(results, resItem)
	}

	return results, nil
}
