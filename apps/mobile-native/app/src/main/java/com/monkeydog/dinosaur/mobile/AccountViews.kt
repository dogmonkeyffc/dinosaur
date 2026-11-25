package com.monkeydog.dinosaur.mobile

import android.view.Gravity
import android.widget.LinearLayout
import com.monkeydog.dinosaur.android.Account

internal fun MainActivity.accountPage(title: String, previous: (() -> Unit)?) {
    page(title, previous)
    content.setPadding(ui(22), ui(26), ui(22), ui(40))
}

internal fun MainActivity.phoneCard() = column().apply {
    background = panel(0xff202125.toInt(), 28)
    setPadding(ui(22), ui(22), ui(22), ui(22))
}

internal fun MainActivity.accountRow(item: Account, selected: Boolean = false) = phoneRow().apply {
    val username = item.user.getString("username")
    isSelected = selected
    addView(phoneText(username.take(1).uppercase(), 24, PHONE_ACCENT, true).apply {
        gravity = Gravity.CENTER
        background = panel(PHONE_SELECTED, 27)
    }, LinearLayout.LayoutParams(ui(54), ui(54)))
    addView(column().apply {
        addView(phoneText(username, 19, bold = true))
        addWithGap(phoneText(item.server.optString("name").ifBlank { item.origin }, 14, PHONE_MUTED).apply {
            maxLines = 1
            ellipsize = android.text.TextUtils.TruncateAt.END
        }, ui(6))
    }, LinearLayout.LayoutParams(0, -2, 1f).apply { leftMargin = ui(16); rightMargin = ui(8) })
    if (selected) addView(phoneText("✓", 24, PHONE_ACCENT))
}
