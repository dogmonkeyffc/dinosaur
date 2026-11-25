package com.monkeydog.dinosaur.mobile

import com.monkeydog.dinosaur.android.json
import com.monkeydog.dinosaur.android.rows
import org.json.JSONArray
import org.json.JSONObject

internal class Admin(private val activity: MainActivity) {
    private val client get() = activity.api

    fun show() = with(activity) {
        page("管理", { catalog.show() }, { catalog.show() })
        val api = client
        task({ api.request("/api/auth/me").getJSONObject("user") }) { user ->
            if (user.getString("role") != "admin") {
                label("当前账号没有管理权限")
            } else {
                libraries()
            }
        }
    }

    private fun libraries(): Unit = with(activity) {
        adminPage({ catalog.show() }, ::destinations)
        content.addView(phoneRow().apply {
            addView(phoneText("媒体库", 32, ADMIN_TEXT), android.widget.LinearLayout.LayoutParams(0, -2, 1f))
            addView(adminAction("＋ 添加媒体库", { editLibrary(null) }, true))
        })
        val api = client
        task({ api.request("/api/admin/libraries").rows("libraries") }) { rows ->
            adminLibraryList(rows, ::scanName, ::library, { catalog.browseLibrary(it.getLong("id")) }, ::scan)
        }
    }

    private fun destinations() = with(activity) {
        val actions = listOf("媒体库设置" to { libraries() }, "刮削管理" to { scraping() },
            "元数据设置" to { metadata() }, "用户管理" to { users() }, "系统设置" to { system() })
        choose("管理导航", actions.map { it.first }) { actions[it].second() }
    }

    private fun scan(item: JSONObject) = with(activity) {
        val api = client
        task({ api.request("/api/admin/libraries/${item.getLong("id")}/scan", "POST", JSONObject()) }) {
            message("已开始扫描")
            libraries()
        }
    }

    private fun scanName(status: String) = when (status) {
        "queued" -> "等待扫描"
        "scanning" -> "扫描中"
        "completed" -> "扫描完成"
        "failed" -> "扫描失败"
        "interrupted" -> "扫描中断"
        else -> ""
    }

    private fun library(item: JSONObject): Unit = with(activity) {
        val id = item.getLong("id")
        val path = "/api/admin/libraries/$id"
        page(item.getString("name"), { libraries() })
        label(item.optString("root_path"))
        item.optJSONObject("scan")?.let { label(scanName(it.optString("status"))) }
        button("编辑") { editLibrary(item) }
        button("访问权限") { access(id) { library(item) } }
        val api = client
        button("重新扫描") { scan(item) }
        button("刷新扫描状态") {
            task({ api.request("/api/admin/libraries").rows("libraries").first { it.getLong("id") == id } }) { library(it) }
        }
        button("移除媒体库") {
            confirm("移除媒体库 ${item.getString("name")}？") {
                task({ api.request(path, "DELETE") }) { libraries() }
            }
        }
    }

    private fun editLibrary(item: JSONObject?): Unit = with(activity) {
        val names = listOf("电影", "电视剧", "普通视频")
        val types = listOf("movie", "tv", "video")
        choose("媒体类型", names, types.indexOf(item?.optString("type"))) { index ->
            fun edit(root: String) = form("媒体库", listOf("名称", "服务器媒体目录"), listOf(item?.optString("name").orEmpty(), root)) { values ->
                val api = client
                val path = "/api/admin/libraries" + (item?.let { "/${it.getLong("id")}" } ?: "")
                val body = json("name" to values[0], "root_path" to values[1], "type" to types[index])
                task({ api.request(path, if (item == null) "POST" else "PUT", body) }) { libraries() }
            }
            val root = item?.optString("root_path").orEmpty()
            choose("服务器媒体目录", listOf("输入路径", "浏览目录")) { choice ->
                if (choice == 0) edit(root)
                else DirectoryPicker(activity).show(root, previous = { libraries() }) { path -> edit(path) }
            }
        }
    }

    private fun access(id: Long, previous: () -> Unit): Unit = with(activity) {
        page("访问权限", previous)
        val api = client
        val path = "/api/admin/libraries/$id/access"
        task({ api.request(path) }) { data ->
            val current = data.getJSONObject("access")
            var restricted = current.getBoolean("restricted")
            val ids = current.optJSONArray("user_ids") ?: JSONArray()
            val selected = (0 until ids.length()).map(ids::getLong).toMutableSet()
            rowSwitch("限制访问", restricted) { value, _ -> restricted = value }
            data.rows("accounts").forEach { item ->
                val userId = item.getLong("id")
                rowSwitch(item.getString("username"), userId in selected) { enabled, _ ->
                    if (enabled) selected.add(userId) else selected.remove(userId)
                }
            }
            button("保存") {
                val body = json("restricted" to restricted, "user_ids" to JSONArray(selected.toList()))
                task({ api.request(path, "PUT", body) }) { previous() }
            }
        }
    }

    private fun users(): Unit = with(activity) {
        page("用户管理", { libraries() }, { catalog.show() })
        val api = client
        button("添加账号") {
            form("添加账号", listOf("用户名", "密码")) { input ->
                task({ api.request("/api/admin/users", "POST", json("username" to input[0], "password" to input[1], "role" to "user")) }) {
                    users()
                }
            }
        }
        task({ api.request("/api/admin/users").rows("users") }) { items ->
            items.forEach { user ->
                val id = user.getLong("id")
                val name = user.getString("username")
                val disabled = user.optBoolean("disabled")
                label(name + if (user.optString("role") == "admin") " · 管理员" else "")
                if (id != account?.user?.optLong("id")) {
                    val action = if (disabled) "启用" else "停用"
                    button("$action $name") {
                        confirm("$action $name？") {
                            task({ api.request("/api/admin/users/$id", "PATCH", json("disabled" to !disabled)) }) { users() }
                        }
                    }
                }
            }
        }
    }

    private fun scraping(tv: Boolean = false): Unit = with(activity) {
        page(if (tv) "剧集刮削" else "电影刮削", { libraries() }, { catalog.show() })
        button(if (tv) "电影刮削" else "剧集刮削") { scraping(!tv) }
        val path = if (tv) "/api/admin/tv/metadata/jobs" else "/api/admin/metadata/jobs"
        val api = client
        button("开始刮削") {
            task({ api.request(path, "POST", JSONObject()) }) { scraping(tv) }
        }
        button("刷新") { scraping(tv) }
        button("刮削结果") { ScrapeResults(activity).show({ scraping(tv) }) }
        task({ api.request(path).rows("jobs") }) { rows ->
            if (rows.isEmpty()) label("暂无任务")
            rows.forEach { job ->
                button("#${job.getLong("id")} · ${job.optString("status")} · ${job.optInt("processed")}/${job.optInt("total")}") {
                    ScrapeResults(activity).job(job.getLong("id")) { scraping(tv) }
                }
            }
        }
    }

    private fun metadata(): Unit = with(activity) {
        page("元数据设置", { libraries() }, { catalog.show() })
        val path = "/api/admin/metadata/tmdb"
        val api = client
        task({ api.request(path) }) { data ->
            label(if (data.optBoolean("configured")) "TMDB 已配置" else "TMDB 未配置")
            button("编辑 TMDB") {
                choose("凭据类型", listOf("API Key", "Token")) { index ->
                    form("编辑 TMDB", listOf("语言", "密钥"), listOf(data.optString("language").ifBlank { "zh-CN" })) { values ->
                        val body = json("credential_type" to if (index == 0) "api_key" else "token",
                            "language" to values[0], "credential" to values[1])
                        task({ api.request(path, "PUT", body) }) { metadata() }
                    }
                }
            }
            button("测试连接") {
                task({ api.request("$path/test", "POST", JSONObject()) }) { message("连接成功") }
            }
            if (data.optBoolean("configured")) button("移除凭据") {
                confirm("移除 TMDB 凭据？") { task({ api.request(path, "DELETE") }) { metadata() } }
            }
        }
    }

    private fun system(): Unit = with(activity) {
        page("系统设置", { libraries() }, { catalog.show() })
        button("刷新") { system() }
        val api = client
        task({ api.request("/api/admin/system/disks") }) { data ->
            data.rows("disks").forEach { disk ->
                label(disk.getString("mount_point"))
                val used = disk.getLong("used_bytes") / 1073741824.0
                val free = disk.getLong("available_bytes") / 1073741824.0
                label("已用 %.1f GiB · 可用 %.1f GiB".format(used, free))
            }
            if (data.optBoolean("incomplete")) label("部分磁盘无法读取")
        }
    }
}
