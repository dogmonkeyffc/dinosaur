package com.monkeydog.dinosaur.mobile

import java.net.URLEncoder

internal class ScrapeResults(private val activity: MainActivity) {
    fun show(previous: () -> Unit, page: Int = 1, query: String = "", library: Long? = null): Unit = with(activity) {
        page("刮削结果", previous)
        val client = api
        button("搜索") { form("搜索", listOf("片名"), listOf(query)) { show(previous, query = it[0], library = library) } }
        button("筛选媒体库") {
            task({ client.request("/api/admin/libraries").rows("libraries") }) { items ->
                choose("筛选媒体库", listOf("全部") + items.map { it.getString("name") }) {
                    show(previous, query = query, library = items.getOrNull(it - 1)?.getLong("id"))
                }
            }
        }
        val filter = library?.let { "&library_id=$it" }.orEmpty()
        val path = "/api/admin/metadata/records?page=$page&page_size=20&q=${URLEncoder.encode(query, "UTF-8")}$filter"
        task({ client.request(path) }) { data ->
            val rows = data.rows("items")
            if (rows.isEmpty()) label("暂无结果")
            rows.forEach { record ->
                val title = record.optString("title").ifBlank { record.optString("filename") }
                val status = record.optString("status")
                button("$title · $status") {
                    val tv = record.optLong("series_id") > 0L
                    val scope = if (tv) "series" else "media"
                    val id = record.optLong(if (tv) "series_id" else "media_id")
                    val jobs = if (tv) "/api/admin/tv/metadata/jobs" else "/api/admin/metadata/jobs"
                    choose(title, listOf("重新刮削", "查看匹配候选")) { action ->
                        if (action == 0) {
                            val field = if (tv) "series_id" else "media_id"
                            task({ client.request(jobs, "POST", json(field to id, "refresh" to true)) }) { message("已提交任务") }
                        } else {
                            task({ client.request("/api/admin/$scope/$id/metadata/candidates").rows("candidates") }) { items ->
                                val labels = items.map { "${it.optString("title")} ${it.optInt("year").takeIf { year -> year > 0 } ?: ""}" }
                                if (items.isEmpty()) message("暂无候选")
                                else choose(title, labels) { index ->
                                    val candidate = items[index]
                                    val source = if (candidate.optString("provider") == "bangumi") "bangumi_id" else "tmdb_id"
                                    val field = if (tv) "series_id" else "media_id"
                                    confirm("使用 ${labels[index]}？") {
                                        task({ client.request(jobs, "POST", json(field to id, source to candidate.getLong("id"))) }) {
                                            message("已提交匹配")
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }
            if (page > 1) button("上一页") { show(previous, page - 1, query, library) }
            if (page * 20 < data.optInt("total")) button("下一页") { show(previous, page + 1, query, library) }
        }
    }

    fun job(id: Long, previous: () -> Unit): Unit = with(activity) {
        page("刮削任务", previous)
        val client = api
        task({ client.request("/api/admin/metadata/jobs/$id") }) { data ->
            val job = data.optJSONObject("job")
            if (job != null) label("${job.optString("status")} · ${job.optInt("processed")}/${job.optInt("total")}")
            val results = data.optJSONObject("results") ?: data
            results.rows("items").forEach { label("${it.optString("title")} · ${it.optString("status")}") }
        }
    }
}
