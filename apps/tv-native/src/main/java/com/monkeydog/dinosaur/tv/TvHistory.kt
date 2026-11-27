package com.monkeydog.dinosaur.tv

import android.app.AlertDialog
import android.widget.LinearLayout
import com.monkeydog.dinosaur.android.Api
import com.monkeydog.dinosaur.android.historyPage
import com.monkeydog.dinosaur.android.historySettings
import com.monkeydog.dinosaur.android.rows
import com.monkeydog.dinosaur.android.setHistoryRecording
import org.json.JSONObject

internal class TvHistory(private val activity: TvActivity, private val api: Api) {
    fun show(): Unit = with(activity) {
        val (page, content) = panelPage("观看记录")
        val listing = column()
        var number = 1
        var changed = false
        var loading = false
        var revision = 0
        fun load(target: Int, focus: Boolean = false) {
            val request = ++revision
            loading = true
            listing.removeAllViews()
            val status = label("正在加载", color = muted)
            listing.addView(status)
            task(page, { api.historyPage(target) }, {
                if (request != revision) return@task
                loading = false
                status.text = it.message ?: "无法加载观看记录"
                listing.addControl(control("重试") { load(target, focus) })
            }) { data ->
                if (request != revision) return@task
                loading = false
                val pages = ((data.optInt("total") + 23) / 24).coerceAtLeast(1)
                if (target > pages) {
                    load(pages, focus)
                } else {
                    number = target
                    render(listing, data.rows("items"), number, pages) { load(it, true) }
                    if (focus) (0 until listing.childCount).map(listing::getChildAt).firstOrNull { it.isFocusable }?.requestFocus()
                }
            }
        }
        val toolbar = row()
        val back = control("返回", ::back)
        toolbar.addControl(back, horizontal = true)
        toolbar.addControl(control("记录观看") { settings { changed = true } }, horizontal = true)
        toolbar.addControl(control("刷新") { load(1) }, horizontal = true)
        content.addView(toolbar)
        content.addView(listing)
        page.focus = back
        page.resume = {
            if (changed || loading) {
                changed = false
                load(number)
            }
        }
        show(page)
        load(1)
    }

    private fun render(listing: LinearLayout, items: List<JSONObject>, number: Int, pages: Int, go: (Int) -> Unit): Unit =
        with(activity) {
            listing.removeAllViews()
            if (items.isEmpty()) listing.addView(label("暂无观看记录", color = muted))
            var month = ""
            items.forEach { item ->
                val date = item.getString("watched_on")
                if (month != date.take(7)) {
                    month = date.take(7)
                    listing.addView(label(month, 19f, accent, bold = true).apply { setPadding(0, dp(10), 0, dp(7)) })
                }
                listing.addControl(control("${date.takeLast(2)} 日    ${item.getString("title")}") { detail(item) }.apply {
                    gravity = android.view.Gravity.CENTER_VERTICAL
                })
            }
            if (pages > 1) listing.addView(row().apply {
                addControl(control("上一页") { go(number - 1) }.apply { isEnabled = number > 1 }, horizontal = true)
                addControl(control("下一页") { go(number + 1) }.apply { isEnabled = number < pages }, horizontal = true)
                addView(label("$number / $pages", color = muted))
            })
        }

    private fun detail(item: JSONObject): Unit = with(activity) {
        val title = item.getString("title")
        val (page, content) = panelPage(title)
        val back = control("返回", ::back)
        content.addControl(back)
        val metadata = column()
        fun add(text: String) {
            metadata.addView(label(text).apply { setPadding(0, 0, 0, dp(12)) })
        }
        add(item.getString("watched_on"))
        item.optString("original_title").takeIf { !item.isNull("original_title") && it.isNotBlank() && it != title }?.let(::add)
        item.optInt("year").takeIf { it > 0 }?.let { add(it.toString()) }
        item.optInt("rating").coerceIn(0, 5).takeIf { it > 0 }?.let { add("★".repeat(it) + "☆".repeat(5 - it)) }
        item.optString("review").takeIf { !item.isNull("review") && it.isNotBlank() }?.let(::add)
        add(when (item.optString("source")) {
            "douban" -> "豆瓣"
            "local" -> "站内"
            else -> "手动记录"
        })
        content.addView(row().apply {
            gravity = android.view.Gravity.TOP
            if (item.optString("poster_url").isNotBlank()) {
                addView(artwork.poster(item, api), LinearLayout.LayoutParams(dp(126), dp(189)).apply { rightMargin = dp(24) })
            }
            addView(metadata, LinearLayout.LayoutParams(0, -2, 1f))
        })
        page.focus = back
        show(page)
    }

    private fun settings(onChange: () -> Unit): Unit = with(activity) {
        val (page, content) = panelPage("记录观看")
        val back = control("返回", ::back)
        content.addControl(back)
        val listing = column()
        content.addView(listing)
        page.focus = back
        show(page)

        fun load() {
            listing.removeAllViews()
            val status = label("正在加载", color = muted)
            listing.addView(status)
            task(page, api::historySettings, {
                status.text = it.message ?: "无法加载记录设置"
                listing.addControl(control("重试", ::load))
            }) { items ->
                listing.removeAllViews()
                if (items.isEmpty()) listing.addView(label("暂无媒体库", color = muted))
                items.forEach { item -> addSetting(page, listing, item, onChange) }
            }
        }
        load()
    }

    private fun addSetting(page: TvPage, listing: LinearLayout, item: JSONObject, onChange: () -> Unit) = with(activity) {
        val name = item.getString("name")
        val id = item.getLong("id")
        var record = item.optBoolean("record_history")
        fun title() = "$name · ${if (record) "已开启" else "已关闭"}"
        val button = control(title()) {}
        fun save() {
            val next = !record
            button.isClickable = false
            task(page, { api.setHistoryRecording(id, next) }, {
                button.isClickable = true
                message(it.message ?: "保存失败")
            }) {
                button.isClickable = true
                record = next
                button.text = title()
            }
        }
        button.setOnClickListener {
            val apply = {
                onChange()
                save()
            }
            if (record) AlertDialog.Builder(this).setTitle("关闭记录并清理该媒体库的进度和自动观看记录？")
                .setNegativeButton("取消", null)
                .setPositiveButton("关闭") { _, _ -> apply() }.show()
            else apply()
        }
        listing.addControl(button)
    }
}
