package com.monkeydog.dinosaur.mobile

import android.graphics.Color
import android.graphics.drawable.GradientDrawable
import android.widget.ImageView
import android.widget.TextView

internal const val PLAYER_TEXT = 0xfff2f6fc.toInt()

internal enum class PlayerIcon(val drawable: Int) {
    BACK(R.drawable.player_back),
    PLAY(R.drawable.player_play),
    PAUSE(R.drawable.player_pause),
    NEXT(R.drawable.player_next),
    BACKWARD(R.drawable.player_backward),
    FORWARD(R.drawable.player_forward),
    AUDIO(R.drawable.player_audio),
    SUBTITLE(R.drawable.player_subtitle),
    MODE(R.drawable.player_mode),
    QUALITY(R.drawable.player_quality),
    RATE(R.drawable.player_rate),
    FULLSCREEN(R.drawable.player_fullscreen),
    COLLAPSE(R.drawable.player_collapse),
}

internal class PlayerAction(activity: MainActivity, icon: PlayerIcon, label: String, action: () -> Unit) :
    ImageView(activity) {
    private var icon = icon

    init {
        scaleType = ScaleType.CENTER
        setImageResource(icon.drawable)
        contentDescription = label
        onTap(action)
    }

    fun update(next: PlayerIcon, label: String) {
        if (icon == next) return
        icon = next
        contentDescription = label
        setImageResource(next.drawable)
    }
}

internal fun playerShade(top: Boolean) = GradientDrawable(
    GradientDrawable.Orientation.TOP_BOTTOM,
    if (top) intArrayOf(0xbb000000.toInt(), Color.TRANSPARENT)
    else intArrayOf(Color.TRANSPARENT, 0xbb000000.toInt(), 0xee000000.toInt()),
)

internal fun clock(ms: Long): String {
    val seconds = ms.coerceAtLeast(0) / 1000
    val minutes = (seconds / 60) % 60
    val hours = seconds / 3600
    return if (hours > 0) "%d:%02d:%02d".format(hours, minutes, seconds % 60)
    else "%02d:%02d".format(minutes, seconds % 60)
}

internal fun TextView.updateText(value: String) {
    if (text.toString() != value) text = value
}
