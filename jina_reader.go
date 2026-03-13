package web

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func handleWebRead(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	rawURL, ok := params["url"].(string)
	if !ok || strings.TrimSpace(rawURL) == "" {
		return nil, fmt.Errorf("url is required")
	}

	respondWith, _ := params["respond_with"].(string)

	return webRead(ctx, strings.TrimSpace(rawURL), strings.TrimSpace(respondWith))
}

func webRead(ctx context.Context, rawURL, respondWith string) (map[string]interface{}, error) {
	if rawURL == "" {
		return nil, fmt.Errorf("url is required")
	}

	encoded := url.QueryEscape(rawURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://r.jina.ai/"+encoded, nil)
	if err != nil {
		return nil, err
	}
	if respondWith != "" {
		req.Header.Set("X-Respond-With", respondWith)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("reader error: status %d, body: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"success": true,
		"url":     rawURL,
		"content": string(body),
	}, nil
}
