package com.monkeydog.dinosaur.tv

import android.widget.LinearLayout
import com.monkeydog.dinosaur.android.Api
import com.monkeydog.dinosaur.android.playback
import com.monkeydog.dinosaur.android.json
import com.monkeydog.dinosaur.android.rows
import org.json.JSONObject

internal class TvDetails(private val activity: TvActivity, private val api: Api) {
    fun play(id: Long) {
        TvPlayback(activity, api).show(id)
    }

    fun movie(id: Long) = detail("/api/media/$id") { content, page, data ->
        val item = data.getJSONObject("item")
        heading(content, item)
        val label = if (item.optDouble("position", 0.0) > 0 && !item.optBoolean("watched")) "继续观看" else "播放"
        val play = activity.control(label) { play(id) }
        content.addControl(play)
        play.requestFocus()
        page.resume = {
            activity.task(page, { api.playback(id) }, {}) {
                play.text = if (it.optDouble("resume_position", 0.0) > 0) "继续观看" else "播放"
            }
        }
    }

    fun work(id: Long) = detail("/api/works/$id") { content, _, data ->
        heading(content, data.getJSONObject("work"))
        val members = data.rows("members")
        activity.tvGrid(content, members, api) { series(it.getLong("id")) }
        if (members.isEmpty()) content.addView(activity.label("暂无篇章", color = muted))
    }

    fun series(id: Long) = detail("/api/series/$id") { content, page, data ->
        val item = data.getJSONObject("series")
        heading(content, item)
        var resume = data.optJSONObject("resume")
        val resumeButton = activity.control("") { resume?.let { play(it.getLong("id")) } }
        fun refreshResume() {
            val itemToResume = resume
            resumeButton.visibility = if (itemToResume == null) android.view.View.GONE else android.view.View.VISIBLE
            if (itemToResume != null) {
                val label = if (itemToResume.optDouble("position", 0.0) > 0 && !itemToResume.optBoolean("watched")) "继续观看" else "播放"
                resumeButton.text = "$label · ${itemToResume.getString("title")}"
            }
        }
        content.addControl(resumeButton)
        refreshResume()
        page.resume = {
            activity.task(page, { api.request("/api/series/$id") }, {}) {
                resume = it.optJSONObject("resume")
                refreshResume()
            }
        }
        val seasons = data.rows("seasons")
        val cards = seasons.map { season ->
            val number = season.getInt("number")
            json("title" to seasonTitle(number), "number" to number, "poster_url" to season.optString("poster_url"))
        }
        activity.tvGrid(content, cards, api) { card -> episodes(id, item.getString("title"), card.getInt("number")) }
        if (seasons.isEmpty()) content.addView(activity.label("暂无季度", color = muted))
    }

    private fun episodes(id: Long, title: String, number: Int) {
        detail("/api/series/$id") { content, page, data ->
            content.addView(activity.label("$title · ${seasonTitle(number)}", 31f, bold = true))
            val episodes = data.rows("seasons").firstOrNull { it.getInt("number") == number }?.rows("episodes").orEmpty()
            if (episodes.isEmpty()) content.addView(activity.label("暂无分集", color = muted))
            val buttons = episodes.associate { episode ->
                val id = episode.getLong("id")
                val title = episode.getString("title")
                val watched = if (episode.optBoolean("watched")) "✓ " else ""
                id to activity.control(watched + title) { play(id) }.also { content.addControl(it) }
            }
            page.resume = {
                activity.task(page, { api.request("/api/series/$id") }, {}) { refreshed ->
                    refreshed.rows("seasons").firstOrNull { it.getInt("number") == number }
                        ?.rows("episodes")?.forEach { item ->
                            val prefix = if (item.optBoolean("watched")) "✓ " else ""
                            buttons[item.getLong("id")]?.text = prefix + item.getString("title")
                        }
                }
            }
        }
    }

    private fun detail(path: String, render: (LinearLayout, TvPage, JSONObject) -> Unit) = with(activity) {
        val content = column().apply { setPadding(dp(48), dp(22), dp(48), dp(28)) }
        val page = TvPage(scroll(content))
        val toolbar = row()
        val back = control("返回", ::back)
        toolbar.addControl(back, horizontal = true)
        content.addView(toolbar)
        val status = label("正在加载", color = muted)
        content.addView(status)
        page.focus = back
        show(page)
        task(page, { api.request(path) }, { status.text = it.message ?: "无法加载详情" }) {
            content.removeView(status)
            render(content, page, it)
        }
    }

    private fun heading(content: LinearLayout, item: JSONObject) = with(activity) {
        content.addView(label(item.getString("title"), 31f, bold = true).apply { setPadding(0, dp(12), 0, dp(14)) })
        item.optJSONObject("metadata")?.optString("overview")?.takeIf(String::isNotBlank)?.let {
            content.addView(label(it, 17f, muted).apply { setPadding(0, 0, 0, dp(14)) })
        }
    }

    private fun seasonTitle(number: Int) = if (number == 0) "特别篇" else "第 $number 季"
}
