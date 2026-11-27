package com.monkeydog.dinosaur.tv

import android.widget.Button
import android.widget.LinearLayout
import com.monkeydog.dinosaur.android.Account
import com.monkeydog.dinosaur.android.Api
import com.monkeydog.dinosaur.android.continueWatching
import com.monkeydog.dinosaur.android.rows
import java.net.URLEncoder
import org.json.JSONObject

internal class TvCatalog(private val activity: TvActivity, private val account: Account) {
    private val api = account.api
    private val details = TvDetails(activity, api)
    private val videos = TvVideoFolders(activity, api)
    private val history = TvHistory(activity, api)
    private var kind = "movie"
    private var query = ""
    private var library = 0L
    private var pageNumber = 1
    private var revision = 0
    private var libraries = emptyList<JSONObject>()
    private val kinds = linkedMapOf("movie" to "电影", "tv" to "电视剧", "video" to "普通视频")

    fun home() = with(activity) {
        val nav = column().apply {
            setBackgroundColor(0xff172024.toInt())
            setPadding(dp(20), dp(24), dp(14), dp(24))
            addView(label("Dinosaur TV", 22f, accent, bold = true).apply { setPadding(0, 0, 0, dp(21)) })
        }
        val body = column().apply { setPadding(dp(24), dp(24), dp(32), dp(22)) }
        val scroller = scroll(body)
        val container = row().apply {
            addView(nav, LinearLayout.LayoutParams(dp(180), -1))
            addView(scroller, LinearLayout.LayoutParams(0, -1, 1f))
        }
        val page = TvPage(container)
        val buttons = linkedMapOf<String, Button>()
        val toolbar = row()
        val listings = column()
        val search = control("搜索") {
            input("搜索片名", query) {
                query = it
                pageNumber = 1
                refresh(page, listings)
            }
        }
        kinds.forEach { (value, title) ->
            val destination = control(title) {
                kind = value
                library = 0
                pageNumber = 1
                buttons.forEach { (id, button) -> button.isSelected = id == value }
                scroller.scrollTo(0, 0)
                refresh(page, listings)
            }
            destination.isSelected = value == kind
            buttons[value] = destination
            nav.addControl(destination)
        }
        nav.addControl(control("观看记录", history::show))
        nav.addView(column(), LinearLayout.LayoutParams(1, 0, 1f))
        nav.addControl(control("退出登录") { logout(page, account) })
        toolbar.addControl(search, horizontal = true)
        toolbar.addControl(control("媒体库") {
            val choices = libraries.filter { it.optString("type") == kind }
            val names = arrayOf("全部媒体库", *choices.map { it.getString("name") }.toTypedArray())
            android.app.AlertDialog.Builder(this).setTitle("媒体库").setItems(names) { _, selected ->
                library = choices.getOrNull(selected - 1)?.getLong("id") ?: 0
                pageNumber = 1
                refresh(page, listings)
            }.show()
        }, horizontal = true)
        toolbar.addControl(control("刷新") { refresh(page, listings) }, horizontal = true)
        body.addView(toolbar)
        body.addView(listings)
        page.focus = buttons.getValue(kind)
        show(page, reset = true)
        task(page, { api.request("/api/libraries").rows("libraries") }, {}) { libraries = it }
        refresh(page, listings)
    }

    private fun refresh(page: TvPage, listing: LinearLayout): Unit = with(activity) {
        val request = ++revision
        val type = kind
        val number = pageNumber
        val folders = type == "video" && query.isBlank()
        val path = if (type == "tv") "series" else "media"
        val encoded = URLEncoder.encode(query, "UTF-8")
        val filter = if (library > 0) "&library_id=$library" else ""
        val recent = query.isBlank() && number == 1
        page.resume = {}
        listing.removeAllViews()
        val status = label("正在加载", color = muted)
        listing.addView(status)
        task(page, {
            val catalog = if (folders) api.request("/api/libraries")
            else api.request("/api/$path?type=$type&page=$number&page_size=24&q=$encoded$filter")
            val resume = if (recent) runCatching { recent(type) }.getOrDefault(emptyList()) else emptyList()
            catalog to resume
        }, {
            if (request == revision) status.text = it.message ?: "无法加载片库"
        }) { (catalog, resume) ->
            if (request == revision) {
                listing.removeAllViews()
                val recentPane = column()
                listing.addView(recentPane)
                renderRecent(page, recentPane, resume, type)
                val first = if (folders) renderVideoLibraries(listing, catalog) else render(page, listing, catalog, type) {
                    pageNumber = it
                    refresh(page, listing)
                }
                page.resume = {
                    if (recent && request == revision) task(page, { recent(type) }, {}) { updates ->
                        val focus = page.view.findFocus()?.tag as? RecentFocus
                        val replacement = renderRecent(page, recentPane, updates, type)[focus?.mediaId]
                        if (focus != null) (replacement ?: first ?: page.focus)?.requestFocus()
                    }
                }
            }
        }
    }

    private data class RecentFocus(val mediaId: Long)

    private fun recent(type: String) = api.continueWatching(type)

    private fun renderRecent(page: TvPage, pane: LinearLayout, media: List<JSONObject>, type: String) = with(activity) {
        val items = media.filter { library == 0L || it.optLong("library_id") == library }
        pane.removeAllViews()
        if (items.isNotEmpty()) {
            pane.addView(label("继续观看", 22f, bold = true).apply { setPadding(0, dp(14), 0, dp(10)) })
        }
        tvGrid(pane, items, api) {
            if (type == "video") videos.play(page, it) else details.play(it.getLong("id"))
        }.mapKeys { it.key.getLong("id") }
            .also { views -> views.forEach { (id, view) -> view.tag = RecentFocus(id) } }
    }

    private fun renderVideoLibraries(listing: LinearLayout, catalog: JSONObject) = with(activity) {
        listing.addView(label("普通视频", 25f, bold = true).apply { setPadding(0, dp(14), 0, dp(10)) })
        val items = catalog.rows("libraries").filter {
            it.optString("type") == "video" && (library == 0L || library == it.optLong("id"))
        }
        if (items.isEmpty()) listing.addView(label("暂无媒体库", color = muted))
        items.map { item ->
            control(item.getString("name")) { videos.library(item) }.also { listing.addControl(it) }
        }.firstOrNull()
    }

    private fun render(
        page: TvPage,
        listing: LinearLayout,
        catalog: JSONObject,
        type: String,
        go: (Int) -> Unit,
    ) = with(activity) {
        val items = catalog.rows("items")
        val title = if (query.isBlank()) kinds.getValue(type) else "搜索：$query"
        listing.addView(label(title, 25f, bold = true).apply { setPadding(0, dp(14), 0, dp(10)) })
        if (items.isEmpty()) listing.addView(label("暂无媒体", color = muted))
        val views = tvGrid(listing, items, api) { item ->
            val id = item.getLong("id")
            val work = item.optLong("work_id")
            when {
                type == "video" -> videos.play(page, item)
                type != "tv" -> details.movie(id)
                work > 0 -> details.work(work)
                else -> details.series(id)
            }
        }
        val pages = ((catalog.optInt("total") + 23) / 24).coerceAtLeast(1)
        if (pages > 1) listing.addView(row().apply {
            addControl(control("上一页") { go(pageNumber - 1) }.apply { isEnabled = pageNumber > 1 }, horizontal = true)
            addControl(control("下一页") { go(pageNumber + 1) }.apply { isEnabled = pageNumber < pages }, horizontal = true)
            addView(label("$pageNumber / $pages", color = muted))
        })
        views.values.firstOrNull()
    }
}

internal fun TvActivity.tvGrid(
    container: LinearLayout,
    items: List<JSONObject>,
    api: Api,
    select: (JSONObject) -> Unit,
): Map<JSONObject, LinearLayout> {
    val columns = ((resources.configuration.screenWidthDp - 210) / 125).coerceIn(3, 8)
    val views = linkedMapOf<JSONObject, LinearLayout>()
    items.chunked(columns).forEach { chunk ->
        val line = row().apply { gravity = android.view.Gravity.TOP }
        chunk.forEach { item ->
            val card = artwork.tile(item, api) { select(item) }
            views[item] = card
            line.addView(card, LinearLayout.LayoutParams(0, -2, 1f).apply {
                rightMargin = dp(10)
                bottomMargin = dp(12)
            })
        }
        if (chunk.size < columns) line.addView(android.view.View(this), LinearLayout.LayoutParams(0, 1, (columns - chunk.size).toFloat()))
        container.addView(line)
    }
    return views
}
