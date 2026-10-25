# NAS 原生部署

本方案直接运行 Linux/amd64 可执行文件，由 systemd 用户服务管理。无需 Docker、NAS 上的 Go 或 Node.js，也不替换系统 FFmpeg。

## 目标环境与目录

当前 NAS 是 Ubuntu 20.04、x86_64、glibc 2.31、Linux 5.15，用户为 `grant`。用户 systemd 已可用且 Linger 已开启。8096 端口已占用，因此默认监听 `0.0.0.0:18096`。

- 程序与工具：`~/.local/share/dinosaur/releases/<版本>/`。
- 当前版本：`~/.local/share/dinosaur/current` 符号链接。
- 服务配置：`~/.local/share/dinosaur/dinosaur.env`。
- 数据库、海报与初始化令牌：`~/.cache/dinosaur/`。
- 用户服务：`~/.config/systemd/user/dinosaur.service`。

新部署不会复制开发机数据库、账号、TMDB 凭据或观看记录。首次访问需创建管理员，再配置媒体库和 TMDB。NAS 当前电影目录可使用 `/mnt/inner/ptfile/dianying`；程序读取原文件，索引和海报写入自己的数据目录。

## 构建发布包

开发机需要 Go、Node.js 和 npm；打包脚本自动调用 `build.sh`，先构建 React 再交叉编译 Go。NAS 运行时仍无需 Node.js。

在开发机准备 Linux x86_64 的 FFmpeg GPL 非 shared 发行包，需同时包含 `bin/ffmpeg`、`bin/ffprobe`、libx264 和 AAC。使用发行方公布的 SHA-256 校验值，不能仅对未知来源下载文件自行计算并当作来源校验。

```bash
python3 scripts/package_native.py \
  --ffmpeg-archive /absolute/path/to/ffmpeg-linux64-gpl.tar.xz \
  --ffmpeg-sha256 '<发行方提供的 SHA-256>' \
  --ffmpeg-source '<发行包的 HTTPS 下载地址>'
```

产物在 `bin/native/`，包含 Go 程序、FFmpeg、ffprobe、许可证、安装器、服务文件和每个文件的 SHA-256 清单；归档旁还生成 `.sha256` 文件。发布包不包含应用数据库或凭据。Go 使用 `CGO_ENABLED=0 GOOS=linux GOARCH=amd64` 构建，不依赖 NAS 的 glibc 来运行 Go 程序。

可从 [FFmpeg 下载页](https://ffmpeg.org/download.html)链接的 [BtbN 发行项目](https://github.com/BtbN/FFmpeg-Builds/releases)获取工具。BtbN Linux 构建要求 glibc 2.28、Linux 4.18 或更高，适用于当前 NAS。系统自带 FFmpeg 4.2.7 缺少当前兼容流所需的 `readrate` 和 `readrate_initial_burst` 参数，不能直接用于完整播放功能。

## 安装与启动

上传产物与对应 `.sha256` 到 NAS 的临时目录，校验并解压，然后以普通用户运行包中的安装器：

```bash
sha256sum -c dinosaur-<版本>-linux-amd64.tar.gz.sha256
tar -xzf dinosaur-<版本>-linux-amd64.tar.gz
python3 dinosaur-<版本>/install.py
```

安装器先校验包文件、架构和工具能力，再写入新版本目录，更新 `current` 链接，安装并启用 `dinosaur.service`，最后验证健康接口。已有配置和数据会保留；新部署健康检查失败会停用新服务，更新失败则尝试切回原程序版本。数据库迁移不会自动回退，更新前仍应备份数据。

若其他 Linux 主机尚未开启用户服务驻留，先执行一次 `sudo loginctl enable-linger <用户名>`。当前 NAS 已开启，不需重复设置。

部署后访问 `http://192.168.31.197:18096`。初始化令牌在 NAS 的 `~/.cache/dinosaur/setup-token`，请在自己的终端读取并填入初始化页面。不要把令牌提交到仓库。TMDB 凭据在首次登录后从“元数据设置”单独配置。

## 常用操作

```bash
systemctl --user status dinosaur.service
journalctl --user -u dinosaur.service -f
systemctl --user restart dinosaur.service
systemctl --user stop dinosaur.service
systemctl --user disable --now dinosaur.service
```

修改 `dinosaur.env` 后重启生效。HTTPS 反向代理需要同时设置 `DINOSAUR_PUBLIC_URL=https://你的域名` 和 `DINOSAUR_COOKIE_SECURE=true`，并让代理保留 Host。局域网直接通过 HTTP 访问时保持默认值。

更新时上传新的发布包并再次运行安装器即可。备份应先停止服务，再复制整个 `~/.cache/dinosaur/` 和 `dinosaur.env`。回退程序版本可以把 `current` 切回旧目录后重启，但需要确认数据库兼容性；恢复旧数据库应使用对应版本的完整备份。
