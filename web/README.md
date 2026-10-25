# 嵌入的前端构建产物

React 源码在 `ui/src/`，第三方固定版本资源在 `ui/public/assets/vendor/`。

`./build.sh` 自动安装锁定依赖、构建 `web/dist/`，再编译 Go。`embed.go` 只嵌入该生产目录；生成文件不提交版本控制。部署运行时不需要 Node.js。

开发与目录说明见 [ui/README.md](../ui/README.md)。
