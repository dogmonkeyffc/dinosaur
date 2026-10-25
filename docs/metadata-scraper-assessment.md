# 中文电影元数据刮削评估

评估日期：2026-10-24；实现状态更新：2026-10-25。TMDB 凭据配置、电影资料刮削、后台任务、中文展示、海报缓存、候选更正与已匹配资料复用已实现；本地 NFO、字段人工编辑与演员资料仍是后续计划。下文保留完整选型与目标方案，当前使用说明以 README 为准。

## 建议

dinosaur 内置刮削使用 TMDB 官方 API，默认请求 `zh-CN`，后续计划兼容本地 NFO。若用户希望先通过独立工具批量整理，可优先试用 tinyMediaManager；若要求独立工具开源，可评估 MediaElch。MoviePilot 的自动化范围更广，单为中文片名引入它偏重。

TMDB 是数据源，不是可直接嵌入的开源刮削程序。Go 可以通过 HTTP 适配器访问，无需为了这一功能新增 Java/Python 运行服务。中文覆盖、译名偏好及实际网络质量仍需用用户片库实测，不能保证每部影片都有简体中文信息。

## 方案比较

| 方案 | 作用及优势 | 本项目的代价 | 判断 |
| --- | --- | --- | --- |
| TMDB 官方 API | 搜索原名、译名、别名，通过语言参数取本地化资料 | 需要 API 凭据、匹配流程、缓存、重试及来源署名 | 推荐作为内置数据源 |
| tinyMediaManager | 独立媒体整理工具，可搜索匹配、编辑并输出 NFO；免费版包含 TMDB 刮削 | 额外维护一个工具；dinosaur 需新增 NFO 读取，当前不会自动读取 | 推荐作为现成工具试用 |
| MediaElch | 开源桌面媒体管理器，使用 NFO，可作为人工整理入口 | 同样需要 NFO 适配；中文结果取决于数据源与语言设置 | 开源独立工具备选 |
| MoviePilot | 覆盖订阅、搜索、下载、整理、刮削与刷新 | 引入完整自动化系统及其配置 | 已有 NAS 自动化需求时考虑 |

豆瓣网页刮削暂不作为首版主依赖。这里没有验证可用于本项目的官方接口或第三方实现，不对其稳定性作承诺；后续可以作为独立 provider 专门评估。

## 当前项目需要调整的地方

- 当前 `media.ParseFilename` 从文件名提取名称与年份，没有中文资料来源；`ffprobe` 负责时长、编码、音轨，不负责电影译名。
- `Store.SaveMedia` 的更新语句会覆盖 `title` 和 `year`。文件变动后的重新扫描可能覆盖人工或在线匹配结果，不能仅增加一次在线请求。
- 将文件信息、技术信息、影片资料分开保存。建议独立 `metadata` 领域与持久化表，按现有 media ID 关联，保留观看记录及文件路径。
- 资料保存 `display_title`、`original_title`、别名、简介、发行日期、provider ID、语言、来源、更新时间和人工锁定状态。
- 扫描更新技术信息；仅未识别影片进入自动识别，已匹配影片复用 provider ID 和资料，用户主动刷新或重新匹配时才更新；HTTP 层只转换请求和检查权限。领域模型不依赖 HTTP 客户端或数据库。

## 中文展示与匹配规则

以下是本项目建议规则，不是 TMDB 保证的自动回退行为：

1. 展示优先级：人工锁定的名称 → 本地 NFO 标题 → TMDB 简体中文标题 → 其他中文译名 → 原名 → 文件名。
2. 默认语言 `zh-CN`；缺少中文资料时显式查其他中文翻译或别名。地区译名有差异，不能用繁简转换替代影片识别，也不能仅因 `zh-CN` 返回非空就断言是中文标题。
3. 已有 TMDB/IMDb ID 时先按 ID 定位；否则使用清理后的片名、年份查询并比较候选。英文文件名可以参与搜索，不需要先机器翻译。
4. 同名、翻拍、年份冲突或多个接近候选时标记“待确认”，让用户选择；不盲取第一个结果。支持输入 TMDB ID 纠正并锁定。
5. 中文片名、原名和别名均参与搜索；匹配后不修改电影文件名或路径。
6. 海报优先中文，无中文海报时允许其他语言；下载失败保留现有抽帧图。

## 接入与验收顺序

1. 增加影片资料存储和人工编辑/锁定，首先保证重扫不覆盖中文标题。
2. 实现只读 NFO 适配：支持同名 `.nfo`；仅单电影目录允许 `movie.nfo`，避免当前平铺电影目录错误套用同一份资料。
3. 实现 TMDB provider 和后台刮削任务，提供语言、凭据、网络连通性测试、匹配候选与重新匹配入口。网络失败不阻断扫描、列表或播放。
4. 首版覆盖中文名称、原名、年份、简介、海报与 TMDB ID，演员和合集后续扩展。
5. 用现有三部电影验证：候选对应正确影片及年份、中文译名可接受、英文及中文搜索有效、重扫和重启保留结果、匹配更正后观看进度不变。
6. 补充同名翻拍、缺少中文、无网络、NFO 修改但视频未变等用例。

资料与下载海报继续放 `~/.cache/dinosaur/`，原始视频保留原目录。缓存需要刷新和失败重试策略；用户锁定字段不自动覆盖。API 凭据只在服务端保存，前端只显示是否已配置。

## 使用条件与验证边界

TMDB FAQ 说明非商业使用可免费申请 API，并要求数据与图片来源署名。集成时按其当前规则展示标志与声明。初次评估时未实测中文覆盖率；实现后使用隔离的上游响应验证了中文回退、模糊匹配、海报缓存、重扫复用及故障保留。实际影片结果还取决于所配置凭据与网络环境。

## 官方参考

- [TMDB 语言参数](https://developer.themoviedb.org/docs/languages)
- [TMDB 数据查找：原名、译名、别名与外部 ID](https://developer.themoviedb.org/docs/finding-data)
- [TMDB 电影搜索接口](https://developer.themoviedb.org/reference/search-movie)
- [TMDB API 申请、费用与署名](https://developer.themoviedb.org/docs/faq)
- [tinyMediaManager 免费版与 PRO 功能](https://www.tinymediamanager.org/help/faq)
- [tinyMediaManager 匹配和 NFO 工作流](https://www.tinymediamanager.org/docs/quickstart)
- [MediaElch 源码](https://github.com/Komet/MediaElch)
- [MediaElch NFO 说明](https://mediaelch.github.io/mediaelch-doc/about.html)
- [MoviePilot 项目](https://github.com/jxxghp/MoviePilot)
- [Jellyfin 本地 NFO 兼容格式](https://jellyfin.org/docs/general/server/metadata/nfo/)
