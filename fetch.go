package web

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"golang.org/x/net/html"
)

var (
	fetchClientMu sync.Mutex
	fetchClient   *http.Client
	fetchProxyURL string
)

func SetFetchProxy(proxyURL string) {
	fetchClientMu.Lock()
	fetchProxyURL = proxyURL
	fetchClient = nil
	fetchClientMu.Unlock()
}

func handleWebFetch(ctx context.Context, params map[string]interface{}) (interface{}, error) {
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

func getFetchClient() *http.Client {
	fetchClientMu.Lock()
	defer fetchClientMu.Unlock()
	if fetchClient != nil {
		return fetchClient
	}
	proxyURL := fetchProxyURL
	if proxyURL == "" {
		proxyURL = os.Getenv("HTTPS_PROXY")
	}
	if proxyURL == "" {
		proxyURL = os.Getenv("HTTP_PROXY")
	}
	transport := &http.Transport{}
	if proxyURL != "" {
		if u, err := url.Parse(proxyURL); err == nil {
			transport.Proxy = http.ProxyURL(u)
		}
	}
	fetchClient = &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}
	fetchClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 5 {
			return fmt.Errorf("too many redirects")
		}
		return nil
	}
	return fetchClient
}

const defaultFetchMaxChars = 50000

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
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := getFetchClient().Do(req)
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
