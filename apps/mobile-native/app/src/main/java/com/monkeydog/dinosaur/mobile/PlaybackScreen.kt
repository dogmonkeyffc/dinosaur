package com.monkeydog.dinosaur.mobile

import com.monkeydog.dinosaur.android.Api
import com.monkeydog.dinosaur.android.json
import com.monkeydog.dinosaur.android.rows
import android.content.pm.ActivityInfo
import android.graphics.Color
import android.os.Handler
import android.os.Looper
import android.view.Gravity
import android.view.View
import android.view.WindowInsets
import android.view.WindowInsetsController
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
    private lateinit var status: TextView
    private lateinit var transport: LinearLayout
    private lateinit var header: LinearLayout
    private lateinit var scrubber: SeekBar
    private lateinit var caption: TextView
    private val captions = PlaybackCaptions(activity, api)
    private val tick = object : Runnable {
        override fun run() {
            val current = player ?: return
            val position = position()
            caption.updateText(captions.at(id, subtitle, info, position))
            if (!scrubber.isPressed) scrubber.progress = position.coerceAtMost(duration).toInt()
            status.updateText("${clock(position)} / ${clock(duration)}")
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
            if (load == revision) status.text = it.message ?: "无法播放"
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
        } else status.text = "播放失败"
    }

    private fun render(): Unit = with(activity) {
        generation++
        back = { if (fullscreen) rotate() else exit() }
        val root = FrameLayout(this).apply { setBackgroundColor(Color.BLACK); fitsSystemWindows = true }
        root.addView((layoutInflater.inflate(R.layout.video_surface, root, false) as PlayerView).apply {
            useController = false
            player = this@PlaybackScreen.player
            setOnClickListener {
                val show = if (transport.visibility == View.VISIBLE) View.GONE else View.VISIBLE
                transport.visibility = show
                header.visibility = show
            }
        }, FrameLayout.LayoutParams(-1, -1))
        header = LinearLayout(this).apply {
            background = playerShade(true)
            addView(playerControl("←", "返回") { exit() })
            addView(TextView(context).apply { text = info.optJSONObject("item")?.optString("title").orEmpty(); textSize = 16f })
        }
        root.addView(header, FrameLayout.LayoutParams(-1, -2, Gravity.TOP))
        caption = TextView(this).apply { gravity = Gravity.CENTER; textSize = 18f; setTextColor(Color.WHITE) }
        root.addView(caption, FrameLayout.LayoutParams(-1, -2, Gravity.BOTTOM).apply { bottomMargin = dp(100) })
        transport = LinearLayout(this).apply { orientation = LinearLayout.VERTICAL; background = playerShade(false) }
        status = TextView(this).apply { text = "正在准备"; setPadding(dp(12), 0, 0, 0) }
        transport.addView(status)
        scrubber = SeekBar(this).apply {
            max = duration.toInt().coerceAtLeast(1)
            contentDescription = "播放进度"
            progressTintList = android.content.res.ColorStateList.valueOf(0xffc2df9e.toInt())
            setOnSeekBarChangeListener(object : SeekBar.OnSeekBarChangeListener {
                override fun onProgressChanged(bar: SeekBar, value: Int, user: Boolean) = Unit
                override fun onStartTrackingTouch(bar: SeekBar) = Unit
                override fun onStopTrackingTouch(bar: SeekBar) {
                    val target = bar.progress.toLong()
                    if (info.optJSONObject("plan")?.optString("mode") == "direct") player?.seekTo(target)
                    else open(this@PlaybackScreen.id, target)
                }
            })
        }
        transport.addView(scrubber)
        transport.addView(playbackControls(
            fullscreen, next > 0, ::toggle, { save(); open(next, 0L, true) }, ::rotate,
            { audioMenu() }, { subtitleMenu() }, { modeMenu() }, { qualityMenu() }, { rateMenu() }
        ))
        root.addView(transport, FrameLayout.LayoutParams(-1, -2, Gravity.BOTTOM))
        setContentView(root)
    }

    private fun toggle() {
        val current = player ?: return
        if (completed) open(id, 0L)
        else if (current.isPlaying) pause()
        else current.play()
    }

    private fun audioMenu() = with(activity) {
        val tracks = info.rows("audio_tracks")
        choose("音轨", tracks.map { it.optString("title").ifBlank { it.optString("language").ifBlank { "音轨 ${it.getInt("index")}" } } }) {
            audio = tracks[it].getInt("index")
            mode = "compatible"
            open(id, position())
        }
    }

    private fun subtitleMenu() = with(activity) {
        val tracks = info.rows("subtitles")
        choose("字幕", listOf("关闭字幕") + tracks.map { it.optString("label") }) {
            subtitle = tracks.getOrNull(it - 1)?.getInt("id")
            autoSubtitle = false
            open(id, position())
        }
    }

    private fun modeMenu() = activity.choose("播放方式", listOf("原片", "兼容播放")) {
        mode = if (it == 0) "original" else "compatible"
        open(id, position())
    }

    private fun qualityMenu() = activity.choose("分辨率", listOf("自动","原分辨率","1080p","720p","480p")) {
        quality = listOf("auto","original","1080","720","480")[it]
        open(id, position())
    }

    private fun rateMenu() {
        val values = listOf(0.5f, 0.75f, 1f, 1.25f, 1.5f, 2f)
        activity.choose("倍速", values.map { "${it}×" }) { rate = values[it]; player?.setPlaybackSpeed(rate) }
    }

    private fun rotate() {
        fullscreen = !fullscreen
        activity.requestedOrientation = if (fullscreen) ActivityInfo.SCREEN_ORIENTATION_SENSOR_LANDSCAPE else ActivityInfo.SCREEN_ORIENTATION_UNSPECIFIED
        if (android.os.Build.VERSION.SDK_INT >= 30) {
            activity.window.insetsController?.apply {
                systemBarsBehavior = WindowInsetsController.BEHAVIOR_SHOW_TRANSIENT_BARS_BY_SWIPE
                if (fullscreen) hide(WindowInsets.Type.systemBars()) else show(WindowInsets.Type.systemBars())
            }
        }
        transport.findViewWithTag<TextView>("fullscreen")?.contentDescription = if (fullscreen) "退出全屏" else "全屏"
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
    fun release() { revision++; pause(); detach(); if (fullscreen) rotate() }
    private fun exit() { release(); activity.playback = null; close() }
}
