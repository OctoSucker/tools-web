package web

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-rod/rod"
)

// escapeXPathString 转义 XPath 字符串中的单引号，返回可嵌入 XPath 的表达式（如 'x' 或 concat('a',\"'\",'b')）
func escapeXPathString(s string) string {
	if s == "" {
		return "''"
	}
	if !strings.Contains(s, "'") {
		return "'" + s + "'"
	}
	parts := strings.Split(s, "'")
	var b strings.Builder
	b.WriteString("concat(")
	for i, p := range parts {
		if i > 0 {
			b.WriteString(", \"'\", ")
		}
		b.WriteString("'")
		b.WriteString(strings.ReplaceAll(p, "\\", "\\\\"))
		b.WriteString("'")
	}
	b.WriteString(")")
	return b.String()
}

// escapeJSRegex 将 ref 转为可在 JS regex 中用于“包含”匹配的表达式（转义特殊字符后包在 .* 中）
func escapeJSRegex(s string) string {
	escaped := regexp.QuoteMeta(s)
	return ".*" + escaped + ".*"
}

// FindElement 按 ref 或文本/aria/placeholder 查找元素，支持 selector 自动恢复与模糊匹配
func FindElement(p *rod.Page, ref string) (*rod.Element, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return nil, fmt.Errorf("ref is required")
	}
	refEscaped := strings.ReplaceAll(ref, `\`, `\\`)
	refEscaped = strings.ReplaceAll(refEscaped, `"`, `\"`)

	// 1) 稳定引用：data-agent-ref
	if matched, _ := regexp.MatchString(`^\d+$`, ref); matched {
		el, err := p.Element("[data-agent-ref=\"" + ref + "\"]")
		if err == nil {
			return el, nil
		}
	}
	if el, err := p.Element("[data-agent-ref=\"" + refEscaped + "\"]"); err == nil {
		return el, nil
	}

	// 2) 文本匹配：先精确，再包含（模糊）
	jsExact := regexp.QuoteMeta(ref)
	for _, tag := range []string{"button", "a", "input", "[role=button]", "[role=link]"} {
		el, err := p.ElementR(tag, jsExact)
		if err == nil {
			return el, nil
		}
	}
	el, err := p.ElementR("*", jsExact)
	if err == nil {
		return el, nil
	}
	// 模糊：包含 ref 的文本
	jsContains := escapeJSRegex(ref)
	for _, tag := range []string{"button", "a", "[role=button]", "[role=link]"} {
		el, err := p.ElementR(tag, jsContains)
		if err == nil {
			return el, nil
		}
	}
	el, err = p.ElementR("*", jsContains)
	if err == nil {
		return el, nil
	}

	// 3) placeholder 查找
	el, err = p.Element("[placeholder=\"" + refEscaped + "\"]")
	if err == nil {
		return el, nil
	}
	el, err = p.ElementR("input[placeholder], textarea[placeholder]", jsContains)
	if err == nil {
		return el, nil
	}

	// 4) aria-label
	el, err = p.Element("[aria-label=\"" + refEscaped + "\"]")
	if err == nil {
		return el, nil
	}
	el, err = p.ElementR("[aria-label]", jsExact)
	if err == nil {
		return el, nil
	}
	el, err = p.ElementR("[aria-label]", jsContains)
	if err == nil {
		return el, nil
	}

	// 5) XPath fallback（转义单引号）
	xpathEscaped := escapeXPathString(ref)
	xpath := "//*[contains(text()," + xpathEscaped + ") or contains(.," + xpathEscaped + ")]"
	el, err = p.ElementX(xpath)
	if err == nil {
		return el, nil
	}

	return nil, fmt.Errorf("element not found: %q (tried ref, text, placeholder, aria, xpath)", ref)
}
