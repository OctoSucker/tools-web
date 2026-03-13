package web

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type SearchResult struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Published   string `json:"published,omitempty"`
	SiteName    string `json:"site_name,omitempty"`
}

type braveSearchResponse struct {
	Web *struct {
		Results []struct {
			Title       string `json:"title"`
			URL         string `json:"url"`
			Description string `json:"description"`
			Age         string `json:"age"`
		} `json:"results"`
	} `json:"web"`
}

type BraveSearchConfig struct {
	APIKey    string
	Count     int
	Country   string
	Language  string
	Freshness string
	Timeout   time.Duration
}

const (
	braveSearchEndpoint      = "https://api.search.brave.com/res/v1/web/search"
	defaultSearchTimeout     = 20 * time.Second
	defaultSearchCountry     = "US"
	defaultSearchLanguage    = ""
	defaultSearchResultCount = 5
	maxSearchResultCount     = 10
)

func handleWebSearch(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	query, ok := params["query"].(string)
	if !ok || strings.TrimSpace(query) == "" {
		return nil, fmt.Errorf("query is required")
	}

	baseCfg := globalSkillWeb.getSearchConfig()

	if v, ok := params["count"].(float64); ok && v > 0 {
		baseCfg.Count = int(v)
	}
	if v, ok := params["count"].(int); ok && v > 0 {
		baseCfg.Count = v
	}
	if v, ok := params["country"].(string); ok && strings.TrimSpace(v) != "" {
		baseCfg.Country = strings.TrimSpace(v)
	}
	if v, ok := params["language"].(string); ok && strings.TrimSpace(v) != "" {
		baseCfg.Language = strings.TrimSpace(v)
	}
	if v, ok := params["freshness"].(string); ok && strings.TrimSpace(v) != "" {
		baseCfg.Freshness = strings.TrimSpace(v)
	}

	if strings.TrimSpace(baseCfg.APIKey) == "" {
		return map[string]interface{}{
			"success": false,
			"error":   "web_search requires Brave Search API key. Please set search_api_key in config for github.com/OctoSucker/tools-web or BRAVE_API_KEY in environment.",
			"hint":    "在 OctoSucker 配置中为 github.com/OctoSucker/tools-web 设置 search_api_key，或在环境中设置 BRAVE_API_KEY（Brave Search API Key）。",
		}, nil
	}

	results, took, err := BraveWebSearch(ctx, query, baseCfg)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
			"hint":    "调用 Brave Search API 失败，请检查 search_api_key/BRAVE_API_KEY 是否正确，以及网络或代理配置是否可用。",
		}, nil
	}

	out := make([]map[string]interface{}, 0, len(results))
	for _, r := range results {
		item := map[string]interface{}{
			"title":       r.Title,
			"url":         r.URL,
			"description": r.Description,
		}
		if r.Published != "" {
			item["published"] = r.Published
		}
		if r.SiteName != "" {
			item["site_name"] = r.SiteName
		}
		out = append(out, item)
	}

	return map[string]interface{}{
		"success":  true,
		"query":    query,
		"count":    len(out),
		"took_ms":  took,
		"results":  out,
		"provider": "brave",
		"hint":     "已返回 Brave 搜索结果。Agent 可从 results 中选择合适的标题与链接，用 send_telegram_message 等工具回复用户。",
	}, nil
}

func normalizeFreshnessToBrave(freshness string) string {
	f := strings.TrimSpace(strings.ToLower(freshness))
	switch f {
	case "":
		return ""
	case "day":
		return "pd"
	case "week":
		return "pw"
	case "month":
		return "pm"
	case "year":
		return "py"
	default:
		// 透传日期范围、pd/pw/pm/py 等 Brave 自身支持的格式
		return freshness
	}
}

func BraveWebSearch(ctx context.Context, q string, cfg BraveSearchConfig) ([]SearchResult, int64, error) {
	query := strings.TrimSpace(q)
	if query == "" {
		return nil, 0, fmt.Errorf("query is required")
	}
	apiKey := strings.TrimSpace(cfg.APIKey)
	if apiKey == "" {
		return nil, 0, fmt.Errorf("Brave Search API key is required")
	}

	count := cfg.Count
	if count <= 0 {
		count = defaultSearchResultCount
	}
	if count > maxSearchResultCount {
		count = maxSearchResultCount
	}

	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = defaultSearchTimeout
	}

	u, err := url.Parse(braveSearchEndpoint)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid brave endpoint: %w", err)
	}
	qv := u.Query()
	qv.Set("q", query)
	qv.Set("count", strconv.Itoa(count))
	country := strings.TrimSpace(cfg.Country)
	if country == "" {
		country = defaultSearchCountry
	}
	if country != "" {
		qv.Set("country", country)
	}
	if cfg.Language != "" {
		qv.Set("search_lang", strings.TrimSpace(cfg.Language))
	}
	if f := normalizeFreshnessToBrave(cfg.Freshness); f != "" {
		qv.Set("freshness", f)
	}
	u.RawQuery = qv.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Subscription-Token", apiKey)

	baseClient := getFetchClient()
	transport := baseClient.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}
	start := time.Now()
	resp, err := client.Do(req)
	elapsed := time.Since(start)
	if err != nil {
		return nil, elapsed.Milliseconds(), fmt.Errorf("brave search: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var detail string
		if resp.Body != nil {
			b, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
			detail = strings.TrimSpace(string(b))
		}
		if detail != "" {
			return nil, elapsed.Milliseconds(), fmt.Errorf("brave search returned %d: %s", resp.StatusCode, detail)
		}
		return nil, elapsed.Milliseconds(), fmt.Errorf("brave search returned %d", resp.StatusCode)
	}

	var parsed braveSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, elapsed.Milliseconds(), fmt.Errorf("decode brave response: %w", err)
	}

	results := make([]SearchResult, 0)
	if parsed.Web == nil {
		return results, elapsed.Milliseconds(), nil
	}
	for _, r := range parsed.Web.Results {
		if r.URL == "" && r.Title == "" && r.Description == "" {
			continue
		}
		results = append(results, SearchResult{
			Title:       strings.TrimSpace(r.Title),
			URL:         strings.TrimSpace(r.URL),
			Description: strings.TrimSpace(r.Description),
			Published:   strings.TrimSpace(r.Age),
		})
	}
	return results, elapsed.Milliseconds(), nil
}
