# skill-web

Web Skill：基于 **go-rod** 的浏览器自动化 + 简单 HTTP 抓取，可选集成 Brave Search API 实现 `web_search` 搜索能力。

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
| `web_search` | 使用 Brave Search API 搜索网络，返回标题、链接与摘要（需要 Brave API Key，可选）。 |

压缩 DOM 中元素现包含 `type`（button/link/input/select/checkbox/radio）、`disabled`、`checked`、`options`（select 的选项列表）；主文本已包含标题、段落、列表项与表格行。

## 配置

可选配置：

```json
{
  "github.com/OctoSucker/skill-web": {
    "fetch_max_chars": 50000,
    "browser_proxy": "http://127.0.0.1:15236",
    "browser_extra_args": ["--proxy-server=http://127.0.0.1:15236"],
    "browser_headless": false,
    "search_api_key": "YOUR_BRAVE_API_KEY",
    "search_count": 5,
    "search_country": "US",
    "search_language": "en",
    "search_freshness": ""
  }
}
```

- `fetch_max_chars`: 默认 50000，web_fetch 与 browser_extract 最多返回字符数
- `browser_proxy`: 可选，浏览器使用的 HTTP 代理（如 Clash 常用 7890）。也可用环境变量 `HTTPS_PROXY` / `HTTP_PROXY`
- `browser_extra_args`: 可选，与 OpenClaw 的 `extraArgs` 一致，直接传给 Chrome（如 `["--proxy-server=http://127.0.0.1:7890"]`）。若含 `--proxy-server` 则优先于 `browser_proxy`
- `browser_headless`: 默认 false。false=有头模式（用户可见 Chrome 窗口，与 OpenClaw 一致）；true=无头模式
- `search_api_key`: 可选，Brave Search API Key，用于 `web_search`。若未配置，则尝试从环境变量 `BRAVE_API_KEY` 读取。
- `search_count`: 可选，`web_search` 默认返回结果条数，默认为 5，最大 10。
- `search_country`: 可选，Brave 搜索地区代码，如 `US`、`ALL`。
- `search_language`: 可选，Brave 搜索语言代码，如 `en`、`zh-hans`，内部映射到 `search_lang`。
- `search_freshness`: 可选，时间范围过滤，如 `day`/`week`/`month`/`year` 或 `YYYY-MM-DDtoYYYY-MM-DD`。

**代理说明**：访问 Google 等需配置代理。常见端口：Clash 7890、SOCKS5 1080。启动时会检测代理可达性，若不可达会打印 WARNING。

浏览器与 `web_fetch` 不需要任何 API Key；只有在使用 `web_search` 时才需要 Brave Search API Key（可选能力）。

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
