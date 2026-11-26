package com.monkeydog.dinosaur.mobile

import android.content.res.ColorStateList
import android.text.TextUtils
import android.view.Gravity
import android.view.View
import android.widget.LinearLayout
import android.widget.ProgressBar
import android.widget.ScrollView
import com.monkeydog.dinosaur.android.Api
import com.monkeydog.dinosaur.android.rows
import org.json.JSONObject
import kotlin.math.roundToInt

internal class MediaDetails(
    private val activity: MainActivity,
    private val artwork: PhoneArtwork,
    private val play: (Long, () -> Unit) -> Unit,
) {
    fun open(item: JSONObject, kind: String, back: () -> Unit) {
        val id = item.getLong("id")
        val work = item.optLong("work_id")
        when {
            kind != "tv" -> movie(id, back)
            work > 0 -> work(work, item.getString("title"), back)
            else -> series(id, item.getString("title"), back)
        }
    }

    private fun movie(id: Long, back: () -> Unit): Unit = with(activity) {
        page("", back)
        val client = api
        task({ client.request("/api/media/$id").getJSONObject("item") }) { item ->
            val position = item.optDouble("position", 0.0)
            val resume = position > 0 && !item.optBoolean("watched")
            val duration = item.optDouble("duration", 0.0)
            val facts = listOfNotNull(item.optInt("year").takeIf { it > 0 }?.toString(),
                duration.takeIf { it > 0 }?.let { "${(it / 60).roundToInt()} 分钟" }).joinToString(" · ")
            val posterWidth = minOf(dp(156), ((resources.displayMetrics.widthPixels - dp(32)) * .39).roundToInt())
            content.addView(phoneRow().apply {
                if (item.optString("poster_url").let { it.isNotBlank() && it != "null" }) {
                    addView(artwork.poster(item, client, 14), LinearLayout.LayoutParams(posterWidth, (posterWidth * 1.5).roundToInt())
                        .apply { rightMargin = dp(18) })
                }
                addView(column().apply {
                    addView(phoneText(item.getString("title"), 31, bold = true))
                    if (facts.isNotBlank()) addWithGap(phoneText(facts, 16, PHONE_MUTED), dp(8))
                    addWithGap(primary(if (resume) "继续观看" else "播放") { play(id) { movie(id, back) } }, dp(12), -2)
                    if (resume && duration > 0) addWithGap(ProgressBar(this@with, null, android.R.attr.progressBarStyleHorizontal).apply {
                        progress = (position / duration * 100).roundToInt().coerceIn(0, 100)
                        progressTintList = ColorStateList.valueOf(PHONE_ACCENT)
                        progressBackgroundTintList = ColorStateList.valueOf(PHONE_BORDER)
                        contentDescription = "观看进度"
                    }, dp(3), minOf(dp(168), resources.displayMetrics.widthPixels - posterWidth - dp(50)), dp(3))
                }, LinearLayout.LayoutParams(0, -2, 1f))
            })
            val description = item.optJSONObject("metadata")?.optString("overview").orEmpty().trim()
            if (description.isNotEmpty()) overview(description)
        }
    }

    private fun MainActivity.overview(summary: String) {
        content.addWithGap(View(this).apply { setBackgroundColor(PHONE_BORDER) }, dp(12), -1, dp(1))
        content.addWithGap(phoneText("简介", 20, bold = true), dp(16))
        val copy = phoneText(summary, 18, PHONE_MUTED).apply {
            setLineSpacing(dp(4).toFloat(), 1f)
            maxLines = 4
            ellipsize = TextUtils.TruncateAt.END
        }
        content.addWithGap(copy, dp(8))
        var expanded = false
        lateinit var disclosure: android.widget.TextView
        disclosure = phoneAction("展开简介", {
            expanded = !expanded
            copy.maxLines = if (expanded) Int.MAX_VALUE else 4
            disclosure.text = if (expanded) "收起简介" else "展开简介"
        }, fill = 0)
        content.addView(disclosure, LinearLayout.LayoutParams(-2, -2))
        copy.post {
            val truncated = copy.layout?.let { it.getEllipsisCount(it.lineCount - 1) > 0 } ?: false
            if (!truncated) disclosure.visibility = View.GONE
        }
    }

    private fun work(id: Long, title: String, back: () -> Unit, offset: Int = 0): Unit = with(activity) {
        page(title, back)
        val client = api
        task({ client.request("/api/works/$id") }) { data ->
            val work = data.optJSONObject("work")
            if (work != null) label("${work.optInt("member_count")} 篇章 · ${work.optInt("season_count")} 季 · ${work.optInt("episode_count")} 集")
            val members = data.rows("members")
            grid(members, client) {
                val position = detailScroll().scrollY
                series(it.getLong("id"), it.getString("title"), { work(id, title, back, position) })
            }
            if (members.isEmpty()) label("暂无篇章")
            restoreDetailScroll(offset)
        }
    }

    private fun series(id: Long, title: String, back: () -> Unit, offset: Int = 0): Unit = with(activity) {
        page(title, back)
        val client = api
        task({ client.request("/api/series/$id") }) { data ->
            val series = data.getJSONObject("series")
            val year = series.optInt("year")
            val overview = series.optJSONObject("metadata")?.optString("overview").orEmpty()
            if (year > 0 || overview.isNotBlank()) content.addView(column().apply {
                background = panel(PHONE_SURFACE, 14)
                setPadding(dp(14), dp(14), dp(14), dp(14))
                if (year > 0) addView(phoneText(year.toString(), 16, PHONE_MUTED))
                if (overview.isNotBlank()) addWithGap(phoneText(overview, 18), dp(8))
            })
            data.optJSONObject("resume")?.let { episode ->
                val verb = if (episode.optDouble("position", 0.0) > 0 && !episode.optBoolean("watched")) "继续观看" else "播放"
                content.addWithGap(primary("$verb · ${episode.getString("title")}") {
                    val position = detailScroll().scrollY
                    play(episode.getLong("id")) { series(id, title, back, position) }
                }, dp(14), -2)
            }
            val seasons = data.rows("seasons").filter { it.optInt("number", -1) >= 0 }.map { season ->
                val number = season.getInt("number")
                val poster = season.optString("poster_url")
                JSONObject().put("number", number).put("title", seasonName(number)).put("poster_url", poster)
            }
            grid(seasons, client) { selected ->
                val position = detailScroll().scrollY
                season(id, title, selected.getInt("number"), { series(id, title, back, position) })
            }
            if (seasons.isEmpty()) label("暂无季度")
            restoreDetailScroll(offset)
        }
    }

    private fun season(id: Long, title: String, number: Int, back: () -> Unit, offset: Int = 0): Unit = with(activity) {
        page(seasonName(number), back)
        val client = api
        task({ client.request("/api/series/$id") }) { data ->
            val series = data.getJSONObject("series")
            val selected = data.rows("seasons").firstOrNull { it.optInt("number") == number }
            val episodes = selected?.rows("episodes").orEmpty()
            val cover = JSONObject().put("title", series.optString("title"))
                .put("poster_url", selected?.optString("poster_url").orEmpty())
                .put("metadata", series.optJSONObject("metadata"))
            content.addView(seasonSummary(cover, title, episodes, artwork, client))
            episodes.forEach { item ->
                content.addWithGap(episodeRow(item) {
                    val position = detailScroll().scrollY
                    play(item.getLong("id")) { season(id, title, number, back, position) }
                }, ui(10))
            }
            if (episodes.isEmpty()) label("暂无分集")
            restoreDetailScroll(offset)
        }
    }

    private fun MainActivity.grid(items: List<JSONObject>, client: Api, open: (JSONObject) -> Unit) {
        val available = resources.displayMetrics.widthPixels - dp(32)
        val columns = maxOf(2, (available + dp(12)) / dp(130))
        val width = (available - dp(12) * (columns - 1)) / columns
        items.chunked(columns).forEach { group ->
            content.addWithGap(phoneRow().apply {
                gravity = Gravity.TOP
                group.forEachIndexed { index, item ->
                    val title = item.getString("title")
                    addView(column().apply {
                        addView(artwork.poster(item, client), LinearLayout.LayoutParams(-1, width * 3 / 2))
                        addWithGap(phoneText(title, 17).apply { maxLines = 2 }, dp(8))
                        contentDescription = title
                        onTap { open(item) }
                    }, LinearLayout.LayoutParams(width, -2).apply { if (index > 0) leftMargin = dp(12) })
                }
            }, dp(14))
        }
    }

    private fun MainActivity.primary(title: String, run: () -> Unit) =
        phoneAction(title, run, size = 18, color = PHONE_BG, fill = PHONE_ACCENT)

    private fun MainActivity.detailScroll() = content.parent as ScrollView

    private fun MainActivity.restoreDetailScroll(offset: Int) {
        if (offset > 0) detailScroll().apply { post { scrollTo(0, offset) } }
    }

    private fun seasonName(number: Int) = if (number == 0) "特别篇" else "第 $number 季"
}
