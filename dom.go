package web

import (
	"encoding/json"
	"strings"

	"github.com/go-rod/rod"
)

// CompressedElement 压缩后的可操作元素（button/link/input/select 等）
type CompressedElement struct {
	Ref     string `json:"ref"`     // 稳定引用
	Type     string   `json:"type"`              // button | link | input | select | checkbox | radio 等
	Text     string   `json:"text"`              // 可见文本
	Aria     string   `json:"aria"`              // aria-label
	Place    string   `json:"place"`             // input placeholder
	Disabled bool     `json:"disabled,omitempty"` // 是否不可用
	Checked  bool     `json:"checked,omitempty"` // checkbox/radio 是否选中
	Options  []string `json:"options,omitempty"` // select 的 option 文本列表
}

// CompressedDOM 压缩 DOM：仅保留可操作元素与关键文本，供 AI 选择动作
type CompressedDOM struct {
	URL      string             `json:"url"`
	Title    string             `json:"title"`
	Elements []CompressedElement `json:"elements"`
	Text     string             `json:"text"` // 主文本块摘要（标题+段落，截断）
}

// compressScript 在页面上下文中执行，部分站点（如 Google）会改写 Array.prototype 等，故用 for 循环与 Array.prototype 显式调用，并 try-catch 兜底
const compressScript = `
(function() {
  try {
    var sel = 'a, button, input:not([type="hidden"]), textarea, select, [role="button"], [role="link"], [contenteditable="true"]';
    var nodes = document.querySelectorAll(sel);
    var out = [];
    var i, el, tag, itype, inputType, text, aria, place, disabled, checked, options, j, o, ot;
    for (i = 0; i < nodes.length; i++) {
      el = nodes[i];
      tag = (el.tagName || '').toLowerCase();
      itype = 'link';
      inputType = (el.type || '').toLowerCase();
      if (tag === 'button' || el.getAttribute('role') === 'button') itype = 'button';
      else if (tag === 'select') itype = 'select';
      else if (tag === 'input' || tag === 'textarea') {
        if (inputType === 'checkbox') itype = 'checkbox';
        else if (inputType === 'radio') itype = 'radio';
        else if (inputType === 'submit' || inputType === 'image') itype = 'button';
        else itype = 'input';
      } else if (tag === 'a' || el.getAttribute('role') === 'link') itype = 'link';
      text = (el.innerText || el.value || '').trim().slice(0, 200);
      aria = (el.getAttribute('aria-label') || '').trim().slice(0, 200);
      place = (el.getAttribute('placeholder') || '').trim().slice(0, 200);
      disabled = el.disabled === true;
      checked = el.checked === true;
      options = [];
      if (tag === 'select' && el.options) {
        for (j = 0; j < el.options.length; j++) {
          o = el.options[j];
          ot = (o.textContent || o.text || '').trim().slice(0, 200);
          if (ot) options.push(ot);
        }
      }
      el.setAttribute('data-agent-ref', String(i));
      out.push({ ref: String(i), type: itype, text: text, aria: aria, place: place, disabled: disabled, checked: checked, options: options });
    }
    var title = (document.title || '').trim();
    var url = (window.location && window.location.href) ? window.location.href : '';
    var main = [];
    var list, n, tbl, rows, r, cells, c, rowText, line, maxLen;
    list = document.querySelectorAll('h1, h2, h3, h4, p');
    for (i = 0; i < list.length; i++) {
      n = list[i];
      text = (n.innerText || n.textContent || '').trim();
      if (text.length > 0) main.push(text.slice(0, 500));
    }
    list = document.querySelectorAll('ul li, ol li');
    for (i = 0; i < list.length; i++) {
      n = list[i];
      text = (n.innerText || n.textContent || '').trim();
      if (text.length > 0) main.push(text.slice(0, 300));
    }
    list = document.querySelectorAll('table');
    for (i = 0; i < list.length; i++) {
      tbl = list[i];
      rows = tbl.querySelectorAll('tr');
      for (r = 0; r < Math.min(rows.length, 50); r++) {
        cells = rows[r].querySelectorAll('th, td');
        rowText = [];
        for (c = 0; c < cells.length; c++) rowText.push((cells[c].innerText || '').trim().slice(0, 100));
        line = rowText.join(' | ');
        if (line.length > 0) main.push(line.slice(0, 500));
      }
    }
    var textBlock = main.length ? Array.prototype.join.call(main, '\\n\\n').slice(0, 12000) : '';
    return JSON.stringify({ url: url, title: title, elements: out, text: textBlock });
  } catch (e) {
    var u = (window.location && window.location.href) ? window.location.href : '';
    var t = (document.title || '').trim();
    var errMsg = (e && e.message) ? String(e.message) : 'unknown';
    return JSON.stringify({ url: u, title: t, elements: [], text: '(DOM compress error: ' + errMsg + ')' });
  }
})();
`

// CompressDOM 在当前页执行压缩脚本，返回可操作元素列表与主文本
func CompressDOM(p *rod.Page) (*CompressedDOM, error) {
	res, err := p.Eval(compressScript)
	if err != nil {
		return nil, err
	}
	raw := res.Value.Str()
	raw = strings.Trim(raw, `"`)
	raw = strings.ReplaceAll(raw, `\"`, `"`)
	var data struct {
		URL      string             `json:"url"`
		Title    string             `json:"title"`
		Elements []CompressedElement `json:"elements"`
		Text     string             `json:"text"`
	}
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		return nil, err
	}
	return &CompressedDOM{
		URL:      data.URL,
		Title:    data.Title,
		Elements: data.Elements,
		Text:     data.Text,
	}, nil
}
