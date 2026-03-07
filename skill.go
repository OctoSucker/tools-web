package web

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"sync"

	skill "github.com/OctoSucker/octosucker-skill"
)

const skillName = "github.com/OctoSucker/skill-web"

// SkillWeb Web Skill（浏览器自动化 + 简单 HTTP 抓取，不依赖 Brave/Serper）
type SkillWeb struct {
	mu            sync.RWMutex
	fetchMaxChars int
}

// Init 从 config 读取 fetch 相关配置（如 fetch_max_chars）
func (s *SkillWeb) Init(config map[string]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.fetchMaxChars = 50000
	if config != nil {
		if v, ok := config["fetch_max_chars"].(float64); ok && v > 0 {
			s.fetchMaxChars = int(v)
		}
		if v, ok := config["fetch_max_chars"].(int); ok && v > 0 {
			s.fetchMaxChars = v
		}
	}
	return nil
}

// Cleanup 清理（含关闭浏览器）
func (s *SkillWeb) Cleanup() error {
	closeBrowser()
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

// RegisterWebSkill 注册 browser_* 与 web_fetch
func RegisterWebSkill(registry *skill.ToolRegistry, agent interface{}) error {
	registry.Register(&skill.Tool{
		Name:        "browser_navigate",
		Description: "在浏览器中打开 URL，返回压缩 DOM（可操作元素列表与主文本）。用于后续 click/type/scroll/extract。",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"url": map[string]interface{}{
					"type":        "string",
					"description": "要打开的网页 URL",
				},
			},
			"required": []string{"url"},
		},
		Handler: handleBrowserNavigate,
	})
	registry.Register(&skill.Tool{
		Name:        "browser_click",
		Description: "点击页面上元素。ref 可为压缩 DOM 中的 ref（如 \"0\"）、或按钮/链接的可见文本（如 \"登录\"）、或 aria-label。",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"ref": map[string]interface{}{
					"type":        "string",
					"description": "元素引用：ref 数字或可见文本或 aria-label",
				},
			},
			"required": []string{"ref"},
		},
		Handler: handleBrowserClick,
	})
	registry.Register(&skill.Tool{
		Name:        "browser_type",
		Description: "在输入框等元素中输入文字。ref 同 browser_click。",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"ref":  map[string]interface{}{"type": "string", "description": "元素引用"},
				"text": map[string]interface{}{"type": "string", "description": "要输入的文字"},
			},
			"required": []string{"ref", "text"},
		},
		Handler: handleBrowserType,
	})
	registry.Register(&skill.Tool{
		Name:        "browser_scroll",
		Description: "滚动当前页面。direction: up | down | top | bottom。",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"direction": map[string]interface{}{"type": "string", "description": "up / down / top / bottom"},
			},
			"required": []string{"direction"},
		},
		Handler: handleBrowserScroll,
	})
	registry.Register(&skill.Tool{
		Name:        "browser_extract",
		Description: "从当前页面提取主文本（标题与段落），用于阅读页面内容。",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"max_chars": map[string]interface{}{"type": "integer", "description": "最多返回字符数（可选）"},
			},
		},
		Handler: handleBrowserExtract,
	})
	registry.Register(&skill.Tool{
		Name:        "browser_hover",
		Description: "将鼠标悬停到元素上，用于触发下拉菜单、tooltip 等。ref 同 browser_click。",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"ref": map[string]interface{}{"type": "string", "description": "元素引用"},
			},
			"required": []string{"ref"},
		},
		Handler: handleBrowserHover,
	})
	registry.Register(&skill.Tool{
		Name:        "browser_select_option",
		Description: "在 select 下拉框中选择选项。ref 为 select 元素引用，value 为选项的可见文本。",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"ref":   map[string]interface{}{"type": "string", "description": "select 元素引用"},
				"value": map[string]interface{}{"type": "string", "description": "要选择的选项文本"},
			},
			"required": []string{"ref", "value"},
		},
		Handler: handleBrowserSelectOption,
	})
	registry.Register(&skill.Tool{
		Name:        "browser_screenshot",
		Description: "截取当前页面截图，返回 base64 编码的 PNG，可用于视觉确认或多模态理解。",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"full_page": map[string]interface{}{"type": "boolean", "description": "为 true 时截取整页长图"},
			},
		},
		Handler: handleBrowserScreenshot,
	})
	registry.Register(&skill.Tool{
		Name:        "browser_go_back",
		Description: "浏览器后退到上一页。",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{},
		},
		Handler: handleBrowserGoBack,
	})
	registry.Register(&skill.Tool{
		Name:        "browser_go_forward",
		Description: "浏览器前进到下一页。",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{},
		},
		Handler: handleBrowserGoForward,
	})
	registry.Register(&skill.Tool{
		Name:        "browser_check",
		Description: "设置 checkbox 或 radio 的选中状态。ref 为元素引用，checked 为 true 表示选中、false 表示取消。",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"ref":     map[string]interface{}{"type": "string", "description": "checkbox/radio 元素引用"},
				"checked": map[string]interface{}{"type": "boolean", "description": "是否选中"},
			},
			"required": []string{"ref", "checked"},
		},
		Handler: handleBrowserCheck,
	})
	registry.Register(&skill.Tool{
		Name:        "browser_wait",
		Description: "等待页面就绪：若提供 selector 则等待该元素出现（最多 timeout_seconds 秒）；否则仅等待 timeout_seconds 秒。",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"selector":         map[string]interface{}{"type": "string", "description": "可选，CSS 选择器，等待该元素出现"},
				"timeout_seconds":  map[string]interface{}{"type": "integer", "description": "超时秒数，默认 2"},
			},
		},
		Handler: handleBrowserWait,
	})

	registry.Register(&skill.Tool{
		Name:        "web_fetch",
		Description: "通过 HTTP 拉取 URL 的网页内容并提取为纯文本（不启动浏览器，适合简单静态页）。",
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

	return nil
}

func handleBrowserNavigate(params map[string]interface{}) (interface{}, error) {
	urlStr, ok := params["url"].(string)
	if !ok || urlStr == "" {
		return nil, fmt.Errorf("url is required")
	}
	dom, err := Navigate(urlStr)
	if err != nil {
		return nil, err
	}
	return domToMap(dom), nil
}

func handleBrowserClick(params map[string]interface{}) (interface{}, error) {
	ref, ok := params["ref"].(string)
	if !ok || ref == "" {
		return nil, fmt.Errorf("ref is required")
	}
	dom, err := Click(ref)
	if err != nil {
		return nil, err
	}
	return domToMap(dom), nil
}

func handleBrowserType(params map[string]interface{}) (interface{}, error) {
	ref, ok := params["ref"].(string)
	if !ok || ref == "" {
		return nil, fmt.Errorf("ref is required")
	}
	text, _ := params["text"].(string)
	dom, err := Type(ref, text)
	if err != nil {
		return nil, err
	}
	return domToMap(dom), nil
}

func handleBrowserScroll(params map[string]interface{}) (interface{}, error) {
	dir, _ := params["direction"].(string)
	if dir == "" {
		dir = "down"
	}
	dom, err := Scroll(dir)
	if err != nil {
		return nil, err
	}
	return domToMap(dom), nil
}

func handleBrowserExtract(params map[string]interface{}) (interface{}, error) {
	maxChars := globalSkillWeb.getFetchMaxChars()
	if v, ok := params["max_chars"].(float64); ok && v > 0 {
		maxChars = int(v)
	}
	if v, ok := params["max_chars"].(int); ok && v > 0 {
		maxChars = v
	}
	text, err := Extract(maxChars)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"success": true, "content": text, "length": len(text)}, nil
}

func handleBrowserHover(params map[string]interface{}) (interface{}, error) {
	ref, ok := params["ref"].(string)
	if !ok || ref == "" {
		return nil, fmt.Errorf("ref is required")
	}
	dom, err := Hover(ref)
	if err != nil {
		return nil, err
	}
	return domToMap(dom), nil
}

func handleBrowserSelectOption(params map[string]interface{}) (interface{}, error) {
	ref, ok := params["ref"].(string)
	if !ok || ref == "" {
		return nil, fmt.Errorf("ref is required")
	}
	value, _ := params["value"].(string)
	dom, err := SelectOption(ref, value)
	if err != nil {
		return nil, err
	}
	return domToMap(dom), nil
}

func handleBrowserScreenshot(params map[string]interface{}) (interface{}, error) {
	fullPage := false
	if v, ok := params["full_page"].(bool); ok {
		fullPage = v
	}
	data, err := Screenshot(fullPage)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"success":      true,
		"image_base64": base64.StdEncoding.EncodeToString(data),
		"format":       "png",
	}, nil
}

func handleBrowserGoBack(params map[string]interface{}) (interface{}, error) {
	dom, err := GoBack()
	if err != nil {
		return nil, err
	}
	return domToMap(dom), nil
}

func handleBrowserGoForward(params map[string]interface{}) (interface{}, error) {
	dom, err := GoForward()
	if err != nil {
		return nil, err
	}
	return domToMap(dom), nil
}

func handleBrowserCheck(params map[string]interface{}) (interface{}, error) {
	ref, ok := params["ref"].(string)
	if !ok || ref == "" {
		return nil, fmt.Errorf("ref is required")
	}
	checked := false
	if v, ok := params["checked"].(bool); ok {
		checked = v
	}
	dom, err := Check(ref, checked)
	if err != nil {
		return nil, err
	}
	return domToMap(dom), nil
}

func handleBrowserWait(params map[string]interface{}) (interface{}, error) {
	selector, _ := params["selector"].(string)
	timeoutSeconds := 2
	if v, ok := params["timeout_seconds"].(float64); ok && v > 0 {
		timeoutSeconds = int(v)
	}
	if v, ok := params["timeout_seconds"].(int); ok && v > 0 {
		timeoutSeconds = v
	}
	dom, err := Wait(selector, timeoutSeconds)
	if err != nil {
		return nil, err
	}
	return domToMap(dom), nil
}

func domToMap(dom *CompressedDOM) map[string]interface{} {
	els := make([]map[string]interface{}, 0, len(dom.Elements))
	for _, e := range dom.Elements {
		m := map[string]interface{}{"ref": e.Ref, "type": e.Type, "text": e.Text, "aria": e.Aria, "place": e.Place}
		if e.Disabled {
			m["disabled"] = true
		}
		if e.Type == "checkbox" || e.Type == "radio" {
			m["checked"] = e.Checked
		}
		if len(e.Options) > 0 {
			m["options"] = e.Options
		}
		els = append(els, m)
	}
	return map[string]interface{}{
		"success": true, "url": dom.URL, "title": dom.Title,
		"elements": els, "text": dom.Text,
	}
}

func handleWebFetch(params map[string]interface{}) (interface{}, error) {
	urlStr, ok := params["url"].(string)
	if !ok || urlStr == "" {
		return nil, fmt.Errorf("url is required")
	}
	maxChars := globalSkillWeb.getFetchMaxChars()
	if v, ok := params["max_chars"].(float64); ok && v > 0 {
		maxChars = int(v)
	}
	if v, ok := params["max_chars"].(int); ok && v > 0 {
		maxChars = v
	}
	if s, ok := params["max_chars"].(string); ok && s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			maxChars = n
		}
	}
	content, err := fetchURL(urlStr, maxChars)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"success": true, "url": urlStr, "content": content, "length": len(content)}, nil
}

var globalSkillWeb *SkillWeb

func init() {
	globalSkillWeb = &SkillWeb{}
	skill.RegisterSkillWithMetadata(
		skillName,
		skill.SkillMetadata{
			Name:        skillName,
			Version:     "0.1.0",
			Description: "Web Skill - 浏览器自动化（rod）与 HTTP 抓取，不依赖 Brave/Serper",
			Author:      "OctoSucker",
			Tags:        []string{"web", "browser", "rod", "fetch"},
		},
		RegisterWebSkill,
		globalSkillWeb,
	)
}
