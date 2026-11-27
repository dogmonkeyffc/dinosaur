package com.monkeydog.dinosaur.mobile

import android.os.SystemClock
import com.monkeydog.dinosaur.android.Api
import com.monkeydog.dinosaur.android.savePlaybackProgress

internal class PlaybackProgress(private val activity: MainActivity, private val api: Api) {
    private var lastSave = 0L

    fun due() = SystemClock.elapsedRealtime() - lastSave >= 10000

    fun save(id: Long, position: Long, watched: Boolean) {
        lastSave = SystemClock.elapsedRealtime()
        activity.worker.execute { runCatching { api.savePlaybackProgress(id, position / 1000.0, watched) } }
    }
}
