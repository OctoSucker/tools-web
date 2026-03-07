package web

import (
	"log"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

var (
	browserMu sync.Mutex
	browser   *rod.Browser
	page      *rod.Page
)

const (
	browserTimeout = 30 * time.Second
	pageLoadWait   = 5 * time.Second
)

// getPage 返回当前页，若无则先启动 browser 并创建空白页。仅应在 withPage 内调用（调用方必须已持 browserMu）。
func getPage() (*rod.Page, error) {
	if page != nil {
		return page, nil
	}
	if err := ensureBrowser(); err != nil {
		return nil, err
	}
	p, err := browser.MustIncognito().Page(proto.TargetCreateTarget{URL: "about:blank"})
	if err != nil {
		return nil, err
	}
	p.MustSetExtraHeaders("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) OctoSucker-skill-web/1.0")
	page = p
	return page, nil
}

// withPage 在持锁下获取 page 并执行 fn，保证同一时刻只有一个 goroutine 操作 page，避免 rod 并发 panic。
func withPage(fn func(*rod.Page) error) error {
	browserMu.Lock()
	defer browserMu.Unlock()
	p, err := getPage()
	if err != nil {
		return err
	}
	return fn(p)
}

func ensureBrowser() error {
	if browser != nil {
		return nil
	}
	l := launcher.New().
		Headless(true).
		Set("no-sandbox").
		Set("disable-setuid-sandbox")

	// 优先使用系统已安装的 Chrome/Chromium，避免首次启动时下载数百 MB 的 Chromium（否则会长时间打印 [launcher.Browser] Progress: XX%）
	if path, ok := launcher.LookPath(); ok && path != "" {
		l = l.Bin(path)
		log.Printf("skill-web: using system browser: %s", path)
	} else {
		log.Printf("skill-web: no system browser found, launcher may download Chromium on first use (this can take several minutes)")
	}

	u, err := l.Launch()
	if err != nil {
		return err
	}
	browser = rod.New().ControlURL(u).MustConnect()
	browser.MustSetCookies() // 无参数表示清空 cookie，避免携带系统浏览器残留
	log.Printf("skill-web: browser started (rod)")
	return nil
}

// closePage 关闭当前页（用于 navigate 到新页前可复用，或 Cleanup 时调用）
func closePage() {
	browserMu.Lock()
	defer browserMu.Unlock()
	if page != nil {
		_ = page.Close()
		page = nil
	}
}

// closeBrowser 关闭浏览器（Skill Cleanup 时调用）
func closeBrowser() {
	browserMu.Lock()
	defer browserMu.Unlock()
	if page != nil {
		_ = page.Close()
		page = nil
	}
	if browser != nil {
		_ = browser.Close()
		browser = nil
	}
	log.Printf("skill-web: browser closed")
}
