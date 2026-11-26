package com.monkeydog.dinosaur.mobile

import android.graphics.Color
import android.graphics.drawable.ClipDrawable
import android.graphics.drawable.ColorDrawable
import android.graphics.drawable.LayerDrawable
import android.text.TextUtils
import android.view.Gravity
import android.view.View
import android.widget.FrameLayout
import android.widget.LinearLayout
import android.widget.PopupWindow
import android.widget.ScrollView
import android.widget.SeekBar
import android.widget.TextView

internal class PlaybackChrome(
    private val activity: MainActivity, private val root: FrameLayout, title: String, duration: Long,
    next: (() -> Unit)?, goBack: () -> Unit, toggle: () -> Unit, rotate: () -> Unit, seek: (Long) -> Unit,
    settings: List<PlayerSetting>, nextLabel: String,
) {
    private val header = activity.phoneRow()
    private val controls = activity.column()
    private val status = text("正在准备", 10)
    private val error = text("", 13)
    private val play = PlayerAction(activity, PlayerIcon.PLAY, "播放", toggle)
    private val full = PlayerAction(activity, PlayerIcon.FULLSCREEN, "全屏", rotate)
    private val track = SeekBar(activity)
    private val hide = Runnable { show(false) }
    private var playing = false
    private var fullscreen = false
    private var seeking = false
    private var popup: PopupWindow? = null

    init {
        with(activity) {
            header.apply {
                background = playerShade(true)
                setPadding(dp(10), dp(12), dp(10), dp(30))
                addView(PlayerAction(activity, PlayerIcon.BACK, "返回", goBack), controlSize())
                addView(text(title, 12).apply {
                    maxLines = 1
                    ellipsize = TextUtils.TruncateAt.END
                }, LinearLayout.LayoutParams(0, -2, 1f).apply { leftMargin = dp(10) })
            }
            root.addView(header, FrameLayout.LayoutParams(-1, -2, Gravity.TOP))
            controls.apply {
                background = playerShade(false)
                setPadding(dp(10), dp(25), dp(10), dp(16))
                addView(status)
                val drawables = LayerDrawable(arrayOf(ColorDrawable(0x40ffffff),
                    ClipDrawable(ColorDrawable(0xffc2df9e.toInt()), Gravity.LEFT, ClipDrawable.HORIZONTAL)))
                drawables.setId(0, android.R.id.background)
                drawables.setId(1, android.R.id.progress)
                track.apply {
                    max = duration.toInt().coerceAtLeast(1)
                    isEnabled = duration > 0
                    thumb = null
                    splitTrack = false
                    progressDrawable = drawables
                    minimumHeight = 0
                    setPadding(0, dp(11), 0, dp(10))
                    contentDescription = "播放进度"
                    setOnSeekBarChangeListener(object : SeekBar.OnSeekBarChangeListener {
                        override fun onProgressChanged(bar: SeekBar, value: Int, user: Boolean) {
                            if (user) time(value.toLong(), duration)
                        }
                        override fun onStartTrackingTouch(bar: SeekBar) { seeking = true; schedule() }
                        override fun onStopTrackingTouch(bar: SeekBar) {
                            seeking = false
                            seek(bar.progress.toLong())
                            schedule()
                        }
                    })
                }
                addView(track, LinearLayout.LayoutParams(-1, dp(24)))
                addView(phoneRow().apply {
                    addView(play, controlSize())
                    val rewind = PlayerAction(activity, PlayerIcon.BACKWARD, "后退 10 秒") { seek(track.progress - 10000L) }
                    val forward = PlayerAction(activity, PlayerIcon.FORWARD, "前进 10 秒") { seek(track.progress + 10000L) }
                    addView(rewind, controlSize())
                    addView(forward, controlSize())
                    listOf(rewind, forward).forEach { it.visibility = View.GONE }
                    addOnLayoutChangeListener { _, left, _, right, _, _, _, _, _ ->
                        val visible = if (right - left > dp(700)) View.VISIBLE else View.GONE
                        listOf(rewind, forward).forEach { if (it.visibility != visible) it.visibility = visible }
                    }
                    next?.let { addView(PlayerAction(activity, PlayerIcon.NEXT, nextLabel, it), controlSize()) }
                    addView(View(activity), LinearLayout.LayoutParams(0, 1, 1f))
                    settings.forEach { item ->
                        addView(PlayerAction(activity, item.icon, item.label) { open(item.options()) }.apply {
                            isEnabled = item.enabled
                            if (!isEnabled) alpha = .35f
                        }, controlSize())
                    }
                    addView(full, controlSize())
                })
            }
            root.addView(controls, FrameLayout.LayoutParams(-1, -2, Gravity.BOTTOM))
            error.visibility = View.GONE
            error.setTextColor(0xffffb4ab.toInt())
            error.setBackgroundColor(0xff18212b.toInt())
            error.setPadding(dp(12), dp(12), dp(12), dp(12))
            root.addView(error, FrameLayout.LayoutParams(-1, -2, Gravity.TOP).apply {
                setMargins(dp(16), dp(80), dp(16), 0)
            })
        }
    }

    private fun MainActivity.controlSize() = LinearLayout.LayoutParams(minOf(dp(40), resources.displayMetrics.widthPixels / 8), dp(40))

    private fun text(value: String, size: Int) = TextView(activity).apply {
        text = value
        textSize = size.toFloat()
        setTextColor(PLAYER_TEXT)
        includeFontPadding = false
    }

    fun tick(position: Long, duration: Long) {
        if (!seeking) {
            track.progress = position.coerceIn(0, duration).toInt()
            time(position, duration)
        }
    }

    private fun time(position: Long, duration: Long) {
        status.updateText("${clock(position)} / ${clock(duration)}")
    }

    fun setPlaying(value: Boolean) {
        if (playing == value) return
        playing = value
        play.update(if (value) PlayerIcon.PAUSE else PlayerIcon.PLAY, if (value) "暂停" else "播放")
        if (!value) show(true) else schedule()
    }

    fun setFullscreen(value: Boolean) {
        fullscreen = value
        full.update(if (value) PlayerIcon.COLLAPSE else PlayerIcon.FULLSCREEN, if (value) "退出全屏" else "全屏")
        show(true)
    }

    fun tap() {
        if (fullscreen && playing) show(controls.visibility != View.VISIBLE)
    }

    fun fail(message: String) {
        error.text = message
        error.visibility = View.VISIBLE
        show(true)
    }

    private fun show(visible: Boolean) {
        val value = if (visible) View.VISIBLE else View.GONE
        header.visibility = value
        controls.visibility = value
        schedule()
    }

    private fun schedule() {
        root.removeCallbacks(hide)
        if (fullscreen && playing && !seeking && popup == null &&
            error.visibility != View.VISIBLE && controls.visibility == View.VISIBLE) root.postDelayed(hide, 5000)
    }

    private fun open(options: List<PlayerOption>) = with(activity) {
        if (options.isEmpty()) return@with
        val list = column()
        val scroll = ScrollView(this).apply {
            background = panel(0xff18212b.toInt(), 8, 0x38ffffff)
            clipToOutline = true
            addView(list)
        }
        val window = PopupWindow(scroll, dp(220), minOf(dp(210), dp(42 * options.size)), true).apply {
            isOutsideTouchable = true
            setBackgroundDrawable(ColorDrawable(Color.TRANSPARENT))
            elevation = dp(5).toFloat()
            setOnDismissListener { popup = null; schedule() }
        }
        options.forEach { option ->
            list.addView(phoneRow().apply {
                minimumHeight = dp(42)
                setPadding(dp(16), dp(8), dp(16), dp(8))
                isSelected = option.selected
                val color = if (option.selected) 0xffb9d7ff.toInt() else PLAYER_TEXT
                if (option.selected) setBackgroundColor(0x24b9d7ff)
                addView(text(option.name, 13).apply { setTextColor(color) }, LinearLayout.LayoutParams(0, -2, 1f))
                if (option.selected) addView(text("✓", 13).apply { setTextColor(color) })
                onTap {
                    window.dismiss()
                    if (!option.selected) option.apply()
                }
            })
        }
        popup = window
        schedule()
        window.showAtLocation(root, Gravity.BOTTOM or Gravity.RIGHT, dp(14), dp(90))
    }

    fun close() {
        popup?.dismiss()
        root.removeCallbacks(hide)
    }
}

internal data class PlayerOption(val name: String, val selected: Boolean, val apply: () -> Unit)
internal data class PlayerSetting(val icon: PlayerIcon, val label: String, val enabled: Boolean = true, val options: () -> List<PlayerOption>)
