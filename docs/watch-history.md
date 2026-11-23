# 观看记录

观看记录独立于媒体文件与播放进度，按账号保存。影片在站内播放到已看阈值或手动标记已看时，自动记录当天；连续进度上报不会重复写入。同一影片同一天去重，不同日期可保留重看记录。取消已看不删除已有日记。升级时将已有已看进度按其更新时间补入时间轴；日期统一使用北京时间。

豆瓣导入保留片名、原名、年份、个人评分（0 表示未评分）、观看日期、短评、豆瓣条目 ID 及 HAR 内的海报。不会根据相似片名自动绑定本地片源，也不改变播放进度。海报存入数据目录的 `history-posters/`，展示时不请求豆瓣；缺少海报时保留文字记录。

## API

| 方法 | 路径 | 用途 |
| --- | --- | --- |
| GET | `/api/history?page=1&page_size=24` | 当前账号的时间轴，日期倒序；每页 1–100 条 |
| POST | `/api/history` | 为当前账号添加手动记录 |
| PUT | `/api/history/{id}` | 修改自己的记录，不改变导入来源或绑定关系 |
| DELETE | `/api/history/{id}` | 删除自己的记录，不修改播放进度 |
| GET | `/api/history/{id}/poster` | 读取自己的记录海报 |

新增与修改使用 JSON：`title`、`original_title`、`year`、`watched_on`（`YYYY-MM-DD`）、`rating`（0–5）、`review`。必需字段为片名与有效日期。Browser 会话写操作需要 CSRF；手机使用独立 Bearer 会话。不接受请求中的用户 ID，管理员也不能通过这些接口读取其他账号的记录。记录海报同样鉴权。

## HAR 导入

```bash
go build -o bin/dinosaur-history ./cmd/dinosaur-history
bin/dinosaur-history --har /path/to/movie.douban.com.har --check
```

检查模式不连接数据库、不保存海报，只报告记录数、合集总数、分页偏移与捕获的海报数。只读取 HAR 中已保存的豆瓣已看页面和海报，不联网补抓。一个 HAR 中包含不同豆瓣账号时拒绝导入。

正式导入前停止服务并备份数据库，使用明确的目标账号及现有数据库：

```bash
systemctl --user stop dinosaur.service
bin/dinosaur-history --har /path/to/movie.douban.com.har \
  --database ~/.cache/dinosaur/dinosaur.db \
  --data-dir ~/.cache/dinosaur --user admin
systemctl --user start dinosaur.service
```

海报按内容哈希保存，记录在一个数据库事务中导入。同一账号、来源、条目及观看日期重复时跳过，不覆盖已经编辑的资料。重复导入不会增加同一批记录。HAR 未捕获的页面不会被补造；本次样本合集显示 110 部，仅捕获偏移 15–105 的 95 部。

手机 0.6.0 的第三个 tab 为“记录”，按月展示时间轴，支持继续加载、下拉刷新、记录详情和豆瓣来源链接；片名搜索仍可从首页搜索栏进入。
