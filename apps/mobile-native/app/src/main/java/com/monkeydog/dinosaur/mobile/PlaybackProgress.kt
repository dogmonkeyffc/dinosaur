package com.monkeydog.dinosaur.mobile

import android.os.SystemClock
import com.monkeydog.dinosaur.android.Api
import com.monkeydog.dinosaur.android.json

internal class PlaybackProgress(private val activity: MainActivity, private val api: Api) {
    private var lastSave = 0L

    fun due() = SystemClock.elapsedRealtime() - lastSave >= 10000

    fun save(id: Long, position: Long, watched: Boolean) {
        lastSave = SystemClock.elapsedRealtime()
        val data = json("position" to position / 1000.0)
        if (watched) data.put("watched", true)
        activity.worker.execute { runCatching { api.request("/api/playback/$id/progress", "PUT", data) } }
    }
}
