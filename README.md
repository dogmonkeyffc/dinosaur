# Dinosaur

使用 Go 开发的自托管媒体服务器，功能目标参考 Jellyfin。目前已完成用户认证、Web 系统设置、媒体库后台扫描、电影与电视剧展示和浏览器播放。电视剧功能见 [使用说明](docs/television.md)，其他规划见 [task.md](task.md)。API 为项目自有协议。

电视端与手机端的目录、共享代码边界及实施顺序见 [App 架构规划](docs/client-app-architecture.md)。

## 启动

源码构建需要 Go 1.27 或更新版本、Node.js 20.19+（或 22.12+）及 npm。发布后的二进制不需要 Node.js。SQLite 使用纯 Go 驱动，无需 CGO；电影元数据读取和封面提取需要系统安装 `ffprobe` 与 `ffmpeg`。缺少工具时视频仍会以文件名进入列表，扫描状态会记录问题。

```sh
./build.sh
./bin/dinosaur
```

默认监听 `127.0.0.1:8096`，自动创建 `~/.cache/dinosaur/dinosaur.db` 并执行迁移。首次启动生成 `~/.cache/dinosaur/setup-token`，文件权限为 `0600`，日志只记录该文件路径。数据库与生成的数据目录也使用受限权限。账号、会话、媒体库、电影索引、扫描记录、封面和初始化令牌统一保存在该目录。相同用户从不同工作目录启动时使用同一份数据。

```sh
curl http://127.0.0.1:8096/health/live
curl http://127.0.0.1:8096/health/ready
```

`/health/live` 检查进程是否响应，`/health/ready` 检查数据库连接。`SIGINT` / `SIGTERM` 会触发优雅退出；监听端口占用或配置错误时返回非零退出码。

## Web 界面

浏览器打开 <http://127.0.0.1:8096>。首次访问显示管理员初始化页面，填写服务器 `~/.cache/dinosaur/setup-token` 文件中的令牌、用户名和密码；创建后进入登录页。登录成功后直接进入系统设置。

- **媒体库**：添加电影或电视剧媒体库，配置名称和服务器绝对目录路径；支持编辑、移除、类型筛选与名称搜索。目录必须存在且可读取；解析符号链接后，同一目录不能重复添加。添加或编辑媒体库后会自动扫描，扫描中暂时禁止编辑、移除该库。配置保存在 SQLite，服务重启后保留。移除不会修改或删除原始文件。
- **我的电影**：左侧进入电影浏览，或从媒体库卡片点击“查看电影”；支持标题搜索、媒体库筛选、分页、封面和详情。详情包含年份、时长、视频编码、分辨率、文件大小及音轨/字幕信息。支持播放、继续观看、从头播放和已看标记；管理员可以手动重新扫描。
- **账户安全**：修改自己的密码，修改后所有会话失效并返回登录页；sidebar 底部可退出登录。
- **权限**：管理员管理媒体库；普通用户仅能查看名称和类型，不能看到服务器目录或操作配置。

界面使用 React + TypeScript + Vite，播放器使用固定版本的 Artplayer。`build.sh` 先执行锁定依赖安装、类型检查和前端构建，再由 Go 将 `web/dist/` 嵌入二进制；运行时无需 Node.js 或外部 CDN。适配桌面和手机屏幕。电影与电视剧扫描、展示和在线播放已经可用。电视剧按剧集和季度归集，支持单集进度、继续观看和自动播放下一集，详见 [电视剧说明](docs/television.md)。

## 创建管理员并登录

以下示例使用 curl 和 jq。用户名为 3–32 个 ASCII 字母、数字、点、下划线或连字符，首字符须为字母或数字；用户名统一转为小写。密码长度为 12–1024 字节。

```sh
# 首次初始化；管理员创建成功后，该入口永久拒绝再次初始化。
curl -sS http://127.0.0.1:8096/api/setup \
  -H 'Content-Type: application/json' \
  -H "X-Setup-Token: $(cat ~/.cache/dinosaur/setup-token)" \
  --data '{"username":"admin","password":"replace-this-example-password"}'

# 登录并将 Cookie 保存到本地文件；不要提交 Cookie 文件到版本控制。
umask 077
curl -sS -c /tmp/dinosaur-cookies.txt \
  http://127.0.0.1:8096/api/auth/login \
  -H 'Content-Type: application/json' \
  --data '{"username":"admin","password":"replace-this-example-password"}'

# 页面刷新后，可通过 me 接口重新取得 CSRF Token。
curl -sS -b /tmp/dinosaur-cookies.txt http://127.0.0.1:8096/api/auth/me
csrf_token=$(curl -sS -b /tmp/dinosaur-cookies.txt \
  http://127.0.0.1:8096/api/auth/me | jq -r .csrf_token)

curl -sS -b /tmp/dinosaur-cookies.txt \
  http://127.0.0.1:8096/api/admin/users \
  -H 'Content-Type: application/json' \
  -H "X-CSRF-Token: $csrf_token" \
  --data '{"username":"viewer","password":"replace-viewer-example-password","role":"user"}'

curl -sS -X POST -b /tmp/dinosaur-cookies.txt \
  http://127.0.0.1:8096/api/auth/logout \
  -H "X-CSRF-Token: $csrf_token"
```

默认不开放注册，管理员通过接口创建用户。初始化不会自动登录。初始化令牌在已有用户时失效，服务重启不会重新开放初始化。

## API

管理及查询接口返回 JSON，视频、封面和字幕返回相应媒体类型；JSON 错误格式为 `{"error":{"code":"unauthorized","message":"authentication required"}}`，响应带 `X-Request-ID` 和 `Cache-Control: no-store`。带请求体的接口要求 `Content-Type: application/json`，拒绝未知字段和超过 16 KiB 的请求体。

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/health/live` | 进程存活检查，无需登录 |
| GET | `/health/ready` | 数据库就绪检查，无需登录 |
| GET | `/api/setup` | 查询是否完成初始化 |
| POST | `/api/setup` | 首次创建管理员；请求头 `X-Setup-Token`，请求体 `username`、`password` |
| POST | `/api/auth/login` | 用户名密码登录，返回用户、`csrf_token` 和 `expires_at`，设置 Cookie |
| GET | `/api/auth/me` | 查询当前用户及 `csrf_token` |
| POST | `/api/auth/logout` | 撤销当前 Session 并清除 Cookie |
| POST | `/api/auth/password` | 修改自己的密码；请求体 `current_password`、`new_password`，撤销自己的全部 Session |
| POST | `/api/admin/users` | 管理员创建用户；请求体 `username`、`password`、`role`；`role` 默认为 `user`，可选 `admin` |
| PATCH | `/api/admin/users/{id}` | 管理员禁用或启用用户；请求体 `{"disabled":true}`；禁用同时撤销该用户全部 Session |
| GET | `/api/libraries` | 登录用户查看媒体库名称与类型，普通用户不返回服务器目录 |
| GET | `/api/admin/libraries` | 管理员查询媒体库配置及目录可用状态 |
| POST | `/api/admin/libraries` | 添加媒体库；请求体 `name`、`type`（`movie` / `tv`）、`root_path` |
| PUT | `/api/admin/libraries/{id}` | 更新媒体库；请求体同上 |
| DELETE | `/api/admin/libraries/{id}` | 移除媒体库配置，保留原文件 |
| POST | `/api/admin/libraries/{id}/scan` | 启动后台媒体扫描，返回 202 和扫描任务；重复扫描返回 409 |
| GET | `/api/admin/libraries/{id}/scan` | 最近一次扫描的状态、总数、已处理数和问题数 |
| GET | `/api/media` | 电影列表；参数 `q`、`library_id`、`page`（默认 1）、`page_size`（默认 24，最多 100） |
| GET | `/api/media/{id}` | 电影详情，包含音轨、字幕轨与受鉴权保护的封面 URL |
| GET | `/api/media/{id}/poster` | 受登录保护的 JPEG 封面；不可用时返回 404 |
| GET | `/api/playback/{id}` | 播放信息：自动播放方案、视频地址、用户进度、音轨、外挂字幕；可选 `audio`、`mode` |
| GET / HEAD | `/api/playback/{id}/stream` | 受鉴权的视频流；原文件支持 Range；兼容模式支持按 `start` 秒重新取流 |
| PUT | `/api/playback/{id}/progress` | 保存自己的进度；请求体 `position`（秒）或 `watched`（布尔） |
| GET | `/api/playback/continue` | 当前用户最近未看完的影片，最多 12 部 |
| GET | `/api/playback/{id}/subtitles/{subtitle}` | 读取指定的外挂 WebVTT/SRT 字幕，SRT 转换为 WebVTT |
| 任意 | `/api/admin/*` | 管理员权限边界，尚未实现的接口返回 501 |
| 任意 | `/api/media/*`、`/api/playback/*` | 登录权限边界，尚未实现的媒体、字幕与播放接口返回 501 |

除初始化和登录外，所有写操作都需要 Cookie 和 `X-CSRF-Token`。最后一个有效管理员不能被禁用。会话默认固定 7 天过期，不自动续期；过期会话在后续登录时清理。退出、改密和禁用后，原 Cookie 立即无法通过后续请求的认证；重启保留尚未过期的会话。

## 配置

参考 [.env.example](.env.example)。程序直接读取环境变量，不自动读取 `.env` 文件；可通过 shell 或进程管理工具注入配置。

| 变量 | 默认值 | 含义 |
| --- | --- | --- |
| `DINOSAUR_ADDR` | `127.0.0.1:8096` | 监听地址 |
| `DINOSAUR_DATA_DIR` | `~/.cache/dinosaur` | 数据目录 |
| `DINOSAUR_DATABASE_PATH` | `<数据目录>/dinosaur.db` | SQLite 文件路径，不支持内存数据库 |
| `DINOSAUR_FFMPEG_PATH` | `ffmpeg` | 封面提取与兼容播放的可执行文件路径 |
| `DINOSAUR_FFPROBE_PATH` | `ffprobe` | 媒体信息探测的可执行文件路径 |
| `DINOSAUR_SETUP_TOKEN` | 自动生成 | 可选初始化令牌，至少 32 字节 |
| `DINOSAUR_SESSION_TTL` | `168h` | Session 固定有效期 |
| `DINOSAUR_REQUEST_TIMEOUT` | `15s` | 当前 JSON API 请求超时 |
| `DINOSAUR_SHUTDOWN_TIMEOUT` | `10s` | 优雅退出等待时间 |
| `DINOSAUR_COOKIE_SECURE` | `false` | 是否要求浏览器仅通过 HTTPS 发送 Cookie |
| `DINOSAUR_PUBLIC_URL` | 无 | 对外访问的完整 Origin，例如 `https://media.example.com`；不含路径 |

时长配置范围为 1 秒至 365 天。FFmpeg / ffprobe 配置必须非空；扫描时调用对应的可执行文件，失败会计入任务问题数。

HTTPS 部署时同时设置 `DINOSAUR_PUBLIC_URL=https://media.example.com` 和 `DINOSAUR_COOKIE_SECURE=true`。反向代理应保留 Host。未配置 Public URL 时，写请求的 Origin 由 Host 和 Cookie Secure 配置推导。不会信任客户端提交的 `X-Forwarded-For` 或 `X-Forwarded-Proto`；反向代理后所有请求将共享代理 IP 的限流额度，后续可增加受信代理配置。

## 认证实现

- 密码使用 Argon2id，独立随机盐 16 字节、输出 32 字节，参数 `m=65536 KiB,t=3,p=4`，随哈希一起存储。最多同时计算两次密码哈希，控制内存开销。参数参考 [Argon2 RFC 9106](https://www.rfc-editor.org/rfc/rfc9106.html#section-4)。
- Session 使用加密安全随机源生成 32 字节 Token；SQLite 只保存其 SHA-256 哈希，数据库不保存原始 Token。
- Cookie 使用 HttpOnly、SameSite=Lax，并根据部署配置 Secure；认证写操作验证 Session 派生的 CSRF Token。所有写请求拒绝不匹配的 Origin 和跨站 Fetch Metadata；登录和初始化额外要求 JSON，以阻止跨站表单提交。
- 登录失败统一返回相同消息，并对不存在的账号执行密码校验。登录限流为每个真实连接 IP 每 15 分钟 20 次、每个账号每 15 分钟 10 次，成功登录也计入额度。初始化每 IP 每 15 分钟 10 次；更改密码每用户每 15 分钟 5 次；创建用户每管理员每 15 分钟 20 次。限流状态驻留内存，重启清空，满额时返回 429 和 Retry-After。
- SQLite 启用外键、WAL 和 5 秒 busy timeout；当前单进程使用一个连接串行执行数据库操作，迁移和涉及撤销会话的更新均使用事务。配置依据 [SQLite PRAGMA 文档](https://sqlite.org/pragma.html)。
- 电影索引、元数据及受保护的封面接口已实现；播放流独立于 JSON API 超时，按每次写入刷新 30 秒连接写入期限，避免整部电影被普通接口超时截断。

## 开发与检查

```sh
make run    # 构建前端后启动 Go
make build  # React 构建 + Go 编译，输出 bin/dinosaur
make fmt
make check  # 前端类型检查/构建 + 250 行检查 + go vet + go test
make race   # 竞态检测，需要本机支持 CGO
```

测试覆盖并发初始化、SQLite 持久化及约束、认证与角色隔离、CSRF、请求来源、会话过期/撤销、密码修改、禁用用户、限流、请求超时和异常恢复。

目录：`cmd/dinosaur` 为启动入口，`internal/config` 为配置，`internal/store` 为迁移和持久化，`internal/auth` 为密码与令牌工具，`internal/server` 为 HTTP API。`internal/library` 定义媒体库和扫描任务模型，`internal/scanner` 调度后台扫描，`internal/media` 定义电影模型、读取媒体信息并生成封面，`ui/src` 按页面和功能组织 React 组件，`web` 嵌入前端构建产物。`internal/playback` 定义播放方案和观看记录规则，`internal/streaming` 适配 FFmpeg 与外挂字幕。

本次 Web 界面已通过真实 Chromium 验证：初始化、错误密码、登录、目录校验、重复目录、媒体库增改删、筛选搜索、刷新保留、手机布局、修改密码与退出。后端 `go vet` 与竞态测试通过。

## 系统设置与磁盘空间

管理员可在“系统设置”（`/settings/system`）查看每个已挂载本地分区的已用、可用空间和占用条，并手动刷新。接口为 `GET /api/admin/system/disks`；普通用户无法访问。

Linux 原生部署从挂载表及文件系统容量接口读取，不遍历媒体目录。同一分区的重复挂载、bind mount 只显示一次；Btrfs 子卷和多设备文件系统按共享文件系统去重。独立分区分别显示，内存盘、Loop/Snap 镜像、光盘、网络共享等非块设备挂载不计入。当前不展示 ZFS 存储池；非 Linux 系统会提示暂不支持。

“已用”按文件系统实际已分配空间计算，“可用”采用普通用户可写空间；保留块不会计入可用空间。占用条按已用与可用之和计算。部分磁盘读取失败时保留其他磁盘，并显示提示。

## 电影扫描与封面

服务启动后会后台扫描已配置的电影和电视剧媒体库，新增和更新媒体库时也会触发扫描。最多同时处理两个媒体库，同一库只允许一个扫描任务。任务状态持久化；重启会中断旧任务并重新扫描。单个 ffprobe 调用限制为 30 秒，封面提取限制为 20 秒。工具参数参考 [ffprobe 文档](https://www.ffmpeg.org/ffprobe.html)。

扫描递归识别 MKV、MP4、M4V、AVI、MOV、WebM、TS、M2TS、MPG、MPEG，跳过隐藏目录和符号链接。标题与年份先从文件名解析；配置 TMDB 后，扫描完成会通过独立后台任务识别未匹配的影片并获取本地化资料。现有 `/home/grant/media` 的三部电影可用于实际验证。

未变更的视频复用元数据和缓存封面，重复扫描保持电影 ID。完整遍历成功后，将已移除文件标记为缺失并从列表隐藏；根目录不可访问或遍历失败时保留原索引。坏文件仍用文件名展示，详情提示元数据不完整，下次扫描重试。修改媒体库路径或类型会清除旧索引，保留源文件。

封面优先使用同名 `-poster.jpg`、`.jpg`、`.png`；对于子目录中的影片，还会查找 `poster.jpg`、`poster.png`、`folder.jpg`、`cover.jpg`。读取的图片必须在媒体根目录内，限制大小与像素数。没有可用图片时从正片前段提取截图，写入数据目录下的 `posters` 缓存。无法生成时界面显示默认封面。移除媒体不影响原始文件；当前版本暂不自动清理孤立封面缓存。

## 代码约定

[AGENTS.md](AGENTS.md) 规定源码单文件最多 250 行，领域模型与 HTTP、数据库适配层分开；前端按功能拆分。`python3 scripts/check_file_lines.py` 会检查 Go、JavaScript、CSS、HTML、SQL 和 Python 源码，`make check` 已包含该检查。

## 数据位置与迁移

默认目录为 `~/.cache/dinosaur/`，包含 `dinosaur.db`、SQLite 运行文件、`posters/` 和 `setup-token`。原始视频保留在媒体库配置指定的目录。`DINOSAUR_DATA_DIR` 可以覆盖整套数据的位置，`DINOSAUR_DATABASE_PATH` 仍可单独覆盖数据库位置；配置中的 `~` 不由程序展开，请使用绝对路径。

从旧版本的项目 `data/` 目录迁移时，先停止旧服务，确认目标目录不存在，再将整个目录移动到 `~/.cache/dinosaur/`，然后启动新版本。程序不会自动搜寻当前工作目录下的旧数据，也不会合并两套数据库。备份时停止服务并复制整个数据目录。

## 浏览器播放

界面使用固定左侧 sidebar 和右侧 content：导航始终保留，页面内容在右侧独立滚动，窄屏下左栏收为图标导航。

在电影详情点击“播放电影”；已有未看完进度时显示“继续观看”和“从头播放”。“我的电影”页面会展示当前用户的继续观看列表。播放页位于 `/watch/{id}`，占满右侧 content；可以返回列表、使用左侧导航切换页面，或刷新后继续观看。离开播放页会保存进度并回收播放资源。播放器支持播放/暂停、绝对时间拖动、前后 10 秒、音量、静音、全屏、音轨和外挂字幕选择。空格控制播放/暂停，左右方向键跳转；操作输入控件时保留控件自身的键盘行为。

自动选择以下方式：

- MP4/M4V 的 H.264 + AAC，或 WebM 的 VP8/VP9 + Opus/Vorbis：原文件直放，使用 HTTP Range 定位。
- 其他 H.264 视频（例如 MKV）：保留原视频画质，转封装为分片 MP4，并将所选音轨转换为双声道 AAC。
- HEVC 等编码：FFmpeg 软件转码为最高 720p 的 H.264 + 双声道 AAC；也可在播放器中手动切换到兼容模式。

兼容流通过 FFmpeg 分片 MP4 直接输出，不生成整部视频缓存；播放器拖动时使用新起点重新取流。转封装跳转可能落在临近关键帧。延迟开始的音轨会补充前段静音，保持时间轴对齐。实现依据 [FFmpeg 分片 MP4 文档](https://ffmpeg.org/ffmpeg-formats.html)。当前没有 HLS 自适应码率或硬件加速。

最多同时运行两个兼容播放任务，其他请求等待最多 5 秒后返回 503。启动等待上限为 20 秒，单条兼容流最长 6 小时；关闭播放器、跳转、切换音轨或浏览器断开时会回收进程。兼容播放暂停超过 15 秒会释放流，恢复播放时从暂停位置重新取流。需要 FFmpeg 支持 libx264 和 AAC 编码。

播放期间约每 10 秒保存一次进度，暂停、关闭和页面离开时也会尝试保存。进度与已看状态按用户隔离，重启后保留；观看达到影片 90% 时自动标记已看。手动标记未看会重置进度。

外挂字幕支持与视频同名的 `.vtt`、`.srt`，以及 `影片名.en.srt` 这样的语言后缀；仅读取媒体库目录内、UTF-8 编码、不超过 2 MiB 的文件。视频拖动后字幕时间轴会跟随偏移。暂不提取内嵌字幕，也不烧录 PGS 等图形字幕；原有影片中的音轨可以切换。

## Artplayer 与 React 评估

当前已用 MIT 开源的 **Artplayer 5.4.0** 替换手写主播放控件，包含音量、倍速、全屏、画中画及进度交互。播放器仍占满右侧 Content；音轨、外挂字幕、播放方式和倍速以统一图标展示在控制条右侧，与画中画、全屏同行；点击各图标选择具体选项，服务端观看记录继续使用原 API。播放画布覆盖整个右侧 Content，返回使用简洁箭头图标，片名浮在画面上；返回与控制条保持可见，不再保留独立的标题栏、状态栏和底部表单。

Artplayer 的视频代理将原文件或分片流的相对媒体时钟转换为整部影片时间轴。兼容流跳转仍由 Go/FFmpeg 重新取流，拖动时做短暂合并，避免每个鼠标移动都启动一个转码任务。离开页面时销毁 Artplayer 实例、字幕和事件绑定，回收媒体连接。

依赖资源放在 `ui/public/assets/vendor`，保留上游发行包及 MIT 许可证，并记录版本、来源和 SHA-256。CSS 从上游分发内容中原样提取为外部样式，继续使用现有 CSP；不需要 CDN。更新方法见 [vendor 说明](ui/public/assets/vendor/README.md)。

前端已迁移 React + TypeScript + Vite，保留 Go API 和单二进制部署。目录、生命周期与统一构建方式见 [React 前端说明](docs/frontend-react-assessment.md)。

## TMDB 配置

管理员可从左侧“元数据设置”进入 `/settings/metadata`，选择 API Read Access Token（推荐）或 API Key、填写凭据，并保存首选资料语言（默认简体中文）。申请入口：https://www.themoviedb.org/settings/api 。

“测试连接”使用当前输入；凭据留空时使用已保存值，不会保存草稿。“保存配置”允许离线保存，留空保留旧凭据；切换凭据类型时需重新填写。“移除凭据”仅删除凭据，保留语言选项。保存后浏览器清空凭据输入，不提供读取或显示已保存密钥的接口。离开页面或退出登录会清除未保存的输入。

配置保存在现有 `~/.cache/dinosaur/dinosaur.db` 中，凭据以原值保存，以便服务器访问上游；数据库沿用受限文件权限。配置接口仅管理员可用，写操作和连接测试要求 CSRF。上游连接使用固定 HTTPS 地址、最多 8 秒等待且不跟随重定向；上游错误不会原样返回密钥。测试只验证 TMDB API 凭据及连通性，不验证海报 CDN。

| 方法 | 接口 | 作用 |
| --- | --- | --- |
| GET | `/api/admin/metadata/tmdb` | 返回凭据类型、语言、是否已配置，不返回凭据 |
| PUT | `/api/admin/metadata/tmdb` | 保存 `credential_type`、`language`、`credential`；凭据留空保留 |
| DELETE | `/api/admin/metadata/tmdb` | 删除保存的凭据 |
| POST | `/api/admin/metadata/tmdb/test` | 测试当前输入或已保存凭据，不修改配置 |

## 中文电影资料刮削

配置好 TMDB 后，扫描完成会自动启动未识别影片的刮削。管理员也可在“我的电影”点击“刮削未识别影片”（遵循当前媒体库筛选），或在“元数据设置”点击“刮削所有未识别影片”。任务在后台串行执行，显示匹配、待确认、失败数量；离开页面不会取消任务，服务停止会中断并记录状态，重启后未识别影片可重试。

- 刮削包括展示片名、原名、简介、上映年份、别名及海报，默认简体中文；缺少简体资料时尝试其他中文翻译，没有中文译名时保留来源名称并提示，不使用机器翻译。
- 自动匹配用文件名解析的片名搜索并核对年份。必须存在唯一年份候选；相差一年还需要规范化片名一致。缺少年份、多个近似候选或无结果会标记待确认，不直接取搜索第一项。
- 已成功识别的影片保存 TMDB ID 与资料。普通扫描、文件修改和批量刮削不再请求已匹配资料；电影文件名、路径与观看记录不变。
- 在电影详情点击“刷新已匹配资料”会按保存的 TMDB ID 重新读取资料；“匹配 / 更正电影”可以按片名和年份搜索候选，或填写官网电影页面的数字 TMDB ID。年份 0 表示不限。手动选择的新匹配也会长期保留。
- 中文片名、原名、文件名解析名与别名均可搜索。普通用户可查看资料，但不能刮削或更正。
- 海报由服务端下载至 `posters/`，通过原有登录保护的图片接口提供；上游只允许固定 TMDB JPEG 图片路径，最大 5 MiB。海报失败不阻断片名与简介保存。刷新失败保留已匹配的资料。
- 电影和剧集分别使用 TMDB 对应的资料接口，剧集操作见 [电视剧说明](docs/television.md)；不读取 NFO，不支持演员资料与合集；匹配评分与资料来源后续可扩展。后台每部影片之间至少间隔 300 毫秒；凭据无效或上游限流会提前结束任务，其他失败记录后继续，不阻塞本地扫描和播放。

| 方法 | 接口 | 作用 |
| --- | --- | --- |
| POST | `/api/admin/metadata/jobs` | 启动后台刮削，支持 `library_id`、`media_id`；单片支持 `refresh` 或指定 `tmdb_id` |
| GET | `/api/admin/metadata/jobs` | 最近 10 次任务与进度 |
| GET | `/api/admin/media/{id}/metadata/candidates?q=...&year=...` | 搜索更正候选，省略参数时从文件名取片名和年份 |

每次最多运行一个刮削任务。重复触发返回 409；扫描完成事件会合并成一次后续任务，继续处理扫描中新出现的未识别影片。所有写接口要求管理员 Session 与 CSRF。数据库新增资料及任务表，与技术信息分开；技术扫描只更新原有文件信息。完整选型记录见 [中文元数据刮削评估](docs/metadata-scraper-assessment.md)。

## 刮削管理后台

管理员可从左侧“刮削管理”进入 `/settings/scraping`。影片状态页按媒体库、识别状态及片名/别名/文件名搜索，展示 TMDB ID、资料语言、更新时间和提示原因；可查看详情、匹配更正、重试未识别影片或主动刷新已匹配资料。“有提示”包含在已识别数量中，用来定位海报未下载、暂无中文译名或刷新失败等情况。

任务历史支持分页和状态筛选，展示任务范围、时间、进度与匹配/待确认/失败/跳过数量。点击“查看结果”查看任务当时的逐片状态和原因。新任务开始时记录待处理项，每处理一部更新快照；未执行的项在提前失败或中断后明确标记。重新匹配不改写旧任务结果；源文件被移除后，历史记录仍保留，但不再提供该影片的操作入口。

升级前的任务只有汇总记录，不补造逐片结果；管理页会提示在“影片状态”查看当前资料。管理页在运行中的任务发生进度变化时自动更新，也可点击“刷新状态”。普通用户无法访问管理接口或侧栏入口。

| 方法 | 接口 | 作用 |
| --- | --- | --- |
| GET | `/api/admin/metadata/records` | 当前影片识别报告，支持 `library_id`、`q`、`status`、`page`、`page_size` |
| GET | `/api/admin/metadata/jobs` | 任务历史，支持 `status`、`page`、`page_size`，默认最近 10 条 |
| GET | `/api/admin/metadata/jobs/{id}` | 任务汇总与逐片快照，支持 `page`、`page_size` |

影片状态支持 `matched`、`unmatched`、`needs_review`、`failed`、`warning`；任务状态支持 `queued`、`running`、`completed`、`failed`、`interrupted`。省略状态表示全部。每页最多 100 条。管理报告只查询本地数据库，不会为查看状态发起 TMDB 请求。

## 电影详情页

电影详情位于 `/movies/{id}`，作为独立页面占满右侧 Content，保留左侧导航。桌面按海报与影片资料两栏排列，手机使用单栏。片名下方直接提供播放、继续观看和已看标记，简介、匹配资料及媒体信息在同一页面查看。

详情支持刷新、浏览器前进后退及加载失败重试。返回箭头回到进入详情前的电影列表或刮削管理；从详情开始播放后，浏览器返回会恢复详情并读取刚保存的观看进度。资料刷新或已看标记更新不会重复添加详情历史记录。

## NAS 原生部署

提供 Linux/amd64 原生发布包与 systemd 用户服务，无需容器。程序、配套 FFmpeg/ffprobe 使用独立版本目录；运行数据仍统一保存在 `~/.cache/dinosaur/`。默认 NAS 端口为 18096，避免与现有 8096 服务冲突。打包、安装、开机启动、更新和备份方法见 [原生部署说明](docs/native-deployment.md)。

## 电视剧

从 sidebar 的“我的剧集”进入剧集列表；媒体库类型选择“电视剧”即可自动扫描。支持 `S01E01` / `1x01` 命名、分季浏览、按集观看记录、跨季下一集、TMDB 整剧中文资料及人工更正。命名规则、API 和当前边界见 [电视剧使用说明](docs/television.md)。
