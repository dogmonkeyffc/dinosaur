package com.monkeydog.dinosaur.mobile

import android.graphics.Color
import android.graphics.drawable.GradientDrawable
import android.view.Gravity
import android.view.View
import android.widget.LinearLayout
import android.widget.TextView

internal fun MainActivity.playerControl(icon: String, label: String, action: () -> Unit) = TextView(this).apply {
    text = icon
    textSize = 18f
    setTextColor(Color.WHITE)
    gravity = Gravity.CENTER
    contentDescription = label
    minWidth = dp(40)
    minHeight = dp(44)
    isFocusable = true
    setOnClickListener { action() }
}

internal fun MainActivity.playbackControls(
    full: Boolean, hasNext: Boolean, play: () -> Unit, next: () -> Unit, rotate: () -> Unit,
    audio: () -> Unit, subtitle: () -> Unit, mode: () -> Unit, quality: () -> Unit, rate: () -> Unit,
) = LinearLayout(this).apply {
    addView(playerControl("▶", "播放或暂停", play))
    if (hasNext) addView(playerControl("▸|", "下一集", next))
    addView(View(this@playbackControls), LinearLayout.LayoutParams(0, 1, 1f))
    addView(playerControl("♫", "音轨", audio))
    addView(playerControl("CC", "字幕", subtitle))
    addView(playerControl("≡", "播放方式", mode))
    addView(playerControl("HD", "分辨率", quality))
    addView(playerControl("×", "倍速", rate))
    addView(playerControl("⛶", if (full) "退出全屏" else "全屏", rotate).apply { tag = "fullscreen" })
}

internal fun playerShade(top: Boolean) = GradientDrawable(
    GradientDrawable.Orientation.TOP_BOTTOM,
    if (top) intArrayOf(0xbb000000.toInt(), Color.TRANSPARENT)
    else intArrayOf(Color.TRANSPARENT, 0xee000000.toInt()),
)

internal fun clock(ms: Long) = "%02d:%02d".format(ms / 60000, (ms / 1000) % 60)

internal fun TextView.updateText(value: String) {
    if (text.toString() != value) text = value
}
