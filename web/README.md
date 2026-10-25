# Web

所有 HTML、CSS、JavaScript 由 `embed.go` 嵌入 Go 服务，无需构建工具。

`app.js` 处理登录与页面切换，`libraries.js` 管理媒体库，`scans.js` 展示扫描任务，`movies.js` 浏览电影，`movie-detail.js` 展示详情；`player.js` 管理播放生命周期，配套文件分别处理播放控件、观看进度和字幕。样式按页面和组件拆分；全部源码遵守单文件 250 行上限。

使用同源 Session Cookie，CSRF Token 仅保留在页面内存，封面同样需要登录。媒体名称和文件名渲染前进行 HTML 转义。支持桌面与手机。

修改资源后重启 `go run ./cmd/dinosaur` 即可查看。

TMDB 刮削通过独立后台任务完成：列表和元数据设置可启动未识别影片的刮削，详情展示中文资料并提供候选搜索、TMDB ID 更正和主动刷新。已匹配资料在扫描时保留，片名、原名和别名支持搜索，下载海报从登录保护的本地接口提供。
