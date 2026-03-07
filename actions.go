package web

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-rod/rod"
)

// Navigate 打开 URL 并返回压缩 DOM
func Navigate(url string) (_ *CompressedDOM, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("navigate: %v", r)
		}
	}()
	var dom *CompressedDOM
	err = withPage(func(p *rod.Page) error {
		p.MustSetExtraHeaders("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) OctoSucker-skill-web/1.0")
		if navErr := p.Navigate(url); navErr != nil {
			return fmt.Errorf("navigate: %w", navErr)
		}
		if waitErr := p.WaitLoad(); waitErr != nil {
			if !strings.Contains(waitErr.Error(), "navigated or closed") && !strings.Contains(waitErr.Error(), "-32000") {
				return fmt.Errorf("wait load: %w", waitErr)
			}
			// 重定向时 WaitLoad 可能报 -32000，稍等后仍尝试取 DOM
		}
		time.Sleep(500 * time.Millisecond)
		var innerErr error
		dom, innerErr = CompressDOM(p)
		return innerErr
	})
	if err != nil {
		return nil, err
	}
	return dom, nil
}

// Click 点击元素（ref 为压缩 DOM 的 ref 或可见文本/aria）
func Click(ref string) (*CompressedDOM, error) {
	var dom *CompressedDOM
	err := withPage(func(p *rod.Page) error {
		el, err := FindElement(p, ref)
		if err != nil {
			return err
		}
		el.MustScrollIntoView()
		el.MustClick()
		time.Sleep(300 * time.Millisecond)
		var innerErr error
		dom, innerErr = CompressDOM(p)
		return innerErr
	})
	if err != nil {
		return nil, err
	}
	return dom, nil
}

// Type 在目标元素输入文字（ref 同 Click）
func Type(ref, text string) (*CompressedDOM, error) {
	var dom *CompressedDOM
	err := withPage(func(p *rod.Page) error {
		el, err := FindElement(p, ref)
		if err != nil {
			return err
		}
		el.MustScrollIntoView()
		el.MustInput(text)
		time.Sleep(300 * time.Millisecond)
		var innerErr error
		dom, innerErr = CompressDOM(p)
		return innerErr
	})
	if err != nil {
		return nil, err
	}
	return dom, nil
}

// Scroll 滚动页面
func Scroll(direction string) (*CompressedDOM, error) {
	var dom *CompressedDOM
	err := withPage(func(p *rod.Page) error {
		body, err := p.Element("body")
		if err != nil {
			return err
		}
		switch direction {
		case "up":
			body.MustEval("window.scrollBy(0, -400)")
		case "top":
			body.MustEval("window.scrollTo(0, 0)")
		case "down":
			body.MustEval("window.scrollBy(0, 400)")
		case "bottom":
			body.MustEval("window.scrollTo(0, document.body.scrollHeight)")
		default:
			body.MustEval("window.scrollBy(0, 400)")
		}
		time.Sleep(200 * time.Millisecond)
		var innerErr error
		dom, innerErr = CompressDOM(p)
		return innerErr
	})
	if err != nil {
		return nil, err
	}
	return dom, nil
}

// Extract 从当前页提取主文本（压缩 DOM 的 text 字段，或再取 body innerText）
func Extract(maxChars int) (string, error) {
	var text string
	err := withPage(func(p *rod.Page) error {
		dom, err := CompressDOM(p)
		if err != nil {
			return err
		}
		text = dom.Text
		if len(text) > maxChars {
			text = text[:maxChars]
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return text, nil
}

// Snapshot 仅返回当前压缩 DOM，不执行动作
func Snapshot() (*CompressedDOM, error) {
	var dom *CompressedDOM
	err := withPage(func(p *rod.Page) error {
		var innerErr error
		dom, innerErr = CompressDOM(p)
		return innerErr
	})
	if err != nil {
		return nil, err
	}
	return dom, nil
}

// Hover 悬停到元素上（用于触发下拉菜单、tooltip 等）
func Hover(ref string) (*CompressedDOM, error) {
	var dom *CompressedDOM
	err := withPage(func(p *rod.Page) error {
		el, err := FindElement(p, ref)
		if err != nil {
			return err
		}
		el.MustScrollIntoView()
		if err := el.Hover(); err != nil {
			return err
		}
		time.Sleep(300 * time.Millisecond)
		var innerErr error
		dom, innerErr = CompressDOM(p)
		return innerErr
	})
	if err != nil {
		return nil, err
	}
	return dom, nil
}

// SelectOption 在 select 元素中选择 option（value 为选项的可见文本或 value 属性）
func SelectOption(ref, value string) (*CompressedDOM, error) {
	var dom *CompressedDOM
	err := withPage(func(p *rod.Page) error {
		el, err := FindElement(p, ref)
		if err != nil {
			return err
		}
		el.MustScrollIntoView()
		el.MustSelect(value)
		time.Sleep(200 * time.Millisecond)
		var innerErr error
		dom, innerErr = CompressDOM(p)
		return innerErr
	})
	if err != nil {
		return nil, err
	}
	return dom, nil
}

// Screenshot 截取当前页面图，返回 base64 PNG。fullPage 为 true 时截整页。
func Screenshot(fullPage bool) ([]byte, error) {
	var data []byte
	err := withPage(func(p *rod.Page) error {
		if fullPage {
			data = p.MustScreenshotFullPage()
		} else {
			data = p.MustScreenshot()
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return data, nil
}

// GoBack 浏览器后退
func GoBack() (*CompressedDOM, error) {
	var dom *CompressedDOM
	err := withPage(func(p *rod.Page) error {
		if err := p.NavigateBack(); err != nil {
			return fmt.Errorf("go back: %w", err)
		}
		p.MustWaitLoad()
		time.Sleep(400 * time.Millisecond)
		var innerErr error
		dom, innerErr = CompressDOM(p)
		return innerErr
	})
	if err != nil {
		return nil, err
	}
	return dom, nil
}

// GoForward 浏览器前进
func GoForward() (*CompressedDOM, error) {
	var dom *CompressedDOM
	err := withPage(func(p *rod.Page) error {
		if err := p.NavigateForward(); err != nil {
			return fmt.Errorf("go forward: %w", err)
		}
		p.MustWaitLoad()
		time.Sleep(400 * time.Millisecond)
		var innerErr error
		dom, innerErr = CompressDOM(p)
		return innerErr
	})
	if err != nil {
		return nil, err
	}
	return dom, nil
}

// Check 设置 checkbox 或 radio 的选中状态（checked=true 选中，false 取消）
func Check(ref string, checked bool) (*CompressedDOM, error) {
	var dom *CompressedDOM
	err := withPage(func(p *rod.Page) error {
		el, err := FindElement(p, ref)
		if err != nil {
			return err
		}
		el.MustScrollIntoView()
		res, err := el.Eval("() => this.checked")
		if err != nil {
			return fmt.Errorf("element is not a checkbox/radio: %w", err)
		}
		cur := res.Value.Bool()
		if cur != checked {
			el.MustClick()
		}
		time.Sleep(200 * time.Millisecond)
		var innerErr error
		dom, innerErr = CompressDOM(p)
		return innerErr
	})
	if err != nil {
		return nil, err
	}
	return dom, nil
}

// Wait 等待：若 selector 非空则等待该元素出现（最多 waitSeconds 秒）；否则仅等待 waitSeconds 秒
func Wait(selector string, waitSeconds int) (*CompressedDOM, error) {
	if waitSeconds <= 0 {
		waitSeconds = 2
	}
	timeout := time.Duration(waitSeconds) * time.Second
	var dom *CompressedDOM
	err := withPage(func(p *rod.Page) error {
		if selector != "" {
			p = p.Timeout(timeout)
			_, err := p.Element(selector)
			if err != nil {
				return fmt.Errorf("wait for selector %q: %w", selector, err)
			}
		} else {
			time.Sleep(timeout)
		}
		var innerErr error
		dom, innerErr = CompressDOM(p)
		return innerErr
	})
	if err != nil {
		return nil, err
	}
	return dom, nil
}
