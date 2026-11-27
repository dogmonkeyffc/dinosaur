package com.monkeydog.dinosaur.mobile

import android.text.InputType
import android.view.Gravity
import android.view.View
import android.widget.EditText
import android.widget.HorizontalScrollView
import android.widget.LinearLayout
import android.widget.TextView
import org.json.JSONObject

internal class AdminLibraryList(
    private val activity: MainActivity,
    initialType: String,
    initialSearch: String,
    private val menu: (View, JSONObject) -> Unit,
) {
    private val types = linkedMapOf("movie" to "电影", "tv" to "电视剧", "video" to "普通视频")
    private val list = activity.column()
    private val filters = activity.phoneRow()
    private val chips = linkedMapOf<String, TextView>()
    private var rows = emptyList<JSONObject>()
    var type = initialType
        private set
    var search = initialSearch
        private set

    init {
        with(activity) {
            (listOf("" to "全部") + types.toList()).forEach { (key, title) ->
                val chip = phoneAction(title, {
                    type = key
                    updateChips()
                    render()
                })
                chips[key] = chip
                filters.addView(chip, LinearLayout.LayoutParams(-2, -2).apply {
                    if (filters.childCount > 0) leftMargin = ui(8)
                })
            }
            updateChips()
            content.addWithGap(adminSearch(search) {
                search = it
                render()
            }, ui(16))
            content.addWithGap(HorizontalScrollView(this).apply {
                isHorizontalScrollBarEnabled = false
                addView(filters)
            }, ui(12))
            content.addWithGap(list, ui(18))
        }
    }

    fun setRows(source: List<JSONObject>) {
        rows = source
        chips.getValue("").text = "全部 ${rows.size}"
        render()
    }

    fun setScan(id: Long, scan: JSONObject) {
        rows.firstOrNull { it.optLong("id") == id }?.put("scan", scan)
        render()
    }

    private fun updateChips() = with(activity) {
        chips.forEach { (key, chip) -> styleAdminFilter(chip, key == type) }
    }

    private fun render() = with(activity) {
        list.removeAllViews()
        val term = search.trim()
        rows.filter { item ->
            (type.isEmpty() || item.optString("type") == type) && item.optString("name").contains(term, true)
        }.forEach { item ->
            list.addWithGap(adminLibraryCard(item, types[item.optString("type")].orEmpty(), menu), ui(16))
        }
        if (list.childCount == 0) list.addView(phoneText(
            if (rows.isEmpty()) "暂无媒体库" else "没有找到匹配的媒体库", 16, PHONE_MUTED,
        ))
    }
}

internal fun JSONObject.scanActive() = optJSONObject("scan")?.optString("status") in listOf("queued", "scanning")

private fun scanName(status: String) = when (status) {
    "queued" -> "等待扫描"
    "scanning" -> "扫描中"
    "completed" -> "扫描完成"
    "failed" -> "扫描失败"
    "interrupted" -> "扫描中断"
    else -> ""
}

private fun MainActivity.styleAdminFilter(chip: TextView, selected: Boolean) {
    chip.isSelected = selected
    chip.setTextColor(if (selected) PHONE_ACCENT else PHONE_MUTED)
    chip.background = panel(if (selected) PHONE_SELECTED else PHONE_SURFACE, 25)
}

private fun MainActivity.adminSearch(search: String, update: (String) -> Unit) = phoneRow().apply {
    background = panel(PHONE_SURFACE, 30)
    setPadding(ui(18), 0, ui(12), 0)
    addView(CatalogIcon(this@adminSearch, "search", PHONE_MUTED), LinearLayout.LayoutParams(ui(24), ui(24)))
    addView(EditText(this@adminSearch).apply {
        setSingleLine()
        inputType = InputType.TYPE_CLASS_TEXT
        hint = "搜索媒体库"
        setText(search)
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
    item: JSONObject, type: String, menu: (View, JSONObject) -> Unit,
) = phoneCard().apply {
    val name = item.getString("name")
    addView(phoneRow().apply {
        addView(phoneText(name, 23, bold = true), LinearLayout.LayoutParams(0, -2, 1f))
        addView(phoneText("⋯", 24, PHONE_ACCENT).apply {
            gravity = Gravity.CENTER
            contentDescription = "管理 $name"
            onTap { menu(this, item) }
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
}
