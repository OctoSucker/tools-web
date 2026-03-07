package web

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode"

	"golang.org/x/net/html"
)

var fetchClient = &http.Client{Timeout: 30 * time.Second}

func init() {
	fetchClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 5 {
			return fmt.Errorf("too many redirects")
		}
		return nil
	}
}

const defaultFetchMaxChars = 50000

// fetchURL 拉取 URL 内容并提取为纯文本（从 HTML 提取）。maxChars 为 0 时使用默认 50000。
func fetchURL(rawURL string, maxChars int) (string, error) {
	if rawURL == "" {
		return "", fmt.Errorf("url is required")
	}
	if maxChars <= 0 {
		maxChars = defaultFetchMaxChars
	}

	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "OctoSucker-skill-web/1.0")

	resp, err := fetchClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fetch returned %d", resp.StatusCode)
	}

	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "text/html") && !strings.Contains(ct, "text/plain") && ct != "" {
		// 非 HTML/文本可限制读取量
		body, err := io.ReadAll(io.LimitReader(resp.Body, int64(maxChars)))
		if err != nil {
			return "", err
		}
		return string(body), nil
	}

	limited := io.LimitReader(resp.Body, int64(maxChars*2)) // 原始 HTML 比纯文本长
	doc, err := html.Parse(limited)
	if err != nil {
		return "", fmt.Errorf("parse HTML: %w", err)
	}

	var b strings.Builder
	extractText(doc, &b)
	out := b.String()
	if len(out) > maxChars {
		out = out[:maxChars]
	}
	return strings.TrimSpace(out), nil
}

func extractText(n *html.Node, w *strings.Builder) {
	if n.Type == html.TextNode {
		s := strings.TrimSpace(n.Data)
		if s != "" {
			if w.Len() > 0 {
				w.WriteByte(' ')
			}
			w.WriteString(normalizeSpace(s))
		}
		return
	}
	if n.Type != html.ElementNode {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extractText(c, w)
		}
		return
	}
	switch n.Data {
	case "script", "style", "head":
		return
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		extractText(c, w)
	}
}

func normalizeSpace(s string) string {
	var b strings.Builder
	prev := true
	for _, r := range s {
		if unicode.IsSpace(r) {
			if !prev {
				b.WriteRune(' ')
				prev = true
			}
		} else {
			b.WriteRune(r)
			prev = false
		}
	}
	return strings.TrimSpace(b.String())
}
