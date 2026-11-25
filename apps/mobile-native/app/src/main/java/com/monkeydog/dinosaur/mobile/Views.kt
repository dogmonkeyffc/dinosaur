package com.monkeydog.dinosaur.mobile

import android.app.AlertDialog
import android.content.res.ColorStateList
import android.view.View
import android.widget.*

internal fun MainActivity.dp(value: Int) = (value * resources.displayMetrics.density + 0.5f).toInt()

internal fun MainActivity.page(
    title: String, previous: (() -> Unit)? = null, close: (() -> Unit)? = null, toolbarAction: View? = null,
) {
    generation++
    back = previous
    val root = LinearLayout(this).apply {
        orientation = LinearLayout.VERTICAL
        setBackgroundColor(PHONE_BG)
        fitsSystemWindows = true
    }
    val heading = phoneRow().apply { setPadding(ui(12), ui(8), ui(12), ui(8)) }
    previous?.let { action ->
        heading.addView(PlayerAction(this, PlayerIcon.BACK, "返回", action), LinearLayout.LayoutParams(ui(48), ui(48)))
    }
    heading.addView(phoneText(title, 24, bold = true), LinearLayout.LayoutParams(0, -2, 1f)
        .apply { leftMargin = if (previous == null) ui(10) else ui(8) })
    toolbarAction?.let { heading.addView(it, LinearLayout.LayoutParams(ui(48), ui(48))) }
    close?.let { action ->
        heading.addView(phoneAction("×", action, size = 26, color = PHONE_TEXT, fill = 0).apply {
            contentDescription = "关闭$title"
        }, LinearLayout.LayoutParams(ui(48), ui(48)))
    }
    root.addView(heading)
    content = LinearLayout(this).apply {
        orientation = LinearLayout.VERTICAL
        setPadding(ui(22), ui(6), ui(22), ui(40))
    }
    root.addView(ScrollView(this).apply { addView(content) }, LinearLayout.LayoutParams(-1, 0, 1f))
    setContentView(root)
}

internal fun MainActivity.label(value: String) = content.addView(phoneText(value, 18).apply {
    setPadding(0, ui(10), 0, ui(10))
})

internal fun MainActivity.button(title: String, action: () -> Unit) =
    content.addWithGap(phoneAction(title, action, color = PHONE_TEXT, fill = 0xff303136.toInt()), ui(8))

internal fun MainActivity.rowSwitch(title: String, value: Boolean, change: (Boolean, Switch) -> Unit) =
    content.addView(switchControl(title, value, change))

internal fun MainActivity.switchControl(title: String, value: Boolean, change: (Boolean, Switch) -> Unit) =
    Switch(this).apply {
        text = title
        isChecked = value
        setTextColor(PHONE_TEXT)
        textSize = 16 * .85f
        thumbTintList = ColorStateList.valueOf(PHONE_ACCENT)
        trackTintList = ColorStateList(
            arrayOf(intArrayOf(android.R.attr.state_checked), intArrayOf()),
            intArrayOf(0xff4c638b.toInt(), 0xff44464e.toInt()),
        )
        minHeight = ui(66)
        setPadding(ui(18), ui(10), ui(18), ui(10))
        setOnClickListener { change(isChecked, this) }
    }

internal fun MainActivity.message(value: String) = Toast.makeText(this, value, Toast.LENGTH_LONG).show()

internal fun MainActivity.confirm(title: String, action: () -> Unit) {
    AlertDialog.Builder(this).setTitle(title).setNegativeButton("取消", null)
        .setPositiveButton("确认") { _, _ -> action() }.show()
}

internal fun MainActivity.choose(title: String, choices: List<String>, selected: Int = -1, action: (Int) -> Unit) {
    AlertDialog.Builder(this).setTitle(title)
        .setSingleChoiceItems(choices.toTypedArray(), selected) { dialog, index ->
            dialog.dismiss()
            action(index)
        }.setNegativeButton("取消", null).show()
}

internal fun MainActivity.form(title: String, labels: List<String>, values: List<String> = emptyList(), submit: (List<String>) -> Unit) {
    val container = LinearLayout(this).apply {
        orientation = LinearLayout.VERTICAL
        setPadding(dp(20), dp(8), dp(20), dp(8))
    }
    val inputs = labels.mapIndexed { index, name ->
        EditText(this).apply {
            hint = name
            contentDescription = name
            setSingleLine()
            if (name.contains("密码") || name.contains("密钥")) inputType = 129
            setText(values.getOrNull(index).orEmpty())
            container.addView(this)
        }
    }
    AlertDialog.Builder(this).setTitle(title).setView(ScrollView(this).apply { addView(container) })
        .setPositiveButton("确认") { _, _ -> submit(inputs.map { it.text.toString() }) }
        .setNegativeButton("取消", null).show()
}
