package web

import (
	"os"
	"strings"
	"sync"

	tools "github.com/OctoSucker/octosucker-tools"
)

const providerName = "github.com/OctoSucker/tools-web"

type SkillWeb struct {
	mu sync.RWMutex

	BraveSearchConfig

	fetchMaxChars int
}

func (s *SkillWeb) Init(config map[string]interface{}, submitTask func(string) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.fetchMaxChars = 50000
	s.BraveSearchConfig = BraveSearchConfig{
		APIKey:    "",
		Count:     defaultSearchResultCount,
		Country:   defaultSearchCountry,
		Language:  defaultSearchLanguage,
		Freshness: "",
		Timeout:   0,
	}
	if config != nil {
		if v, ok := config["fetch_max_chars"].(float64); ok && v > 0 {
			s.fetchMaxChars = int(v)
		}
		if v, ok := config["fetch_max_chars"].(int); ok && v > 0 {
			s.fetchMaxChars = v
		}
		if v, ok := config["search_api_key"].(string); ok && strings.TrimSpace(v) != "" {
			s.APIKey = strings.TrimSpace(v)
		}
		if v, ok := config["search_count"].(float64); ok && v > 0 {
			s.Count = int(v)
		}
		if v, ok := config["search_count"].(int); ok && v > 0 {
			s.Count = v
		}
		if v, ok := config["search_country"].(string); ok && strings.TrimSpace(v) != "" {
			s.Country = strings.TrimSpace(v)
		}
		if v, ok := config["search_language"].(string); ok && strings.TrimSpace(v) != "" {
			s.Language = strings.TrimSpace(v)
		}
		if v, ok := config["search_freshness"].(string); ok && strings.TrimSpace(v) != "" {
			s.Freshness = strings.TrimSpace(v)
		}
	}
	if s.APIKey == "" {
		if env := strings.TrimSpace(os.Getenv("BRAVE_API_KEY")); env != "" {
			s.APIKey = env
		}
	}
	return nil
}

func (s *SkillWeb) Cleanup() error {
	s.mu.Lock()
	s.fetchMaxChars = 0
	s.mu.Unlock()
	return nil
}

func (s *SkillWeb) getFetchMaxChars() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.fetchMaxChars
}

func (s *SkillWeb) getSearchConfig() BraveSearchConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.BraveSearchConfig
}

func (s *SkillWeb) Register(registry *tools.ToolRegistry, agent interface{}, providerName string) error {
	registry.RegisterTool(providerName, &tools.Tool{
		Name:        "web_fetch",
		Description: "通过 HTTP 拉取 URL 的纯文本（不启动浏览器，仅适合静态页）。用户若要求「用 Google 搜索并给我结果」时必须用 browser_navigate，禁止用 web_fetch 拉 Google 或新闻站。",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"url":       map[string]interface{}{"type": "string", "description": "要拉取的网页 URL"},
				"max_chars": map[string]interface{}{"type": "integer", "description": "最多返回字符数（可选）"},
			},
			"required": []string{"url"},
		},
		Handler: handleWebFetch,
	})

	registry.RegisterTool(providerName, &tools.Tool{
		Name:        "web_read",
		Description: "使用 Jina Reader（https://r.jina.ai）将任意 URL 转换为适合 LLM 的 Markdown 文本，仅用于阅读网页内容，不做交互。",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"url": map[string]interface{}{
					"type":        "string",
					"description": "要读取的网页 URL",
				},
				"respond_with": map[string]interface{}{
					"type":        "string",
					"description": "可选：控制返回格式的 X-Respond-With 头，如 markdown/html/text",
				},
			},
			"required": []string{"url"},
		},
		Handler: handleWebRead,
	})

	registry.RegisterTool(providerName, &tools.Tool{
		Name:        "web_search",
		Description: "使用 Brave Search API 搜索网络，返回标题、链接与摘要。适用于「查新闻、找资料」等场景，不会打开浏览器。需要在配置中提供 search_api_key 或环境变量 BRAVE_API_KEY。",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]interface{}{"type": "string", "description": "搜索关键词"},
				"count": map[string]interface{}{"type": "integer", "description": "返回结果条数（1-10，可选）"},
				"country": map[string]interface{}{
					"type":        "string",
					"description": "国家代码，如 US、ALL（可选）",
				},
				"language": map[string]interface{}{
					"type":        "string",
					"description": "语言代码，如 en、zh-hans（可选）",
				},
				"freshness": map[string]interface{}{
					"type":        "string",
					"description": "时间范围：day/week/month/year 或 YYYY-MM-DDtoYYYY-MM-DD（可选）",
				},
			},
			"required": []string{"query"},
		},
		Handler: handleWebSearch,
	})

	return nil
}

var globalSkillWeb *SkillWeb

func init() {
	globalSkillWeb = &SkillWeb{}
	tools.RegisterToolProvider(&tools.ToolProviderInfo{
		Name:        providerName,
		Description: "Web - HTTP 抓取与搜索（Jina Reader + Brave Search），不依赖浏览器自动化",
		Provider:    globalSkillWeb,
	})
}
