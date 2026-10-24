# React 迁移评估

## 结论

建议在继续扩展剧集、元数据匹配、用户管理之前，逐步迁移到 **React + TypeScript + Vite**。本次先完成 Artplayer 接入，不同时重写全站。Go 继续负责 API、Session、扫描、转码和 SQLite；部署仍然是一个嵌入前端构建产物的 Go 二进制。

这个建议基于当前项目的页面与状态复杂度，不是播放解码能力的需求：React 不会改善 HEVC 解码或替代 FFmpeg。

## 当前实现与实际问题

- 页面是 `html/template` 输出的原生 HTML，CSS 与多个经典 JavaScript 脚本通过 `go:embed` 分发。没有 JSX、React、npm 构建或 Node 服务。
- 功能已经包括认证、设置、电影列表、详情、扫描轮询、播放、字幕、观看记录及浏览器历史导航。
- `$()`、`state`、`player` 等全局对象和脚本加载顺序存在隐式依赖；异步请求需要手动检查用户、页面和请求版本。
- 电影卡片、详情与媒体库列表依靠字符串拼接和 `innerHTML` 更新，表单、错误、加载状态和事件解绑需要人工保持一致。
- 播放器离开页面时必须取消计时器、请求、字幕和 FFmpeg 连接；这些生命周期已经值得用组件边界统一管理。

## 收益与代价

| 项目 | 保持原生实现 | React + TypeScript |
| --- | --- | --- |
| 常规开发依赖 | Go 即可 | 增加 Node、包管理与前端构建 |
| 页面和状态复用 | 依赖约定与手动 DOM 操作 | 组件、Hook、类型化 API |
| 播放器集成 | 当前 Artplayer 适配层 | 容器 ref + Effect，清理时 destroy |
| 路由与登录状态 | 自写 history 与状态切换 | 路由组件、认证上下文、明确的页面挂载边界 |
| 部署方式 | Go 嵌入原始资源 | Go 嵌入编译后的静态资源，无需 Node 运行时 |
| 本次成本 | Artplayer 直接替换主控件 | 还要迁移所有页面及其回归用例 |

## 建议的结构和技术选择

- 使用 Vite 的 React + TypeScript 模板，前端源码放 `ui/src`，构建产物放 `web/dist`；Go 的嵌入入口只分发产物及 SPA HTML。
- 使用 React Router 管理 `/login`、`/setup`、`/movies`、`/settings`、`/watch/:id`。保留固定 Sidebar + Content，避免改变已有页面地址。
- 按 `auth`、`libraries`、`movies`、`playback` 组织 feature；认证和当前用户放 Context，其他页面状态放本地 Hook。初期不引入 Redux 等额外全局状态库。
- API 层统一处理 Cookie、CSRF、错误和 AbortController；类型与现有 JSON DTO 对齐。所有异步页面请求在卸载时取消，并拒绝使用旧用户的响应。
- `PlayerPage` 负责加载播放信息和观看状态；Artplayer 容器由 `useRef` 持有，通过 Effect 初始化和销毁。视频时间轴适配器与 React 无关，组件只传入时长、位置、跳转和资源恢复回调。
- Cookie 保持 HttpOnly，同源请求继续携带 Session；不把认证令牌或观看进度迁移到 localStorage。
- 保留当前 CSS 视觉风格；迁移时按页面与组件整理，初期不同时引入 UI 框架或 CSS-in-JS。
- 不使用 Next.js 或 SSR：当前是登录后的私人应用，已有 Go 后端，没有需要新增服务端 JavaScript 的功能。

## 实施顺序与验收

1. 建立 Vite/TypeScript 构建和 Go 嵌入流程，固定依赖版本；将行数检查扩展到前端源码，生成产物与第三方发布文件不作为手写源码维护。
2. 迁移 Sidebar、Content、路由、初始化和登录，验证 CSRF、退出、会话过期、刷新与浏览器前进后退。
3. 迁移媒体库设置、电影列表、详情和扫描状态，保持 API 不变，验证搜索、分页和旧数据读取。
4. 迁移播放器、字幕及观看记录，复用已经验证的时间轴和 FFmpeg 适配逻辑。验证直放、MKV 转封装、HEVC 转码、音轨切换、断点续播与离开页面回收。
5. 通过后删除旧的全局脚本和模板中的交互 DOM；保证所有手写源文件仍不超过 250 行。

迁移期间同一个 DOM 子树只能由 React 或旧脚本中的一方管理。可以按路由逐页切换；正式发布前完成统一入口，不长期保留两套认证或播放器状态系统。

## 构建与运行约定

开发环境新增符合 Vite 官方要求的 Node.js 和 npm；需要版本锁文件与可重复的 `npm ci`。CI 先构建前端，再构建 Go。开发模式由 Vite 将 `/api` 请求代理到 Go；生产环境统一由 Go 分发，Session 与媒体请求同源。

若要求克隆后只用 Go 就能构建，则需提交构建产物或在发布中提供已经打包的二进制；推荐日常源码构建明确执行前端构建，用户直接使用发布二进制。引入 React 不改变数据库或 `~/.cache/dinosaur/` 的数据位置。

## 参考

- [React：向现有项目添加 React](https://react.dev/learn/add-react-to-an-existing-project)
- [React：与外部系统同步](https://react.dev/learn/synchronizing-with-effects)
- [Vite：开始使用](https://vite.dev/guide/)
- [Vite：后端集成](https://vite.dev/guide/backend-integration)
- [Artplayer：实例销毁和播放属性](https://github.artplayer.org/document/en/advanced/property.html)
