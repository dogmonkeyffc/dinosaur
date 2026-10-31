# Dinosaur App 初版

手机和电视是两个独立的 Expo 工程，共享 TypeScript API、凭据存储、基础组件与原生播放适配。现有 Web 工程仍在 `ui/`。

当前版本交付 Android 手机和 Android TV / Google TV 两个客户端。三星 Tizen、LG webOS 不能安装 Android TV APK。

## 登录流程

1. 在已登录的 Web 中进入“账户安全”，检查登录账号并生成二维码。
2. 手机 App 点击“扫描 Web 登录二维码”，扫码后确认服务器地址及设备名称。二维码两分钟有效，仅能兑换一次；刷新二维码、退出对应 Web 会话或修改密码会使旧二维码失效。
3. 手机可以保存多个服务器/账号，在“切换账号”中选择本次要用于电视登录的账号。
4. 电视 App 启动后通过 `_dinosaur._tcp.local.` 自动查找局域网服务器，选中服务器后展示六位数字配对码。也可手动填写服务器地址。
5. 手机选择“登录电视”，输入六位码，确认目标电视名称和当前账号后授权。电视轮询取得独立设备会话并进入媒体列表。配对码五分钟有效，不能单独用于领取登录 Token。

电视刷新配对码或退出配对页时会撤销旧请求。手机和电视离线退出会先清除本机账号；未能联系服务器时，原设备会话仍可从 Web 撤销或等待到期。

手机和电视均支持电影/剧集浏览、搜索、选集、原片播放、兼容模式、观看进度、下一集、原生音轨与字幕选择。内嵌文字字幕由服务端分段转换，PGS 等图片字幕通过兼容转码显示；ASS 字体与特效不保留。进入影片时自动选择中文，其次英语；手动切换或关闭后保留当前影片的选择，下一集重新自动选择。当前播放器为 Expo Video，在 Android 上使用原生 ExoPlayer。

Web“已登录设备”可单独退出手机或电视；改密和禁用账号撤销其全部会话。电视授权使用手机当前账号，客户端不能指定任意用户 ID 或提升角色。

## 网络

服务端新增 `DINOSAUR_SERVER_NAME`（默认 Dinosaur）和 `DINOSAUR_DISCOVERY_ENABLED`（默认 true）。监听回环地址时不发布发现信息；局域网部署应绑定 LAN 地址或 `0.0.0.0`。发现信息在启用组播的私有 IPv4 网卡上发布，需要同一局域网允许 UDP 5353；跨 VLAN、访客 Wi-Fi 和部分模拟器不能传递组播，可用手动连接。

发现信息优先提供 `DINOSAUR_PUBLIC_URL`，连接失败时尝试发布的局域网 IP。手机、电视最终连接的是同一个服务器 ID，配对码仅在创建它的服务器上有效。

HTTPS 必须使用设备信任的证书。局域网测试可直接使用 `http://192.168.x.x:端口`；App 会显示该连接使用 HTTP。HTTP 地址只接受私有 IPv4、回环地址及 `.local` 主机，其他主机必须使用 HTTPS。自定义域名在 Android 上还需要对应 DNS 解析，手机浏览器信任的私有证书不一定被原生 App 信任。

## 开发

原生工程使用 Expo SDK 56 与对应 `react-native-tvos` 0.85 系列。依据 Expo TV 的单仓库要求，手机也使用同一 TV fork，但仅电视工程开启 TV 构建配置。Web 的 React 依赖仍独立。

```sh
npm ci
npm run typecheck
npm run test:client
npm run mobile   # 手机开发服务
npm run tv       # 电视开发服务
```

使用 development build 或实际原生构建；Expo Go 不包含电视发现模块。首次生成原生项目必须使用匹配的模板：

```sh
cd apps/mobile  # 或 apps/tv
npx expo prebuild --platform android --no-install --template expo-template-bare-minimum@56.0.38
npx expo run:android
```

## Android 安装包

需要 Node.js 22.13+ 或 24.3+、JDK 17、Python 3 和 Android SDK。SDK 需包含 Platform 36、Build Tools 36、NDK 27.1.12297006 以及 CMake 3.22.1；配置 `ANDROID_HOME`，并接受 Android SDK 许可。

```sh
scripts/build_apps.sh all
scripts/build_apps.sh mobile
scripts/build_apps.sh tv
```

产物为 `bin/apps/dinosaur-mobile-preview.apk`、`bin/apps/dinosaur-tv-preview.apk` 及 SHA-256 文件。默认包含 arm64-v8a 和 x86_64，可通过 `DINOSAUR_ANDROID_ABIS` 修改。初版沿用模板调试签名作内部安装，正式发布需要独立签名与版本管理；此 APK 不含 Metro 开发服务器依赖。

Go 服务仍通过根目录 `build.sh` 打包，不需要 Android SDK。App 要求服务端支持 `/api/server`、设备会话及配对 API；旧版 NAS 服务需升级后才能使用这些登录功能。

`android/` 由已提交的 Expo 配置与插件生成，构建缓存、APK 和本机凭据不提交到 Git。

新增内嵌字幕功能需要同时升级服务端并重新构建手机/电视 APK；旧的 preview 安装包未包含图片字幕切换与分段续载适配。

## 初版验证范围

已在 Android 模拟器验证 Web 二维码载荷登录、手机多账号切换、手机授权电视登录、受保护的封面与视频播放、观看进度回传；已通过局域网查询验证服务端 mDNS 广播。自动化检查覆盖设备会话隔离、一次性兑换、并发兑换、配对过期与撤销、CSRF 及账号权限。

手机摄像头实际扫码、真实 Android TV 的遥控器体验及同一 Wi-Fi 下的自动发现仍需真机验收。模拟器网络不能传递组播，模拟器电视登录使用了手动服务器地址。
