package com.monkeydog.dinosaur.mobile

import android.graphics.Typeface
import android.text.InputType
import android.view.Gravity
import android.widget.EditText
import android.widget.HorizontalScrollView
import android.widget.LinearLayout
import android.widget.ScrollView
import org.json.JSONObject

internal const val ADMIN_TEXT = 0xffe1e6db.toInt()
private const val ADMIN_MUTED = 0xffc4cbbd.toInt()
private const val ADMIN_BG = 0xff121510.toInt()
private const val ADMIN_SURFACE = 0xff1e231c.toInt()
private const val ADMIN_LINE = 0xff43483e.toInt()
private const val ADMIN_ACCENT = 0xffa9d18f.toInt()

internal fun MainActivity.adminPage(close: () -> Unit, menu: () -> Unit) {
    page("管理", close, close)
    val scroller = content.parent as ScrollView
    scroller.setBackgroundColor(ADMIN_BG)
    val body = column().apply { setPadding(dp(16), dp(16), dp(16), dp(28)) }
    content.setPadding(0, 0, 0, 0)
    content.addView(phoneRow().apply {
        setPadding(dp(16), dp(12), dp(16), dp(12))
        setBackgroundColor(ADMIN_SURFACE)
        addView(phoneText("▩", 22, 0xffdce8d2.toInt(), true).apply {
            gravity = Gravity.CENTER
            background = panel(0xff384832.toInt(), 14)
        }, LinearLayout.LayoutParams(dp(40), dp(40)))
        addView(phoneText("Dinosaur", 19, ADMIN_TEXT, true), LinearLayout.LayoutParams(0, -2, 1f)
            .apply { leftMargin = dp(12) })
        addView(adminAction("☰ 菜单", menu).apply { contentDescription = "管理导航" })
    })
    content.addView(body)
    content = body
}

internal fun MainActivity.adminAction(title: String, action: () -> Unit, primary: Boolean = false) =
    phoneText(title, 15, if (primary) 0xff18350f.toInt() else ADMIN_ACCENT, true).apply {
        gravity = Gravity.CENTER
        setPadding(dp(10), dp(7), dp(10), dp(7))
        minimumHeight = dp(34)
        if (primary) background = panel(ADMIN_ACCENT, 4)
        onTap(action)
    }

internal fun MainActivity.adminLibraryList(
    rows: List<JSONObject>, scanName: (String) -> String,
    more: (JSONObject) -> Unit, browse: (JSONObject) -> Unit, scan: (JSONObject) -> Unit,
) {
    val types = linkedMapOf("" to "全部 ${rows.size}", "movie" to "电影", "tv" to "电视剧", "video" to "普通视频")
    val filters = phoneRow()
    val list = column()
    var type = ""
    var search = ""
    fun refresh() {
        list.removeAllViews()
        rows.filter { item ->
            (type.isEmpty() || item.optString("type") == type) && item.optString("name").contains(search.trim(), true)
        }.forEach { item ->
            list.addWithGap(adminLibraryCard(item, types[item.optString("type")].orEmpty(), scanName,
                { more(item) }, { browse(item) }, { scan(item) }), dp(16))
        }
        if (list.childCount == 0) list.addView(phoneText(
            if (rows.isEmpty()) "暂无媒体库" else "没有找到匹配的媒体库", 16, ADMIN_MUTED))
    }
    types.forEach { (key, title) ->
        filters.addView(phoneText(title, 15, ADMIN_TEXT).apply {
            gravity = Gravity.CENTER
            setPadding(dp(7), dp(8), dp(7), dp(8))
            tag = key
            isSelected = key == type
            background = panel(if (isSelected) 0xff343a32.toInt() else ADMIN_BG, 4, ADMIN_LINE)
            onTap {
                type = key
                for (index in 0 until filters.childCount) {
                    val button = filters.getChildAt(index)
                    button.isSelected = button.tag == type
                    button.background = panel(if (button.isSelected) 0xff343a32.toInt() else ADMIN_BG, 4, ADMIN_LINE)
                }
                refresh()
            }
        })
    }
    content.addWithGap(HorizontalScrollView(this).apply {
        isHorizontalScrollBarEnabled = false
        addView(filters)
    }, dp(18))
    content.addWithGap(EditText(this).apply {
        setSingleLine()
        inputType = InputType.TYPE_CLASS_TEXT
        hint = "搜索媒体库"
        textSize = 15f
        setTextColor(ADMIN_TEXT)
        setHintTextColor(ADMIN_MUTED)
        setPadding(dp(16), dp(8), dp(16), dp(8))
        background = panel(ADMIN_BG, 4, ADMIN_LINE)
        addTextChangedListener(object : android.text.TextWatcher {
            override fun beforeTextChanged(s: CharSequence?, start: Int, count: Int, after: Int) = Unit
            override fun afterTextChanged(s: android.text.Editable?) = Unit
            override fun onTextChanged(s: CharSequence?, start: Int, before: Int, count: Int) {
                search = s.toString()
                refresh()
            }
        })
    }, dp(8))
    content.addWithGap(list, dp(16))
    refresh()
}

private fun MainActivity.adminLibraryCard(
    item: JSONObject, type: String, scanName: (String) -> String, more: () -> Unit, browse: () -> Unit, scan: () -> Unit,
) = column().apply {
    background = panel(ADMIN_SURFACE, 4)
    setPadding(dp(16), dp(12), dp(16), dp(16))
    addView(phoneRow().apply {
        addView(phoneText(if (item.optString("type") == "tv") "▣" else "▤", 20, ADMIN_TEXT))
        addView(android.view.View(this@adminLibraryCard), LinearLayout.LayoutParams(0, 1, 1f))
        addView(phoneText(type, 15, ADMIN_MUTED))
        addView(adminAction("⋯", more).apply { contentDescription = "管理 ${item.getString("name")}" })
    })
    addWithGap(phoneText(item.getString("name"), 23, ADMIN_TEXT, true), dp(12))
    val path = item.optString("root_path")
    if (path.isNotBlank()) addWithGap(phoneText(path, 16, ADMIN_MUTED).apply { typeface = Typeface.MONOSPACE }, dp(14))
    val job = item.optJSONObject("scan")
    val status = job?.optString("status").orEmpty()
    val available = item.optBoolean("available", true)
    addWithGap(phoneText(if (available) "● 目录可访问" else "● 目录暂不可用", 15,
        if (available) ADMIN_MUTED else 0xffffb4ab.toInt()), dp(14))
    val media = when (item.optString("type")) { "tv" -> "剧集"; "video" -> "视频"; else -> "电影" }
    val scanning = status == "queued" || status == "scanning"
    addView(phoneRow().apply {
        addView(adminAction("查看$media ${item.optInt("media_count")}", browse))
        addView(android.view.View(this@adminLibraryCard), LinearLayout.LayoutParams(0, 1, 1f))
        addView(adminAction(if (scanning) "正在扫描…" else "重新扫描", scan).apply { isEnabled = !scanning })
    })
    val summary = scanName(status)
    if (summary.isNotBlank()) {
        val processed = job?.optInt("processed") ?: 0
        addWithGap(phoneText("$summary · $processed 个$media", 15, ADMIN_MUTED), dp(4))
    }
}
