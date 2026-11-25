package com.monkeydog.dinosaur.mobile

import com.monkeydog.dinosaur.android.Api
import com.monkeydog.dinosaur.android.rows
import android.graphics.BitmapFactory
import android.graphics.Color
import android.view.View
import android.widget.ImageView
import android.widget.LinearLayout
import android.widget.TextView
import org.json.JSONObject
import java.net.HttpURLConnection
import java.net.URL
import java.net.URLEncoder
import java.util.concurrent.Executors

internal class Catalog(private val activity: MainActivity) {
    private val artwork = Executors.newFixedThreadPool(2)
    private var currentTab = 0
    private var currentLibrary = 0L

    fun show(tab: Int = currentTab, library: Long = currentLibrary, query: String = "", page: Int = 1): Unit = with(activity) {
        currentTab = tab
        currentLibrary = library
        page(listOf("首页", "片库", "观看记录")[tab], if (tab == 0) null else ({ show(0) }))
        button("头像与设置") { settings.show() }
        val tabs = LinearLayout(this)
        listOf("首页", "片库", "记录").forEachIndexed { index, title ->
            tabs.addView(android.widget.Button(this).apply {
                text = title
                isEnabled = index != tab
                setOnClickListener { show(index, if (index == 1) 0 else currentLibrary) }
            }, LinearLayout.LayoutParams(0, -2, 1f))
        }
        content.addView(tabs)
        val client = api
        if (tab == 2) {
            task({ client.request("/api/history?page=$page&page_size=24") }) { data ->
                var previous = ""
                data.rows("items").forEach { item ->
                    val date = item.getString("watched_on")
                    val month = date.take(7)
                    if (previous != month) label(month)
                    previous = month
                    card(item, client, date) {
                        labelHistory(item)
                    }
                }
                pagination(data, page) { show(2, library, query, it) }
            }
        } else {
            task({ client.request("/api/libraries").rows("libraries") }) { libraries ->
                val options = libraries.filter { tab == 1 || !settings.hidden(it.getLong("id")) }
                val selected = options.firstOrNull { it.getLong("id") == library } ?: options.firstOrNull()
                if (selected == null) {
                    label("暂无媒体库")
                } else {
                    val id = selected.getLong("id")
                    currentLibrary = id
                    button(selected.getString("name")) {
                        choose("媒体库", options.map { it.getString("name") }) { show(tab, options[it].getLong("id")) }
                    }
                    button(if (query.isBlank()) "搜索媒体" else "搜索：$query") {
                        form("搜索媒体", listOf("片名"), listOf(query)) { show(tab, id, it.first()) }
                    }
                    val kind = selected.getString("type")
                    val encoded = URLEncoder.encode(query, "UTF-8")
                    val route = if (kind == "tv") "series" else "media"
                    task({
                        val data = client.request("/api/$route?type=$kind&library_id=$id&page=$page&page_size=24&q=$encoded")
                        val recent = if (tab == 0 && query.isBlank() && page == 1) {
                            client.request("/api/playback/continue?type=$kind").rows("items")
                                .filter { it.optLong("library_id") == id }
                        } else emptyList()
                        data to recent
                    }) { (data, recent) ->
                        if (recent.isNotEmpty()) label("继续观看")
                        recent.forEach { item -> card(item, client) { play(item.getLong("id")) } }
                        if (recent.isNotEmpty()) label(selected.getString("name"))
                        data.rows("items").forEach { item ->
                            card(item, client) {
                                if (kind == "tv") {
                                    val work = item.optLong("work_id")
                                    if (work > 0) work(work) else series(item.getLong("id"), { show() })
                                } else movie(item.getLong("id"))
                            }
                        }
                        pagination(data, page) { show(tab, id, query, it) }
                    }
                }
            }
        }
    }

    private fun MainActivity.labelHistory(item: JSONObject) {
        android.app.AlertDialog.Builder(this).setTitle(item.getString("title"))
            .setMessage(listOf(item.optString("watched_on"), item.optInt("rating").takeIf { it > 0 }?.let { "★".repeat(it.coerceAtMost(5)) }.orEmpty(),
                item.optString("review")).filter(String::isNotBlank).joinToString("\n"))
            .setPositiveButton("关闭", null).show()
    }

    private fun MainActivity.pagination(data: JSONObject, page: Int, move: (Int) -> Unit) {
        if (data.optInt("total") == 0) label("暂无内容")
        if (page > 1) button("上一页") { move(page - 1) }
        if (page * 24 < data.optInt("total")) button("下一页") { move(page + 1) }
    }

    private fun movie(id: Long): Unit = with(activity) {
        page("详情", { show() })
        val client = api
        task({ client.request("/api/media/$id").getJSONObject("item") }) { item ->
            card(item, client) { play(id, { movie(id) }) }
            button(if (item.optDouble("position", 0.0) > 0 && !item.optBoolean("watched")) "继续观看" else "播放") {
                play(id, { movie(id) })
            }
            item.optJSONObject("metadata")?.optString("overview")?.takeIf(String::isNotBlank)?.let { label(it) }
        }
    }

    private fun work(id: Long): Unit = with(activity) {
        page("剧集合集", { show() })
        val client = api
        task({ client.request("/api/works/$id") }) { data ->
            data.rows("members").forEach { item ->
                card(item, client) { series(item.getLong("id")) { work(id) } }
            }
        }
    }

    private fun series(id: Long, previous: () -> Unit): Unit = with(activity) {
        page("剧集", previous)
        val client = api
        task({ client.request("/api/series/$id") }) { data ->
            label(data.getJSONObject("series").getString("title"))
            data.optJSONObject("resume")?.let { item ->
                button("播放 · " + item.getString("title")) { play(item.getLong("id")) { series(id, previous) } }
            }
            data.rows("seasons").forEach { season ->
                label("第 ${season.getInt("number")} 季")
                season.rows("episodes").forEach { item ->
                    button((if (item.optBoolean("watched")) "✓ " else "") + item.getString("title")) {
                        play(item.getLong("id")) { series(id, previous) }
                    }
                }
            }
        }
    }

    private fun play(id: Long, close: () -> Unit = { show() }) {
        activity.playback?.release()
        activity.playback = PlaybackScreen(activity, activity.api, close).apply { open(id) }
    }

    private fun MainActivity.card(item: JSONObject, client: Api, detail: String = "", open: () -> Unit) {
        val row = LinearLayout(this).apply {
            gravity = android.view.Gravity.CENTER_VERTICAL
            setPadding(dp(12), dp(8), dp(12), dp(8))
            setBackgroundColor(Color.rgb(38, 39, 44))
            isClickable = true
            isFocusable = true
            setOnClickListener { open() }
        }
        val text = TextView(this).apply {
            setTextColor(Color.WHITE)
            textSize = 16f
            text = listOf(item.getString("title"), detail).filter(String::isNotBlank).joinToString("\n")
            setPadding(dp(12), 0, 0, 0)
        }
        val poster = item.optString("poster_url").takeIf(String::isNotBlank)
        if (poster != null) {
            val image = ImageView(this).apply { scaleType = ImageView.ScaleType.CENTER_CROP }
            row.addView(image, LinearLayout.LayoutParams(dp(54), dp(72)))
            artwork.execute {
                val bitmap = runCatching {
                    val connection = URL(client.resource(poster).toString()).openConnection() as HttpURLConnection
                    try {
                        connection.instanceFollowRedirects = false
                        connection.connectTimeout = 5000
                        connection.readTimeout = 5000
                        client.headers.forEach(connection::setRequestProperty)
                        if (connection.responseCode == 200) {
                            connection.inputStream.use { BitmapFactory.decodeStream(it, null, BitmapFactory.Options().apply { inSampleSize = 2 }) }
                        } else null
                    } finally { connection.disconnect() }
                }.getOrNull()
                runOnUiThread { if (image.isAttachedToWindow) image.setImageBitmap(bitmap) }
            }
        }
        row.addView(text, LinearLayout.LayoutParams(0, -2, 1f))
        content.addView(row, LinearLayout.LayoutParams(-1, -2).apply { bottomMargin = dp(6) })
    }

    fun dispose() { artwork.shutdownNow() }
}
