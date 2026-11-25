package com.monkeydog.dinosaur.mobile

import com.monkeydog.dinosaur.android.Api
import com.monkeydog.dinosaur.android.json
import com.monkeydog.dinosaur.android.rows
import android.graphics.Color
import android.os.Handler
import android.os.Looper
import android.view.Gravity
import android.widget.*
import androidx.media3.common.MediaItem
import androidx.media3.common.Player
import androidx.media3.datasource.okhttp.OkHttpDataSource
import androidx.media3.exoplayer.ExoPlayer
import androidx.media3.exoplayer.source.DefaultMediaSourceFactory
import androidx.media3.ui.PlayerView
import org.json.JSONObject

@androidx.media3.common.util.UnstableApi
internal class PlaybackScreen(private val activity: MainActivity, private val api: Api, private val close: () -> Unit) {
    private val handler = Handler(Looper.getMainLooper())
    private var player: ExoPlayer? = null
    private var id = 0L
    private var next = 0L
    private var mode = "original"
    private var quality = "auto"
    private var audio: Int? = null
    private var subtitle: Int? = null
    private var autoSubtitle = true
    private var offset = 0L
    private var duration = 0L
    private var completed = false
    private var fullscreen = false
    private var rate = 1f
    private var revision = 0
    private var info = JSONObject()
    private var lastSave = 0L
    private var chrome: PlaybackChrome? = null
    private lateinit var caption: TextView
    private val captions = PlaybackCaptions(activity, api)
    private val tick = object : Runnable {
        override fun run() {
            val current = player ?: return
            val position = position()
            caption.updateText(captions.at(id, subtitle, info, position))
            chrome?.tick(position, duration)
            if (current.isPlaying && System.currentTimeMillis() - lastSave >= 10000) save()
            handler.postDelayed(this, 500)
        }
    }

    fun open(mediaId: Long, start: Long? = null, nextEpisode: Boolean = false): Unit = with(activity) {
        val load = ++revision
        detach()
        captions.reset()
        id = mediaId
        next = 0
        if (nextEpisode) {
            audio = null
            subtitle = null
            autoSubtitle = true
        }
        completed = false
        val parameters = buildString {
            append("mode=$mode&quality=$quality&start=${(start ?: 0) / 1000.0}")
            audio?.let { append("&audio=$it") }
            if (autoSubtitle) append("&subtitle=auto")
            else info.rows("subtitles").firstOrNull { it.optInt("id") == subtitle && it.optString("delivery") == "burn" }
                ?.let { append("&subtitle=${it.getInt("id")}") }
        }
        render()
        task({ api.request("/api/playback/$mediaId?$parameters") }, {
            if (load == revision) chrome?.fail(it.message ?: "无法播放")
        }) { result ->
            if (load == revision) prepare(result, start)
        }
    }

    private fun prepare(result: JSONObject, requested: Long?) {
        info = result
        val plan = result.getJSONObject("plan")
        val item = result.getJSONObject("item")
        val start = requested ?: (result.optDouble("resume_position", 0.0) * 1000).toLong()
        val direct = plan.getString("mode") == "direct"
        offset = if (direct) 0 else start
        duration = (item.optDouble("duration", 0.0) * 1000).toLong().coerceIn(0, Int.MAX_VALUE.toLong())
        next = result.optJSONObject("next_episode")?.optLong("id") ?: 0L
        if (autoSubtitle) {
            subtitle = if (result.isNull("selected_subtitle")) null else result.getInt("selected_subtitle")
            autoSubtitle = false
        }
        val source = MediaItem.Builder().setUri(api.stream(result.getString("stream_url"), offset))
        val http = OkHttpDataSource.Factory(Api.http).setDefaultRequestProperties(api.headers)
        val current = ExoPlayer.Builder(activity).setMediaSourceFactory(DefaultMediaSourceFactory(http)).build()
        player = current
        render()
        current.addListener(object : Player.Listener {
            override fun onIsPlayingChanged(isPlaying: Boolean) { chrome?.setPlaying(isPlaying) }

            override fun onPlaybackStateChanged(state: Int) {
                if (state != Player.STATE_ENDED || completed) return
                val observed = maxOf(current.currentPosition, current.duration.coerceAtLeast(0)) + offset
                if (duration > 0 && observed < duration - maxOf(5000L, duration / 100)) {
                    recover()
                    return
                }
                completed = true
                save(true)
                if (next > 0) open(next, 0L, true)
            }

            override fun onPlayerError(error: androidx.media3.common.PlaybackException) { recover() }
        })
        current.setMediaItem(source.build(), if (direct) start else 0)
        current.setPlaybackSpeed(rate)
        current.prepare()
        current.play()
        handler.post(tick)
    }

    private fun recover() {
        if (mode == "original") {
            mode = "compatible"
            open(id, position())
        } else chrome?.fail("播放失败")
    }

    private fun render(): Unit = with(activity) {
        generation++
        back = { if (fullscreen) rotate() else exit() }
        chrome?.close()
        val root = playbackRoot { fullscreen }
        root.addView((layoutInflater.inflate(R.layout.video_surface, root, false) as PlayerView).apply {
            useController = false
            player = this@PlaybackScreen.player
            contentDescription = "显示播放控制"
            setOnClickListener { chrome?.tap() }
        }, FrameLayout.LayoutParams(-1, -1))
        caption = TextView(this).apply { gravity = Gravity.CENTER; textSize = 18f; setTextColor(Color.WHITE) }
        root.addView(caption, FrameLayout.LayoutParams(-1, -2, Gravity.BOTTOM).apply { bottomMargin = dp(100) })
        val item = info.optJSONObject("item")
        val title = listOf(item?.optString("series_title"), item?.optString("title"))
            .filterNotNull().filter(String::isNotBlank).joinToString(" · ")
        val advance = if (next > 0) ({ save(); open(next, 0L, true) }) else null
        chrome = PlaybackChrome(this, root, title, if (player == null) 0 else duration, advance,
            { if (fullscreen) rotate() else exit() }, ::toggle, ::rotate, ::seek, settings()).apply {
            setFullscreen(fullscreen)
        }
        setContentView(root)
    }

    private fun seek(position: Long) {
        if (duration <= 0 || player == null) return
        val target = position.coerceIn(0, duration)
        if (info.optJSONObject("plan")?.optString("mode") == "direct") player?.seekTo(target)
        else open(id, target)
    }

    private fun toggle() {
        val current = player ?: return
        if (completed) open(id, 0L)
        else if (current.isPlaying) pause()
        else current.play()
    }

    private fun settings(): List<PlayerSetting> {
        val tracks = info.rows("audio_tracks")
        val subtitles = info.rows("subtitles")
        val available = player != null
        return listOf(
            PlayerSetting(PlayerIcon.AUDIO, "音轨", available && tracks.isNotEmpty()) {
                val selected = audio ?: info.optInt("selected_audio", -1).takeIf { it >= 0 }
                    ?: tracks.firstOrNull()?.optInt("index")
                tracks.map { track ->
                    val index = track.getInt("index")
                    val name = track.optString("title").ifBlank { track.optString("language").ifBlank { "音轨 $index" } }
                    PlayerOption(name, index == selected) { audio = index; mode = "compatible"; open(id, position()) }
                }
            },
            PlayerSetting(PlayerIcon.SUBTITLE, "字幕", available && subtitles.isNotEmpty()) {
                listOf(PlayerOption("关闭字幕", subtitle == null) { changeSubtitle(null) }) +
                    subtitles.map { track ->
                        val key = track.getInt("id")
                        PlayerOption(track.optString("label"), subtitle == key) { changeSubtitle(key) }
                    }
            },
            PlayerSetting(PlayerIcon.MODE, "播放方式", available) {
                listOf(
                    PlayerOption("自动选择", mode == "original") { mode = "original"; open(id, position()) },
                    PlayerOption("兼容模式", mode == "compatible") { mode = "compatible"; open(id, position()) },
                )
            },
            PlayerSetting(PlayerIcon.QUALITY, "分辨率", available) {
                listOf("auto" to "自动", "original" to "原分辨率", "1080" to "1080p", "720" to "720p", "480" to "480p")
                    .map { (value, label) ->
                        PlayerOption(label, quality == value) { quality = value; open(id, position()) }
                    }
            },
            PlayerSetting(PlayerIcon.RATE, "倍速", available) {
                listOf(0.5f, 0.75f, 1f, 1.25f, 1.5f, 2f).map { value ->
                    PlayerOption("${value}×", rate == value) { rate = value; player?.setPlaybackSpeed(value) }
                }
            },
        )
    }

    private fun changeSubtitle(track: Int?) {
        subtitle = track
        autoSubtitle = false
        open(id, position())
    }

    private fun rotate() {
        fullscreen = !fullscreen
        activity.setPlaybackFullscreen(fullscreen)
        chrome?.setFullscreen(fullscreen)
    }

    private fun position() = offset + (player?.currentPosition ?: 0)
    private fun save(watched: Boolean = false) {
        if (player == null || completed && !watched) return
        lastSave = System.currentTimeMillis()
        val path = "/api/playback/$id/progress"
        val data = json("position" to (if (watched && duration > 0) duration else position()) / 1000.0)
        if (watched) data.put("watched", true)
        activity.worker.execute { runCatching { api.request(path, "PUT", data) } }
    }
    fun pause() { player?.pause(); save() }
    private fun detach() { handler.removeCallbacks(tick); player?.release(); player = null }
    fun release() { revision++; pause(); detach(); chrome?.close(); if (fullscreen) rotate() }
    private fun exit() { release(); activity.playback = null; close() }
}
