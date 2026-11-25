package com.monkeydog.dinosaur.mobile

import com.monkeydog.dinosaur.android.Api
import com.monkeydog.dinosaur.android.Cue
import com.monkeydog.dinosaur.android.vttCues
import com.monkeydog.dinosaur.android.rows
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
