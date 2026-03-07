# skill-web

Web Skill：基于 **go-rod** 的浏览器自动化 + 简单 HTTP 抓取，**不依赖 Brave/Serper** 等付费搜索 API。

## 工具

| 工具 | 说明 |
|------|------|
| `browser_navigate` | 在浏览器中打开 URL，返回压缩 DOM（可操作元素 + 主文本）。 |
| `browser_click` | 点击元素（ref 为数字或可见文本、placeholder、aria-label，支持模糊匹配）。 |
| `browser_type` | 在输入框等元素中输入文字。 |
| `browser_scroll` | 滚动页面（up / down / top / bottom）。 |
| `browser_extract` | 从当前页提取主文本。 |
| `browser_hover` | 悬停到元素上，用于触发下拉菜单、tooltip。 |
| `browser_select_option` | 在 select 下拉框中选择选项（value 为选项文本）。 |
| `browser_screenshot` | 截取当前页截图，返回 base64 PNG（可 full_page 长图）。 |
| `browser_go_back` | 浏览器后退。 |
| `browser_go_forward` | 浏览器前进。 |
| `browser_check` | 设置 checkbox/radio 的选中状态（checked true/false）。 |
| `browser_wait` | 等待：可指定 selector 等待元素出现，或仅等待 timeout_seconds。 |
| `web_fetch` | 通过 HTTP 拉取 URL 并提取纯文本（不启动浏览器，适合静态页）。 |

压缩 DOM 中元素现包含 `type`（button/link/input/select/checkbox/radio）、`disabled`、`checked`、`options`（select 的选项列表）；主文本已包含标题、段落、列表项与表格行。

## 配置

可选，仅用于 `fetch_max_chars`（默认 50000）：

```json
{
  "web": {
    "fetch_max_chars": 50000
  }
}
```

无需任何 API Key。

**浏览器**：`browser_*` 工具会优先使用系统已安装的 Chrome/Chromium（通过 `PATH` 查找）。若未安装，rod 会在首次调用时自动下载 Chromium（体积较大、耗时长，且会打印大量 `[launcher.Browser] Progress: XX%`）。建议本机安装 Chrome 或 Chromium 以避免首次下载。

- **自动下载时的存放位置**（rod 源码 `lib/launcher/browser.go`）：
  - macOS / Linux：`$HOME/.cache/rod/browser/chromium-<revision>/`（如 `~/.cache/rod/browser/chromium-1321438/`）
  - Windows：`%APPDATA%\rod\browser\chromium-<revision>\`
  若下载中断（如 unexpected EOF），该目录可能不完整或已被清理。

## 安装

主项目中：

```bash
go get github.com/OctoSucker/skill-web@latest
```

并保留空白导入：`import _ "github.com/OctoSucker/skill-web"`。
