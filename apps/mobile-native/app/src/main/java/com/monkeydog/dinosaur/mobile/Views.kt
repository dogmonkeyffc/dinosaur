package com.monkeydog.dinosaur.mobile

import android.app.AlertDialog
import android.graphics.Color
import android.widget.*

internal fun MainActivity.dp(value: Int) = (value * resources.displayMetrics.density + 0.5f).toInt()

internal fun MainActivity.page(title: String, previous: (() -> Unit)? = null) {
    generation++
    back = previous
    val root = LinearLayout(this).apply {
        orientation = LinearLayout.VERTICAL
        setBackgroundColor(Color.rgb(17, 17, 19))
        fitsSystemWindows = true
    }
    val heading = LinearLayout(this).apply { gravity = android.view.Gravity.CENTER_VERTICAL }
    previous?.let { action ->
        heading.addView(Button(this).apply {
            text = "‹"
            contentDescription = "返回"
            setOnClickListener { action() }
        })
    }
    heading.addView(TextView(this).apply {
        text = title
        textSize = 22f
        setTextColor(Color.WHITE)
        setPadding(dp(16), dp(16), dp(16), dp(16))
    })
    root.addView(heading)
    content = LinearLayout(this).apply {
        orientation = LinearLayout.VERTICAL
        setPadding(dp(16), 0, dp(16), dp(24))
    }
    root.addView(ScrollView(this).apply { addView(content) }, LinearLayout.LayoutParams(-1, 0, 1f))
    setContentView(root)
}

internal fun MainActivity.label(value: String) {
    content.addView(TextView(this).apply {
        text = value
        textSize = 16f
        setTextColor(Color.rgb(230, 230, 236))
        setPadding(0, dp(10), 0, dp(10))
    })
}

internal fun MainActivity.button(title: String, action: () -> Unit) {
    content.addView(Button(this).apply {
        text = title
        isAllCaps = false
        setOnClickListener { action() }
    })
}

internal fun MainActivity.rowSwitch(title: String, value: Boolean, change: (Boolean, Switch) -> Unit) {
    content.addView(Switch(this).apply {
        text = title
        isChecked = value
        setPadding(0, dp(10), 0, dp(10))
        setOnClickListener { change(isChecked, this) }
    })
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
