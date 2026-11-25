package com.monkeydog.dinosaur.mobile

import android.text.InputType
import android.view.Gravity
import android.widget.EditText
import android.widget.HorizontalScrollView
import android.widget.LinearLayout
import android.widget.TextView
import org.json.JSONObject

internal fun MainActivity.adminMenuButton(open: () -> Unit) = phoneText("☰", 24, PHONE_ACCENT).apply {
    gravity = Gravity.CENTER
    contentDescription = "管理导航"
    onTap(open)
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
                { more(item) }, { browse(item) }, { scan(item) }), ui(16))
        }
        if (list.childCount == 0) list.addView(phoneText(
            if (rows.isEmpty()) "暂无媒体库" else "没有找到匹配的媒体库", 16, PHONE_MUTED))
    }
    types.forEach { (key, title) ->
        filters.addView(phoneAction(title, {
            type = key
            for (index in 0 until filters.childCount) {
                val chip = filters.getChildAt(index) as TextView
                styleAdminFilter(chip, chip.tag == type)
            }
            refresh()
        }).apply {
            tag = key
            styleAdminFilter(this, key == type)
        }, LinearLayout.LayoutParams(-2, -2).apply { if (filters.childCount > 0) leftMargin = ui(8) })
    }
    content.addWithGap(adminSearch { query ->
        search = query
        refresh()
    }, ui(16))
    content.addWithGap(HorizontalScrollView(this).apply {
        isHorizontalScrollBarEnabled = false
        addView(filters)
    }, ui(12))
    content.addWithGap(list, ui(18))
    refresh()
}

private fun MainActivity.styleAdminFilter(chip: TextView, selected: Boolean) {
    chip.isSelected = selected
    chip.setTextColor(if (selected) PHONE_ACCENT else PHONE_MUTED)
    chip.background = panel(if (selected) PHONE_SELECTED else PHONE_SURFACE, 25)
}

private fun MainActivity.adminSearch(update: (String) -> Unit) = phoneRow().apply {
    background = panel(PHONE_SURFACE, 30)
    setPadding(ui(18), 0, ui(12), 0)
    addView(CatalogIcon(this@adminSearch, "search", PHONE_MUTED), LinearLayout.LayoutParams(ui(24), ui(24)))
    addView(EditText(this@adminSearch).apply {
        setSingleLine()
        inputType = InputType.TYPE_CLASS_TEXT
        hint = "搜索媒体库"
        textSize = 17 * .85f
        setTextColor(PHONE_TEXT)
        setHintTextColor(PHONE_MUTED)
        setPadding(ui(14), 0, 0, 0)
        background = null
        addTextChangedListener(object : android.text.TextWatcher {
            override fun beforeTextChanged(s: CharSequence?, start: Int, count: Int, after: Int) = Unit
            override fun afterTextChanged(s: android.text.Editable?) = Unit
            override fun onTextChanged(s: CharSequence?, start: Int, before: Int, count: Int) {
                update(s.toString())
            }
        })
    }, LinearLayout.LayoutParams(0, ui(56), 1f))
}

private fun MainActivity.adminLibraryCard(
    item: JSONObject, type: String, scanName: (String) -> String, more: () -> Unit, browse: () -> Unit, scan: () -> Unit,
) = phoneCard().apply {
    val name = item.getString("name")
    addView(phoneRow().apply {
        addView(phoneText(name, 23, bold = true), LinearLayout.LayoutParams(0, -2, 1f))
        addView(phoneText("⋯", 24, PHONE_ACCENT).apply {
            gravity = Gravity.CENTER
            contentDescription = "管理 $name"
            onTap(more)
        }, LinearLayout.LayoutParams(ui(48), ui(48)))
    })
    if (type.isNotBlank() && type != name) addView(phoneText(type, 13, PHONE_MUTED))
    val path = item.optString("root_path")
    if (path.isNotBlank()) addWithGap(phoneText(path, 15, PHONE_MUTED), ui(10))
    val job = item.optJSONObject("scan")
    val status = job?.optString("status").orEmpty()
    val available = item.optBoolean("available", true)
    addWithGap(phoneText(if (available) "目录可访问" else "目录暂不可用", 14,
        if (available) PHONE_MUTED else 0xffffb4ab.toInt()), ui(12))
    val media = when (item.optString("type")) {
        "tv" -> "剧集"
        "video" -> "视频"
        else -> "电影"
    }
    val summary = scanName(status)
    if (summary.isNotBlank()) {
        val processed = job?.optInt("processed") ?: 0
        addWithGap(phoneText("$summary · $processed 个$media", 14, PHONE_MUTED), ui(6))
    }
    val scanning = status == "queued" || status == "scanning"
    addWithGap(phoneRow().apply {
        addView(phoneAction("查看$media ${item.optInt("media_count")}", browse, color = PHONE_ACCENT, fill = PHONE_SELECTED),
            LinearLayout.LayoutParams(0, -2, 1f))
        addView(phoneAction(if (scanning) "正在扫描…" else "重新扫描", scan, color = PHONE_TEXT).apply {
            isEnabled = !scanning
            if (scanning) alpha = .5f
        }, LinearLayout.LayoutParams(0, -2, 1f).apply { leftMargin = ui(10) })
    }, ui(16))
}
