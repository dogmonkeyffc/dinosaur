package com.monkeydog.dinosaur.mobile

import android.content.res.ColorStateList
import android.text.TextUtils
import android.view.Gravity
import android.view.View
import android.widget.FrameLayout
import android.widget.ImageView
import android.widget.LinearLayout
import android.widget.ProgressBar
import com.monkeydog.dinosaur.android.Api
import org.json.JSONObject
import kotlin.math.roundToInt

private val episodeCode = Regex("S\\d+E\\d+(?:\\s*[-–]\\s*E?\\d+)?", RegexOption.IGNORE_CASE)

internal fun MainActivity.seasonSummary(
    series: JSONObject, fallbackTitle: String, episodes: List<JSONObject>, artwork: PhoneArtwork, client: Api,
) = phoneRow().apply {
    setPadding(0, ui(6), 0, ui(14))
    val name = series.optJSONObject("metadata")?.optString("title").orEmpty().ifBlank { fallbackTitle }
    if (series.optString("poster_url").let { it.isNotBlank() && it != "null" }) {
        addView(artwork.poster(series, client, 14), LinearLayout.LayoutParams(ui(68), ui(102)).apply {
            rightMargin = ui(18)
        })
    }
    addView(column().apply {
        addView(phoneText(name, 26, bold = true))
        if (episodes.isNotEmpty()) {
            val watched = episodes.count { it.optBoolean("watched") }
            val count = "${episodes.size} 集"
            val summary = if (watched > 0) "$count · 已看 $watched 集" else count
            addWithGap(phoneText(summary, 15, PHONE_MUTED), ui(8))
        }
    }, LinearLayout.LayoutParams(0, -2, 1f))
}

internal fun MainActivity.episodeRow(item: JSONObject, play: () -> Unit) = column().apply {
    val sourceTitle = item.optString("title")
    val number = item.optInt("episode_number")
    val end = item.optInt("episode_end")
    val numberLabel = when {
        number <= 0 -> ""
        end > number -> "第 $number–$end 集"
        else -> "第 $number 集"
    }
    val generatedTitle = sourceTitle.isBlank() || episodeCode.matches(sourceTitle)
    val title = if (generatedTitle && numberLabel.isNotBlank()) numberLabel else sourceTitle.ifBlank { item.optString("filename") }
    val watched = item.optBoolean("watched")
    val missing = item.optBoolean("missing")
    val position = item.optDouble("position", 0.0)
    val duration = item.optDouble("duration", 0.0)
    val resume = !watched && !missing && position > 0
    val facts = listOfNotNull(
        numberLabel.takeIf { it.isNotBlank() && it != title },
        duration.takeIf { it.isFinite() && it > 0 }?.let { clock((it * 1000).toLong()) },
        when {
            missing -> "文件不可用"
            watched -> "已看"
            resume -> "看到 ${clock((position * 1000).toLong())}"
            else -> null
        },
    ).joinToString(" · ")

    background = panel(if (resume) PHONE_SELECTED else PHONE_SURFACE, 20)
    clipToOutline = true
    minimumHeight = ui(68)
    setPadding(ui(18), ui(12), ui(14), ui(12))
    addView(phoneRow().apply {
        addView(column().apply {
            addView(phoneText(title, 19, bold = true).apply {
                maxLines = 2
                ellipsize = TextUtils.TruncateAt.END
            })
            if (facts.isNotEmpty()) addWithGap(phoneText(facts, 14, if (resume) PHONE_ACCENT else PHONE_MUTED), ui(5))
        }, LinearLayout.LayoutParams(0, -2, 1f))
        if (!missing) addView(FrameLayout(this@episodeRow).apply {
            background = panel(if (resume) PHONE_ACCENT else 0xff1d1e21.toInt(), 22)
            addView(ImageView(this@episodeRow).apply {
                setImageResource(R.drawable.player_play)
                setColorFilter(if (resume) PHONE_SELECTED else PHONE_ACCENT)
                importantForAccessibility = View.IMPORTANT_FOR_ACCESSIBILITY_NO
            }, FrameLayout.LayoutParams(ui(22), ui(22), Gravity.CENTER))
        }, LinearLayout.LayoutParams(ui(40), ui(40)).apply { leftMargin = ui(12) })
    })
    if (resume && duration.isFinite() && duration > 0) {
        addWithGap(ProgressBar(this@episodeRow, null, android.R.attr.progressBarStyleHorizontal).apply {
            progress = (position / duration * 100).coerceIn(0.0, 100.0).roundToInt()
            progressTintList = ColorStateList.valueOf(PHONE_ACCENT)
            progressBackgroundTintList = ColorStateList.valueOf(0xff445b7c.toInt())
            importantForAccessibility = View.IMPORTANT_FOR_ACCESSIBILITY_NO
        }, ui(9), -1, ui(3))
    }
    contentDescription = listOf(title, facts).filter(String::isNotBlank).joinToString("，")
    if (missing) isEnabled = false else onTap(play)
}
