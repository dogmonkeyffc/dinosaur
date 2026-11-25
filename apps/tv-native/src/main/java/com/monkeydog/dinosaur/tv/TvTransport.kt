package com.monkeydog.dinosaur.tv

import android.app.AlertDialog
import android.content.res.ColorStateList
import android.graphics.Color
import android.view.Gravity
import android.view.KeyEvent
import android.view.View
import android.widget.Button
import android.widget.FrameLayout
import android.widget.ProgressBar
import androidx.media3.ui.PlayerView

@androidx.media3.common.util.UnstableApi
internal class TvTransport(private val activity: TvActivity) {
    val view = PlayerView(activity).apply {
        useController = false
        isFocusable = true
        isFocusableInTouchMode = true
        contentDescription = "播放画面"
    }
    private val heading = activity.label("", 26f, bold = true)
    private val subtitle = activity.label("", 23f).apply {
        gravity = Gravity.CENTER
        setShadowLayer(4f, 1f, 2f, Color.BLACK)
    }
    private val position = activity.label("", 15f, muted)
    private val bar = ProgressBar(activity, null, android.R.attr.progressBarStyleHorizontal).apply {
        max = 1000
        progressTintList = ColorStateList.valueOf(accent)
        importantForAccessibility = View.IMPORTANT_FOR_ACCESSIBILITY_NO
    }
    private val actions = activity.row()
    private val overlay = activity.column().apply {
        setBackgroundColor(0xf012191b.toInt())
        setPadding(activity.dp(38), activity.dp(14), activity.dp(38), activity.dp(22))
        addView(heading)
        addView(position)
        addView(bar)
        addView(actions)
    }
    val root = FrameLayout(activity).apply {
        setBackgroundColor(Color.BLACK)
        addView(view, FrameLayout.LayoutParams(-1, -1))
        addView(subtitle, FrameLayout.LayoutParams(-1, -2, Gravity.BOTTOM).apply {
            leftMargin = activity.dp(36)
            rightMargin = activity.dp(36)
        })
        addView(this@TvTransport.overlay, FrameLayout.LayoutParams(-1, -2, Gravity.BOTTOM))
    }
    private var playButton: Button? = null
    private var isPlaying = false
    private var isReady = false
    private var dialog: AlertDialog? = null
    private val hide = Runnable {
        if (isPlaying && dialog == null) {
            view.requestFocus()
            overlay.visibility = View.GONE
            subtitleMargin(false)
        }
    }

    fun loading() {
        isReady = false
        isPlaying = false
        playButton = null
        heading.text = "正在加载"
        position.text = ""
        bar.progress = 0
        subtitle.text = ""
        actions.removeAllViews()
        val back = activity.control("返回", activity::back)
        actions.addControl(back, horizontal = true)
        reveal(false)
        back.requestFocus()
    }

    fun ready(
        title: String,
        hasNext: Boolean,
        toggle: () -> Unit,
        seek: (Long) -> Unit,
        next: () -> Unit,
        audio: () -> Unit,
        subtitles: () -> Unit,
        settings: () -> Unit,
    ) {
        isReady = true
        heading.text = title
        actions.removeAllViews()
        val play = button("暂停", toggle)
        playButton = play
        button("后退 15 秒") { seek(-15000) }
        button("前进 15 秒") { seek(15000) }
        if (hasNext) button("下一集", next)
        button("音轨", audio)
        button("字幕", subtitles)
        button("播放设置", settings)
        reveal()
        play.requestFocus()
    }

    fun failure(message: String, retry: () -> Unit) {
        loading()
        heading.text = message
        button("重试", retry).requestFocus()
    }

    private fun button(title: String, action: () -> Unit) = activity.control(title) {
        postpone()
        action()
    }.also { actions.addControl(it, horizontal = true) }

    fun update(time: Long, duration: Long) {
        val value = "${clock(time)} / ${clock(duration)}"
        if (position.text != value) position.text = value
        if (duration > 0) bar.progress = (1000.0 * time / duration).toInt().coerceIn(0, 1000)
    }

    private fun clock(time: Long) = "%02d:%02d".format(time / 60000, (time / 1000) % 60)

    fun caption(text: String) {
        if (subtitle.text != text) subtitle.text = text
    }

    fun playing(value: Boolean, ended: Boolean) {
        isPlaying = value
        playButton?.text = if (ended) "重播" else if (value) "暂停" else "播放"
        if (value) postpone() else reveal()
    }

    fun reveal(focus: Boolean = true) {
        val wasHidden = overlay.visibility != View.VISIBLE
        overlay.visibility = View.VISIBLE
        subtitleMargin(true)
        if (wasHidden && focus) playButton?.requestFocus()
        postpone()
    }

    private fun subtitleMargin(raised: Boolean) {
        val params = subtitle.layoutParams as FrameLayout.LayoutParams
        val margin = activity.dp(if (raised) 162 else 42)
        if (params.bottomMargin != margin) {
            params.bottomMargin = margin
            subtitle.layoutParams = params
        }
    }

    private fun postpone() {
        activity.main.removeCallbacks(hide)
        if (isPlaying && dialog == null) activity.main.postDelayed(hide, 6000)
    }

    fun pick(title: String, choices: List<String>, selected: (Int) -> Unit) {
        val next = AlertDialog.Builder(activity).setTitle(title)
        if (choices.isEmpty()) next.setMessage("暂无可选项目").setPositiveButton("确定", null)
        else next.setItems(choices.toTypedArray()) { _, index -> selected(index) }
        val menu = next.create()
        dialog = menu
        menu.setOnDismissListener {
            if (dialog === menu) dialog = null
            postpone()
        }
        menu.show()
        postpone()
    }

    fun key(event: KeyEvent, toggle: () -> Unit, seek: (Long) -> Unit, next: () -> Unit): Boolean {
        if (dialog != null || !isReady) return false
        val code = event.keyCode
        val handled = when (code) {
            KeyEvent.KEYCODE_MEDIA_PLAY_PAUSE -> toggle
            KeyEvent.KEYCODE_MEDIA_NEXT -> next
            KeyEvent.KEYCODE_MEDIA_FAST_FORWARD -> ({ seek(15000) })
            KeyEvent.KEYCODE_MEDIA_REWIND -> ({ seek(-15000) })
            else -> null
        }
        if (handled != null) {
            if (event.action == KeyEvent.ACTION_DOWN && event.repeatCount == 0) {
                reveal()
                handled()
            }
            return true
        }
        val directional = code in listOf(KeyEvent.KEYCODE_DPAD_UP, KeyEvent.KEYCODE_DPAD_DOWN, KeyEvent.KEYCODE_DPAD_LEFT,
            KeyEvent.KEYCODE_DPAD_RIGHT, KeyEvent.KEYCODE_DPAD_CENTER, KeyEvent.KEYCODE_ENTER)
        if (!directional) return false
        if (overlay.visibility == View.GONE && event.action == KeyEvent.ACTION_DOWN) {
            reveal()
            return true
        }
        if (event.action == KeyEvent.ACTION_DOWN) postpone()
        return false
    }

    fun close() {
        dialog?.dismiss()
        activity.main.removeCallbacks(hide)
    }
}
