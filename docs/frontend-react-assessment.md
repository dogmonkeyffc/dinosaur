# React 前端与构建

前端已从全局 HTML/JavaScript 脚本迁移到 React + TypeScript + Vite。Go 保留认证、媒体库、扫描、TMDB、播放和持久化 API，生产环境依然只有一个嵌入静态前端的 Go 二进制。

## 组件与生命周期

认证、布局、媒体库、电影/剧集列表与详情、刮削、设置和播放器均由 React 组件渲染。路由保留 `/settings`、`/movies`、`/series`、`/watch/:id` 等已有地址；账户安全拥有独立的 `/settings/account`。

认证状态由 Context 管理，CSRF 只保存在内存。切换用户会取消旧会话请求，页面卸载和查询变化会取消旧的数据读取。后台扫描与刮削在挂载期间轮询；离开页面后清理定时器。

Artplayer 继续使用已验证的固定版本资源。React 只管理播放器容器与实例生命周期，Artplayer 管理其内部节点；组件卸载时销毁实例。原生媒体时间与源时间轴的适配、兼容流跳转、音轨/外挂字幕、暂停释放与继续播放、下一集和进度保存保留。进度写入串行执行，读取详情前等待待提交进度。

React 仍然按组件与职责拆分文件；其价值是减少全局变量和跨页面 DOM 操作，使布局及交互有明确归属，而非保证每次修改的文件数量更少。

## 构建

```sh
./build.sh
./bin/dinosaur
```

脚本执行锁定依赖安装、类型检查、Vite 生产构建、250 行源码检查与 Go 编译。前端产物在 `web/dist/`，由 Go `embed` 嵌入。`make build` 和 NAS 原生包共用这条流程。构建机需要 Node.js/npm，NAS 只运行二进制及配套媒体工具。

完整开发、交叉编译与目录说明见 [ui/README.md](../ui/README.md)。

## 参考

- [React：从零构建应用](https://react.dev/learn/build-a-react-app-from-scratch)
- [React：与外部系统同步](https://react.dev/learn/synchronizing-with-effects)
- [Vite：后端集成](https://vite.dev/guide/backend-integration.html)
- [Artplayer：实例属性和销毁](https://github.artplayer.org/document/en/advanced/property.html)
