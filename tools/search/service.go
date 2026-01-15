package search

import (
	"context"
	"fmt"
)

// SearchType 搜索类型
type SearchType string

const (
	SearchTypeGoogle SearchType = "google"
	SearchTypeTavily SearchType = "tavily"
	SearchTypeBocha  SearchType = "bocha"
	SearchTypeSerper SearchType = "serper"
)

// SearchOptions 搜索选项
type SearchOptions struct {
	SearchDepth              string   `json:"search_depth"`               // basic, fast, ultra-fast, advanced
	IncludeAnswer            bool     `json:"include_answer"`             // 是否包含AI回答
	IncludeRawContent        bool     `json:"include_raw_content"`        // 是否包含原始内容 (markdown/text/true)
	MaxResults               int      `json:"max_results"`                // 最大结果数 (1-20)
	IncludeImages            bool     `json:"include_images"`             // 是否包含图片
	IncludeImageDescriptions bool     `json:"include_image_descriptions"` // 是否包含图片描述 (需 include_images=true)
	IncludeFavicon           bool     `json:"include_favicon"`            // 是否包含 favicon URL
	IncludeDomains           []string `json:"include_domains"`            // 包含的域名 (最多300个)
	ExcludeDomains           []string `json:"exclude_domains"`            // 排除的域名 (最多150个)
	Topic                    string   `json:"topic"`                      // 搜索类别: general, news, finance
	TimeRange                string   `json:"time_range"`                 // 时间范围: day/d, week/w, month/m, year/y
	ChunksPerSource          int      `json:"chunks_per_source"`          // 每个来源的块数 (1-3, 仅 advanced 模式)
	Country                  string   `json:"country"`                    // 国家过滤 (仅 general topic)
}

// Provider 搜索提供者接口
type Provider interface {
	Search(ctx context.Context, query string, options *SearchOptions) ([]SearchResultItem, error)
}

// SearchResultItem 单条搜索结果
type SearchResultItem struct {
	Title   string   `json:"title"`
	URL     string   `json:"url"`
	Content string   `json:"content"`
	Images  []string `json:"images,omitempty"` // 图片链接列表
	Answer  string   `json:"answer,omitempty"` // AI回答
}

// SearchResult 搜索结果
type SearchResult struct {
	Type  SearchType         `json:"type"`
	Query string             `json:"query"`
	Items []SearchResultItem `json:"items"`
}

// Config 搜索配置
type Config struct {
	GoogleAPIKey string
	GoogleCX     string
	TavilyAPIKey string
	BochaAPIKey  string
	SerperAPIKey string
}

// Service 搜索服务
type Service struct {
	providers map[SearchType]Provider
}

// NewService 创建搜索服务
func NewService(cfg *Config) *Service {
	s := &Service{
		providers: make(map[SearchType]Provider),
	}

	// 根据配置初始化可用的 Provider
	if cfg.GoogleAPIKey != "" && cfg.GoogleCX != "" {
		s.providers[SearchTypeGoogle] = &GoogleProvider{
			APIKey: cfg.GoogleAPIKey,
			CX:     cfg.GoogleCX,
		}
	}

	if cfg.TavilyAPIKey != "" {
		s.providers[SearchTypeTavily] = &TavilyProvider{
			APIKey: cfg.TavilyAPIKey,
		}
	}

	if cfg.BochaAPIKey != "" {
		s.providers[SearchTypeBocha] = &BochaProvider{
			APIKey: cfg.BochaAPIKey,
		}
	}

	if cfg.SerperAPIKey != "" {
		s.providers[SearchTypeSerper] = &SerperProvider{
			APIKey: cfg.SerperAPIKey,
		}
	}

	return s
}

// Search 执行搜索
func (s *Service) Search(ctx context.Context, searchType SearchType, query string, options *SearchOptions) (*SearchResult, error) {
	provider, ok := s.providers[searchType]
	if !ok {
		return nil, fmt.Errorf("search provider '%s' not configured", searchType)
	}

	items, err := provider.Search(ctx, query, options)
	if err != nil {
		return nil, err
	}

	return &SearchResult{
		Type:  searchType,
		Query: query,
		Items: items,
	}, nil
}

// GetAvailableProviders 获取可用的搜索提供者
func (s *Service) GetAvailableProviders() []SearchType {
	types := make([]SearchType, 0, len(s.providers))
	for t := range s.providers {
		types = append(types, t)
	}
	return types
}

// IsAvailable 检查搜索类型是否可用
func (s *Service) IsAvailable(searchType SearchType) bool {
	_, ok := s.providers[searchType]
	return ok
}

// ValidSearchTypes 所有有效的搜索类型
var ValidSearchTypes = []SearchType{
	SearchTypeGoogle,
	SearchTypeTavily,
	SearchTypeBocha,
	SearchTypeSerper,
}

// IsValidSearchType 检查是否为有效的搜索类型
func IsValidSearchType(t string) bool {
	for _, vt := range ValidSearchTypes {
		if string(vt) == t {
			return true
		}
	}
	return false
}
