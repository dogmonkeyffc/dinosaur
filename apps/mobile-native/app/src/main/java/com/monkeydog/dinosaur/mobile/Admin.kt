package com.monkeydog.dinosaur.mobile

import com.monkeydog.dinosaur.android.json
import com.monkeydog.dinosaur.android.rows
import org.json.JSONObject

internal class Admin(private val activity: MainActivity) {
    private val client get() = activity.api

    fun libraries() = AdminLibraries(activity).show()

    fun users(): Unit = with(activity) {
        page("用户管理", { settings.show() })
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

    fun scraping(tv: Boolean = false): Unit = with(activity) {
        page(if (tv) "剧集刮削" else "电影刮削", { settings.show() })
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

    fun metadata(): Unit = with(activity) {
        page("元数据设置", { settings.show() })
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

    fun system(): Unit = with(activity) {
        page("系统设置", { settings.show() })
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
