package com.monkeydog.dinosaur.mobile

import android.content.res.ColorStateList
import android.graphics.Typeface
import android.graphics.drawable.GradientDrawable
import android.graphics.drawable.RippleDrawable
import android.view.Gravity
import android.view.View
import android.widget.FrameLayout
import android.widget.LinearLayout
import android.widget.TextView
import kotlin.math.roundToInt

internal const val PHONE_BG = 0xff111113.toInt()
internal const val PHONE_SURFACE = 0xff28292c.toInt()
internal const val PHONE_TEXT = 0xffe5e5e9.toInt()
internal const val PHONE_MUTED = 0xffb5b6be.toInt()
internal const val PHONE_ACCENT = 0xffa8c7fa.toInt()
internal const val PHONE_BORDER = 0xff323337.toInt()
internal const val PHONE_SELECTED = 0xff284064.toInt()

internal fun MainActivity.ui(value: Number) = (value.toFloat() * .85f * resources.displayMetrics.density).roundToInt()

internal fun MainActivity.panel(color: Int, radius: Int, stroke: Int? = null) = GradientDrawable().apply {
    setColor(color)
    cornerRadius = ui(radius).toFloat()
    stroke?.let { setStroke(ui(1).coerceAtLeast(1), it) }
}

internal fun MainActivity.column() = LinearLayout(this).apply { orientation = LinearLayout.VERTICAL }

internal fun MainActivity.phoneRow() = LinearLayout(this).apply { gravity = Gravity.CENTER_VERTICAL }

internal fun MainActivity.phoneText(value: String, size: Int, color: Int = PHONE_TEXT, bold: Boolean = false) =
    TextView(this).apply {
        text = value
        textSize = size * .85f
        includeFontPadding = false
        setTextColor(color)
        if (bold) typeface = Typeface.create("sans-serif-medium", Typeface.NORMAL)
    }

internal fun MainActivity.phoneAction(
    title: String, action: () -> Unit, size: Int = 16, color: Int = PHONE_ACCENT, fill: Int = PHONE_SURFACE,
) = phoneText(title, size, color, true).apply {
    gravity = Gravity.CENTER
    minHeight = ui(48)
    setPadding(ui(16), ui(8), ui(16), ui(8))
    background = panel(fill, 25)
    onTap(action)
}

internal fun View.onTap(action: () -> Unit) {
    isFocusable = true
    setOnClickListener { action() }
    foreground = RippleDrawable(ColorStateList.valueOf(0x24ffffff), null, GradientDrawable().apply { setColor(-1) })
}

internal fun LinearLayout.addWithGap(view: View, gap: Int, width: Int = -1, height: Int = -2) =
    addView(view, LinearLayout.LayoutParams(width, height).apply { if (childCount > 0) topMargin = gap })

internal fun MainActivity.avatar(open: () -> Unit) = FrameLayout(this).apply {
    background = panel(PHONE_BG, 22).apply { setStroke(ui(2), PHONE_ACCENT) }
    contentDescription = "账号与设置"
    val initial = account?.user?.optString("username").orEmpty().trim().take(1).uppercase()
    addView(phoneText(initial, 19, 0xffc4ead1.toInt(), true).apply {
        gravity = Gravity.CENTER
        background = panel(0xff354a40.toInt(), 17)
    }, FrameLayout.LayoutParams(ui(34), ui(34), Gravity.CENTER))
    onTap(open)
}
