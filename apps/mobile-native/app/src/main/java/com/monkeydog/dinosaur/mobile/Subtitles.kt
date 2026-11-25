package com.monkeydog.dinosaur.mobile

import android.text.Html

internal data class Cue(val start: Long, val end: Long, val text: String)

internal fun vttCues(source: String): List<Cue> {
    return source.replace("\r", "").split(Regex("\\n[ \\t]*\\n")).mapNotNull { block ->
        val lines = block.lines()
        val index = lines.indexOfFirst { " --> " in it }
        if (index < 0 || lines.firstOrNull()?.startsWith("NOTE") == true) return@mapNotNull null
        val points = lines[index].trim().split(" --> ", limit = 2)
        val start = vttTime(points[0]) ?: return@mapNotNull null
        val end = vttTime(points[1].substringBefore(' ')) ?: return@mapNotNull null
        val text = Html.fromHtml(lines.drop(index + 1).joinToString("<br>"), Html.FROM_HTML_MODE_LEGACY).toString()
        Cue(start, end, text)
    }
}

private fun vttTime(value: String): Long? {
    val parts = value.split(':')
    if (parts.size !in 2..3) return null
    val seconds = parts.last().toDoubleOrNull()?.takeIf { it.isFinite() && it in 0.0..60.0 } ?: return null
    val minutes = parts[parts.lastIndex - 1].toLongOrNull()?.takeIf { it in 0..59 } ?: return null
    val hours = if (parts.size == 3) parts.first().toLongOrNull()?.takeIf { it in 0..999 } ?: return null else 0
    return ((hours * 3600 + minutes * 60 + seconds) * 1000).toLong()
}

internal class PlaybackCaptions(private val activity: MainActivity, private val api: Api) {
    private var cues = emptyList<Cue>()
    private var window = -1L

    fun reset() {
        cues = emptyList()
        window = -1L
    }

    fun at(id: Long, selected: Int?, info: org.json.JSONObject, position: Long): String {
        val track = info.rows("subtitles").firstOrNull {
            it.optInt("id") == selected && it.optString("delivery") != "burn"
        } ?: return ""
        val next = if (track.optString("kind") == "embedded") position / 60000 else 0
        if (window != next) {
            window = next
            val path = "/api/playback/$id/subtitles/${track.getInt("id")}?start=${next * 60}"
            activity.task({ vttCues(api.text(path)) }, { }) {
                if (window == next) cues = it
            }
        }
        return cues.filter { position >= it.start && position < it.end }.joinToString("\n") { it.text }
    }
}
