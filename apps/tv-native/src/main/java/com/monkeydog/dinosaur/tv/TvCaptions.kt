package com.monkeydog.dinosaur.tv

import com.monkeydog.dinosaur.android.Api
import com.monkeydog.dinosaur.android.Cue
import com.monkeydog.dinosaur.android.rows
import com.monkeydog.dinosaur.android.vttCues
import org.json.JSONObject

internal class TvCaptions(private val activity: TvActivity, private val api: Api, private val page: TvPage) {
    private var cues = emptyList<Cue>()
    private var segment = -1L
    private var request = 0

    fun reset() {
        request++
        cues = emptyList()
        segment = -1L
    }

    fun at(mediaId: Long, selected: Int?, info: JSONObject, milliseconds: Long): String {
        val track = info.rows("subtitles").firstOrNull {
            it.optInt("id") == selected && it.optString("delivery") != "burn"
        } ?: return ""
        val index = if (track.optString("kind") == "embedded") milliseconds / 60000 else 0L
        if (segment != index) {
            segment = index
            val load = ++request
            val path = "/api/playback/$mediaId/subtitles/${track.getInt("id")}?start=${index * 60}"
            activity.task(page, { vttCues(api.text(path)) }, {}) { if (load == request) cues = it }
        }
        return cues.asSequence().filter { milliseconds >= it.start && milliseconds < it.end }.joinToString("\n") { it.text }
    }
}
