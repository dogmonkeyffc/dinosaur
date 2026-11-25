package com.monkeydog.dinosaur.tv

import android.app.AlertDialog
import android.content.Context
import android.content.res.ColorStateList
import android.graphics.Color
import android.graphics.Typeface
import android.graphics.drawable.GradientDrawable
import android.graphics.drawable.StateListDrawable
import android.text.InputType
import android.view.Gravity
import android.view.View
import android.view.inputmethod.InputMethodManager
import android.widget.Button
import android.widget.EditText
import android.widget.FrameLayout
import android.widget.LinearLayout
import android.widget.ScrollView
import android.widget.TextView

internal const val surface = 0xff111619.toInt()
internal const val panel = 0xff1d282c.toInt()
internal const val accent = 0xffbeebc8.toInt()
internal const val muted = 0xffacbabc.toInt()

internal fun Context.dp(value: Int) = (value * resources.displayMetrics.density + .5f).toInt()

internal fun Context.column() = LinearLayout(this).apply { orientation = LinearLayout.VERTICAL }
internal fun Context.row() = LinearLayout(this).apply { gravity = Gravity.CENTER_VERTICAL }

internal fun Context.label(value: String, size: Float = 18f, color: Int = Color.WHITE, bold: Boolean = false) =
    TextView(this).apply {
        text = value
        textSize = size
        setTextColor(color)
        if (bold) setTypeface(typeface, Typeface.BOLD)
    }

internal fun Context.shape(color: Int, outlined: Boolean = false) = GradientDrawable().apply {
    setColor(color)
    cornerRadius = dp(10).toFloat()
    if (outlined) setStroke(dp(3), accent)
}

internal fun Context.focusSurface() = StateListDrawable().apply {
    addState(intArrayOf(android.R.attr.state_focused), shape(0xff354c48.toInt(), true))
    addState(intArrayOf(android.R.attr.state_selected), shape(0xff293a39.toInt()))
    addState(intArrayOf(), shape(panel))
}

internal fun Context.control(title: String, action: () -> Unit) = Button(this).apply {
    text = title
    textSize = 16f
    isAllCaps = false
    isFocusable = true
    isFocusableInTouchMode = true
    minHeight = dp(48)
    minimumHeight = dp(48)
    setPadding(dp(15), dp(4), dp(15), dp(4))
    background = focusSurface()
    setTextColor(ColorStateList(
        arrayOf(intArrayOf(-android.R.attr.state_enabled), intArrayOf()),
        intArrayOf(muted, Color.WHITE),
    ))
    setOnClickListener { action() }
}

internal fun LinearLayout.addControl(control: View, horizontal: Boolean = false) {
    val params = LinearLayout.LayoutParams(if (horizontal) -2 else -1, -2)
    if (horizontal) params.rightMargin = context.dp(10)
    else params.bottomMargin = context.dp(10)
    addView(control, params)
}

internal fun Context.scroll(content: View) = ScrollView(this).apply {
    isFocusable = false
    isVerticalScrollBarEnabled = false
    clipToPadding = false
    addView(content, FrameLayout.LayoutParams(-1, -2))
}

internal fun TvActivity.input(title: String, existing: String, address: Boolean = false, apply: (String) -> Unit) {
    val field = EditText(this).apply {
        inputType = if (address) InputType.TYPE_CLASS_TEXT or InputType.TYPE_TEXT_VARIATION_URI else InputType.TYPE_CLASS_TEXT
        setText(existing)
        setSelection(text.length)
        setSingleLine(true)
    }
    val wrapper = column().apply {
        setPadding(dp(24), dp(12), dp(24), 0)
        addView(field)
    }
    val dialog = AlertDialog.Builder(this).setTitle(title).setView(wrapper)
        .setNegativeButton("取消", null).setPositiveButton("确定") { _, _ -> apply(field.text.toString().trim()) }.create()
    field.setOnEditorActionListener { _, _, _ ->
        dialog.getButton(AlertDialog.BUTTON_POSITIVE).performClick()
        true
    }
    dialog.setOnShowListener {
        field.requestFocus()
        field.post { getSystemService(InputMethodManager::class.java).showSoftInput(field, InputMethodManager.SHOW_IMPLICIT) }
    }
    dialog.show()
}
