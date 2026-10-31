# Dinosaur Agent API

Agent 在自己的环境中查找资料、判断匹配并下载海报，通过本 API 写入 Dinosaur。只需要 NAS 地址和独立 Agent API Key；无需浏览器登录、CSRF Token 或在 NAS 配置每个资料网站的凭据。

## 获取 Key

管理员在“账户安全 → Agent API Key”创建 Key，选择媒体库、读取/读写权限及有效期。完整 Key 仅在创建时返回一次，数据库仅保存哈希。列表中可以撤销 Key。Key 只能用于 `/api/agent/`，不能播放视频、修改账户、删除媒体或访问服务器文件。

程序管理接口需要原有管理员会话：

- `GET /api/admin/agent-keys`：仅列出当前管理员的 Key 元信息。
- `POST /api/admin/agent-keys`：`{"name":"识别 Agent","library_id":0,"access":"read_write","expires_in_days":90}`。`library_id:0` 表示全部媒体库，`access` 为 `read` 或 `read_write`，有效期 1–365 天；返回 `key` 与一次性的 `token`。
- `DELETE /api/admin/agent-keys/{id}`：撤销自己的 Key。

所有 Agent 请求使用：

```http
Authorization: Bearer dino_agent_完整的Key
```

不要同时发送浏览器 Cookie，不接受查询参数中的 Key。Key 过期、撤销、所属管理员被禁用或降为普通用户后，认证失败。删除指定媒体库会同时撤销对应 Key，不会扩大为全部媒体库权限。每个 Key 最多 1000 请求 / 15 分钟。

## 读取任务与上下文

| 请求 | 返回 |
| --- | --- |
| `GET /api/agent/help` | 本文档 |
| `GET /api/agent/tasks` | 默认包含未识别、待确认和失败的电影/整剧，分页 |
| `GET /api/agent/movies/{id}` | 电影的名称、年份、媒体库、路径、技术信息、现有资料与版本 |
| `GET /api/agent/series/{id}` | 整剧上下文及分集文件列表 |
| `GET /api/agent/movies/{id}/metadata` | 可直接编辑并回传的 JSON 资料；ETag 为当前版本 |
| `GET /api/agent/series/{id}/metadata` | 整剧 JSON 资料；ETag 为当前版本 |

任务筛选支持 `library_id`、`q`、`page`、`page_size`（最多 100）。`status` 默认为 `pending`，可用 `all`、`matched`、`unmatched`、`needs_review`、`failed`、`warning`。

任务中的 `media_id` 是电影目标，`series_id` 是整剧目标；整剧的 `media_id` 为 0，不能把分集播放 ID 当作整剧 ID。`version` 是当前资料版本；`relative_path` 为代表文件路径。详情提供 `library_root` 和文件的相对 `path`。整剧最多返回前 1000 个文件，实际数量见 `file_count`。

## 提交 JSON

```http
PUT /api/agent/movies/123/metadata
Authorization: Bearer dino_agent_完整的Key
Content-Type: application/json
If-Match: "0"

{
  "title": "菊次郎的夏天",
  "original_title": "菊次郎の夏",
  "year": 1999,
  "release_date": "1999-06-05",
  "overview": "经核实的剧情简介。",
  "language": "zh-CN",
  "original_language": "ja",
  "aliases": ["Kikujiro"],
  "ids": {"tmdb": "4291", "imdb": "tt0199683"},
  "source_url": "https://www.themoviedb.org/movie/4291"
}
```

整剧使用 `PUT /api/agent/series/{id}/metadata`，字段相同。`title` 必填，其余可省略；`ids` 可为空，不强制依赖任何外部网站 ID。ID 值使用字符串，支持 `tmdb`、`bangumi`、`imdb` 等命名空间。上映日期为 `YYYY-MM-DD`，语言使用 `ja`、`zh-CN` 等代码，不根据中文片名推测原声语言。

PUT 是整份资料替换：省略的文字、别名和外部 ID 会被清空，现有海报保留。局部修改应先 GET 当前 JSON，再编辑、回传。正文最多 128 KiB，简介最多 32000 字节。

必须把读取时的 ETag（或任务 `version` 加双引号）放入 `If-Match`。示例的 `"0"` 只是示意，不应固定使用。成功返回新 `version` 和新 ETag。其他用户、Agent、刮削或资料提示状态更新都会推进版本；过期写入返回 409，需重新读取并判断，不能盲目重复旧写入。

成功提交后来源为 `agent`，标记已确认。普通扫描、自动刮削及已在运行的自动识别结果都不会覆盖它。更新已确认资料仍使用同一 API 与最新版本。保留本地媒体 ID、文件和观看记录。

## NFO 导入与导出

- `GET /api/agent/movies/{id}/metadata.nfo`
- `GET /api/agent/series/{id}/metadata.nfo`
- 同一个资料 PUT 接口将 `Content-Type` 改为 `application/xml`（也接受 `text/xml`），原始正文为 NFO XML；同样要求 `If-Match`。

支持 Kodi NFO 的核心子集：电影根节点 `movie`，整剧根节点 `tvshow`，字段 `title`、`originaltitle`、`plot`、`year`、`premiered`、`uniqueid`。`uniqueid type="themoviedb"` 归一化为 `tmdb`。

Dinosaur 扩展放在 `<dinosaur>` 内：`language`、`original_language`、`source_url`、可重复的 `alias`。导出会对文本进行 XML 转义，支持 UTF-8。导入成功来源为 `nfo`，具有与 Agent JSON 相同的确认、版本和覆盖保护。

不导入观看历史、播放地址或图片 URL；不解析外部实体、DTD 或外部资源。不处理单集剧情、演职员表和其他未列出的 NFO 字段。NFO 导入/导出只通过 HTTP 传输，不会在媒体目录自动写文件或扫描读取已有 NFO。

格式参考：[Kodi 电影 NFO](https://kodi.wiki/view/NFO_files/Movies)、[Kodi 整剧 NFO](https://kodi.wiki/view/NFO_files/TV_shows)。

## 上传海报

先确认资料，再执行：

```http
POST /api/agent/movies/123/artwork
Authorization: Bearer dino_agent_完整的Key
Content-Type: image/jpeg
If-Match: "1"

JPEG 文件原始字节
```

整剧使用 `/api/agent/series/{id}/artwork`。支持 JPEG/PNG，最大 5 MiB、8 百万像素、边长最多 4000。服务端转换为本地 JPEG，成功后推进资料版本并标记为 Agent 管理。`GET` 同一地址可用 Key 查看海报。图片由 Agent 下载后上传，Dinosaur 不抓取任意远程图片 URL。

## 最小调用示例

```sh
# 在进程环境中提供 DINO_API_KEY，避免把完整 Key 写进提示词、日志或仓库。
DINO_URL='https://dinosaur.monkeydog.com'
curl -fsS -H "Authorization: Bearer $DINO_API_KEY" \
  "$DINO_URL/api/agent/tasks"
curl -fsS -D metadata.headers -o metadata.json \
  -H "Authorization: Bearer $DINO_API_KEY" \
  "$DINO_URL/api/agent/movies/123/metadata"
# 识别并编辑 metadata.json；将 metadata.headers 中的 ETag 用于 If-Match。
curl -fsS -X PUT \
  -H "Authorization: Bearer $DINO_API_KEY" \
  -H 'Content-Type: application/json' -H 'If-Match: "当前版本"' \
  --data-binary @metadata.json "$DINO_URL/api/agent/movies/123/metadata"
```

读取任务 → 核对文件和资料来源 → 提交可信结果 → 上传海报。无法确定时保留任务，避免猜测后提交为已确认。

## 状态码

- 401：缺少 Key、凭据混用、过期、撤销或所属账号失效。
- 403：只读 Key 写入，或访问超出媒体库范围。
- 404：目标不存在、缺失，或把分集 ID 当作电影/整剧目标。
- 409：版本冲突；海报上传前尚未确认资料也返回 409。
- 428：缺少 `If-Match`。
- 400/413/415：字段或图片无效、请求过大、不支持的格式。
- 429：超过请求频率；按响应 `Retry-After` 重试。
