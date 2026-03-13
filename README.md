# tools-web

Web Tool Provider：HTTP 抓取、Jina Reader 转 Markdown、Brave Search API 搜索。不包含浏览器自动化（无 go-rod）。

## 工具

| 工具 | 说明 |
|------|------|
| `web_fetch` | 通过 HTTP 拉取 URL 的纯文本（不启动浏览器，适合静态页）。参数：`url`（必填）、`max_chars`（可选）。 |
| `web_read` | 使用 [Jina Reader](https://r.jina.ai) 将任意 URL 转为适合 LLM 的 Markdown 文本，仅用于阅读网页内容。参数：`url`（必填）、`respond_with`（可选，如 markdown/html/text）。 |
| `web_search` | 使用 Brave Search API 搜索网络，返回标题、链接与摘要。参数：`query`（必填）、`count`/`country`/`language`/`freshness`（可选）。需配置 `search_api_key` 或环境变量 `BRAVE_API_KEY`。 |

## 配置

在 Agent 配置的 `tool_providers["github.com/OctoSucker/tools-web"]` 下：

| 键 | 说明 |
|------|------|
| `fetch_max_chars` | `web_fetch` 最大返回字符数，默认 50000。 |
| `search_api_key` | Brave Search API Key；未配置时尝试环境变量 `BRAVE_API_KEY`。未设置时 `web_search` 会返回明确错误与提示。 |
| `search_count` | `web_search` 默认返回条数，默认 5，最大 10。 |
| `search_country` | Brave 搜索地区代码，如 `US`、`ALL`，默认 `US`。 |
| `search_language` | Brave 搜索语言代码，如 `en`、`zh-hans`。 |
| `search_freshness` | 时间范围：`day`/`week`/`month`/`year` 或 `YYYY-MM-DDtoYYYY-MM-DD`。 |

示例（`config/agent_config.json`）：

```json
"github.com/OctoSucker/tools-web": {
  "fetch_max_chars": 50000,
  "search_api_key": "YOUR_BRAVE_API_KEY",
  "search_count": 5,
  "search_country": "US"
}
```

- `web_fetch` 与 `web_read` 不需要 API Key。
- 仅在使用 `web_search` 时需要 Brave Search API Key（在 [Brave Search API](https://brave.com/search/api/) 申请）。

## 安装

主项目中：

```bash
go get github.com/OctoSucker/tools-web@latest
```

并保留空白导入：`_ "github.com/OctoSucker/tools-web"`。
