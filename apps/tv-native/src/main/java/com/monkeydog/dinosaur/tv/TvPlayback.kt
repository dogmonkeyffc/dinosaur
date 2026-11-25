package com.monkeydog.dinosaur.tv

import android.view.KeyEvent
import androidx.media3.common.AudioAttributes
import androidx.media3.common.MediaItem
import androidx.media3.common.PlaybackException
import androidx.media3.common.Player
import androidx.media3.datasource.okhttp.OkHttpDataSource
import androidx.media3.exoplayer.ExoPlayer
import androidx.media3.exoplayer.source.DefaultMediaSourceFactory
import com.monkeydog.dinosaur.android.Api
import com.monkeydog.dinosaur.android.json
import com.monkeydog.dinosaur.android.rows
import org.json.JSONObject

@androidx.media3.common.util.UnstableApi
internal class TvPlayback(private val activity: TvActivity, private val api: Api) {
    private val controls = TvTransport(activity)
    private val page = TvPage(controls.root)
    private val captions = TvCaptions(activity, api, page)
    private var media: ExoPlayer? = null
    private var id = 0L
    private var next = 0L
    private var revision = 0
    private var offset = 0L
    private var duration = 0L
    private var sourceDuration = 0.0
    private var direct = false
    private var ended = false
    private var compatible = false
    private var quality = "auto"
    private var audio: Int? = null
    private var subtitle: Int? = null
    private var autoSubtitle = true
    private var speed = 1f
    private var info = JSONObject()
    private var lastSave = 0L
    private val tick = object : Runnable {
        override fun run() {
            val player = media ?: return
            val time = position()
            controls.update(time, duration)
            controls.caption(captions.at(id, subtitle, info, time))
            if (player.isPlaying && System.currentTimeMillis() - lastSave >= 10000) save()
            activity.main.postDelayed(this, 500)
        }
    }

    fun show(mediaId: Long) {
        page.leave = {
            revision++
            pause()
            detach()
            controls.close()
            activity.player = null
        }
        activity.player = this
        activity.show(page)
        open(mediaId)
    }

    private fun open(mediaId: Long, start: Long? = null, advance: Boolean = false,
                     autoplay: Boolean = advance || (media?.playWhenReady ?: true)) {
        val request = ++revision
        detach()
        captions.reset()
        id = mediaId
        next = 0
        ended = false
        if (advance) {
            audio = null
            subtitle = null
            autoSubtitle = true
        }
        val query = buildString {
            val mode = if (compatible) "compatible" else "original"
            append("mode=$mode&quality=$quality&start=${(start ?: 0L) / 1000.0}")
            audio?.let { append("&audio=$it") }
            if (autoSubtitle) append("&subtitle=auto")
            else if (info.rows("subtitles").any { it.optInt("id") == subtitle && it.optString("delivery") == "burn" }) {
                append("&subtitle=$subtitle")
            }
        }
        controls.loading()
        activity.task(page, { api.request("/api/playback/$mediaId?$query") }, {
            if (request == revision) controls.failure(it.message ?: "无法播放") { open(id, start, autoplay = autoplay) }
        }) { result ->
            if (request == revision) {
                runCatching { prepare(result, start, autoplay) }.onFailure {
                    controls.failure(it.message ?: "无法播放") { open(id, start, autoplay = autoplay) }
                }
            }
        }
    }

    private fun prepare(result: JSONObject, requested: Long?, autoplay: Boolean) {
        info = result
        val item = result.getJSONObject("item")
        val start = requested ?: (result.optDouble("resume_position", 0.0) * 1000).toLong()
        direct = result.getJSONObject("plan").getString("mode") == "direct"
        offset = if (direct) 0 else start
        sourceDuration = item.optDouble("duration", 0.0).takeIf { it.isFinite() && it > 0 } ?: 0.0
        duration = (sourceDuration * 1000).toLong()
        next = result.optJSONObject("next_episode")?.optLong("id") ?: 0L
        if (autoSubtitle) {
            subtitle = if (result.isNull("selected_subtitle")) null else result.getInt("selected_subtitle")
            autoSubtitle = false
        }
        val source = MediaItem.fromUri(api.stream(result.getString("stream_url"), offset))
        val http = OkHttpDataSource.Factory(Api.http).setDefaultRequestProperties(api.headers)
        val player = ExoPlayer.Builder(activity).setMediaSourceFactory(DefaultMediaSourceFactory(http)).build()
        media = player
        controls.view.player = player
        controls.ready(item.getString("title"), next > 0, ::toggle, ::seek, ::next, ::audioMenu, ::subtitleMenu, ::options)
        player.addListener(object : Player.Listener {
            override fun onPlaybackStateChanged(state: Int) {
                if (state != Player.STATE_ENDED || ended || media !== player) return
                val observed = offset + maxOf(player.currentPosition, player.duration.coerceAtLeast(0))
                if (duration > 0 && observed < duration - maxOf(5000L, duration / 100)) {
                    recover("播放提前中断")
                    return
                }
                ended = true
                save(true)
                if (next > 0) next() else controls.reveal()
            }

            override fun onIsPlayingChanged(isPlaying: Boolean) {
                if (media === player) controls.playing(isPlaying, ended)
            }

            override fun onPlayerError(error: PlaybackException) {
                if (media === player) recover("播放器无法播放此视频")
            }
        })
        player.setAudioAttributes(AudioAttributes.DEFAULT, true)
        player.setHandleAudioBecomingNoisy(true)
        player.setMediaItem(source, if (direct) start else 0)
        player.setPlaybackSpeed(speed)
        player.prepare()
        if (autoplay && activity.started) player.play()
        activity.main.post(tick)
    }

    private fun recover(message: String) {
        val resume = position()
        if (!compatible) {
            compatible = true
            open(id, resume)
        } else {
            detach()
            controls.failure(message) { open(id, resume) }
        }
    }

    private fun position() = offset + (media?.currentPosition ?: 0)

    private fun save(completed: Boolean = false) {
        if (media == null || (ended && !completed)) return
        val seconds = if (completed && sourceDuration > 0) sourceDuration else position() / 1000.0
        if (!completed && seconds <= 0) return
        val path = "/api/playback/$id/progress"
        val body = json("position" to seconds)
        if (completed) body.put("watched", true)
        lastSave = System.currentTimeMillis()
        activity.worker.execute { runCatching { api.request(path, "PUT", body) } }
    }

    private fun seek(delta: Long) {
        val player = media ?: return
        val upper = if (duration > 0) (duration - 1000).coerceAtLeast(0) else Long.MAX_VALUE
        val target = (position() + delta).coerceIn(0, upper)
        ended = false
        if (direct) player.seekTo(target) else open(id, target)
    }

    private fun next() {
        if (next <= 0) return
        save()
        open(next, 0L, advance = true)
    }

    private fun toggle() {
        val player = media ?: return
        if (ended) open(id, 0L, autoplay = true)
        else if (player.playWhenReady) pause() else player.play()
    }

    private fun audioMenu() {
        val tracks = info.rows("audio_tracks")
        val names = tracks.map {
            it.optString("title").ifBlank { it.optString("language").ifBlank { "音轨 ${it.getInt("index")}" } }
        }
        controls.pick("音轨", names) {
            audio = tracks[it].getInt("index")
            compatible = true
            open(id, position())
        }
    }

    private fun subtitleMenu() {
        val tracks = info.rows("subtitles")
        controls.pick("字幕", listOf("关闭字幕") + tracks.map { it.getString("label") }) {
            autoSubtitle = false
            subtitle = tracks.getOrNull(it - 1)?.getInt("id")
            open(id, position())
        }
    }

    private fun options() {
        controls.pick("播放设置", listOf("播放方式", "分辨率", "倍速")) { option ->
            when (option) {
                0 -> controls.pick("播放方式", listOf("原片", "兼容播放")) {
                    compatible = it == 1
                    open(id, position())
                }
                1 -> controls.pick("分辨率", listOf("自动", "原分辨率", "1080p", "720p", "480p")) {
                    quality = listOf("auto", "original", "1080", "720", "480")[it]
                    open(id, position())
                }
                2 -> controls.pick("倍速", listOf("0.5×", "0.75×", "1×", "1.25×", "1.5×", "2×")) {
                    speed = listOf(.5f, .75f, 1f, 1.25f, 1.5f, 2f)[it]
                    media?.setPlaybackSpeed(speed)
                }
            }
        }
    }

    fun key(event: KeyEvent) = controls.key(event, ::toggle, ::seek, ::next)

    fun pause() {
        media?.pause()
        save()
    }

    private fun detach() {
        activity.main.removeCallbacks(tick)
        val player = media
        media = null
        controls.view.player = null
        player?.release()
    }
}
