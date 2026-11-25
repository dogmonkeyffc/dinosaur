package com.monkeydog.dinosaur.mobile

import android.view.Gravity
import android.widget.LinearLayout
import android.widget.ProgressBar

internal fun MainActivity.loading() = phoneRow().apply {
    gravity = Gravity.CENTER
    setPadding(0, ui(22), 0, ui(22))
    addView(ProgressBar(this@loading), LinearLayout.LayoutParams(ui(26), ui(26)))
}

internal fun MainActivity.status(text: String) = phoneText(text, 15, PHONE_MUTED).apply {
    setPadding(ui(24), ui(12), ui(24), ui(12))
}

internal fun MainActivity.feedback(body: LinearLayout, error: String, retry: () -> Unit) {
    body.removeAllViews()
    body.addView(status(error))
    body.addView(phoneAction("重试", retry), LinearLayout.LayoutParams(-2, -2).apply { leftMargin = ui(24) })
}
