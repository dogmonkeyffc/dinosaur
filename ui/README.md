# React 前端

React + TypeScript + Vite，页面直接调用同源 Go API。生产运行只有一个 Go 进程。

视觉规范见 [DESIGN.md](./DESIGN.md)，全局主题入口为 `src/styles/sketch.css`。

## 统一构建

```sh
./build.sh
./bin/dinosaur
```

构建需要 Go 1.27+、Node.js 20.19+（或 22.12+）和 npm。脚本每次执行 `npm ci --include=dev`，再进行 TypeScript 类型检查与 Vite 构建，输出 `web/dist/`，最后编译 Go。依赖固定在 `package.json` 和 `package-lock.json`；不提交 `node_modules` 或 `web/dist` 的生成内容。

```sh
./build.sh --frontend-only
./build.sh --output /absolute/path/dinosaur
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 ./build.sh
```

`web/dist/.gitkeep` 只用于保留 Go 嵌入目录；没有构建产物时服务会明确提示先运行构建。直接执行 `go build` 不会自动运行前端构建。`make build`、`make run`、`make check` 和 `scripts/package_native.py` 已连接统一流程。NAS 原生包打包时自动构建前端，不需要另一步手工复制。

## 开发

先运行一次 `./build.sh --frontend-only`，随后在两个终端分别执行：

```sh
# 终端一
 go run ./cmd/dinosaur
# 终端二
 npm --prefix ui run dev
```

访问 Vite 输出的本地地址（默认 5173）。`/api` 和 `/health` 代理到 `http://127.0.0.1:8096`，页面热更新。使用浏览器同源 Cookie 与 CSRF，开发环境不要设置生产地址的 `DINOSAUR_PUBLIC_URL`。

## 组织方式

- `src/auth`：初始化、登录、会话与过期处理。
- `src/layout`：桌面侧栏、手机顶部菜单和内容布局。
- `src/libraries`：媒体库与扫描操作。
- `src/catalog`：电影、剧集列表、详情与单集。
- `src/metadata`、`src/scraping`：TMDB 配置、候选匹配和后台任务。
- `src/playback`：Artplayer 实例、源时间轴、字幕和控件；组件卸载时销毁实例、取消请求并保存观看进度。
- `src/settings`：账户和磁盘显示。
- `src/lib`：类型、API、可取消请求、格式化与串行进度写入。
- `src/styles`：页面和组件样式；`public/assets/vendor` 保留固定版本第三方资源及许可证。

每个手写源码文件不超过 250 行，`python3 scripts/check_file_lines.py` 同时检查前端与 Go。生成产物和第三方发行文件不按手写源码计数。`npm --prefix ui run format` 格式化前端源码。

Go 只分发生产构建：HTML 使用 `no-store`，带内容哈希的 JS/CSS/SVG 长期缓存，固定路径第三方资源需重新验证。刷新已有页面地址仍返回 SPA 入口，未知 API 和资源保持 404。
