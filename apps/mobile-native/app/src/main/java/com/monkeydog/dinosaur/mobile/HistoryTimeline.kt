package com.monkeydog.dinosaur.mobile

import android.app.Dialog
import android.content.Intent
import android.net.Uri
import android.view.Gravity
import android.view.View
import android.widget.FrameLayout
import android.widget.LinearLayout
import android.widget.ScrollView
import android.widget.TextView
import com.monkeydog.dinosaur.android.Api
import com.monkeydog.dinosaur.android.rows
import org.json.JSONObject

internal class HistoryTimeline(private val activity: MainActivity, private val artwork: PhoneArtwork) {
    fun show(chrome: CatalogChrome, client: Api): Unit = with(activity) {
        chrome.header.removeAllViews()
        chrome.body.removeAllViews()
        if (chrome.stage.childCount > 1) chrome.stage.removeViewAt(1)
        val (heading, count) = heading()
        chrome.header.addView(heading)
        val pinned = monthHeading("").apply {
            importantForAccessibility = View.IMPORTANT_FOR_ACCESSIBILITY_NO
            visibility = View.GONE
        }
        chrome.stage.addView(pinned, FrameLayout.LayoutParams(-1, -2, Gravity.TOP))
        val sections = mutableListOf<TextView>()
        var loadingPage = false
        var more = true
        var month = ""
        var number = 1
        var failure: View? = null
        fun load() {
            if (loadingPage || !more) return
            loadingPage = true
            failure?.let(chrome.body::removeView)
            failure = null
            val spinner = loading()
            chrome.body.addView(spinner)
            task({ client.request("/api/history?page=$number&page_size=24") }, { error ->
                loadingPage = false
                chrome.body.removeView(spinner)
                failure = column().apply { feedback(this, error.message ?: "无法加载观看记录", ::load) }
                    .also(chrome.body::addView)
            }) { data ->
                loadingPage = false
                chrome.body.removeView(spinner)
                val total = data.optInt("total")
                val items = data.rows("items")
                count.text = total.toString()
                month = append(chrome.body, items, client, month, sections)
                more = items.isNotEmpty() && number * 24 < total
                if (number == 1 && items.isEmpty()) chrome.body.addView(status("暂无观看记录"))
                number++
            }
        }
        chrome.scroller.setOnScrollChangeListener { _, _, y, _, oldY ->
            val current = sections.indexOfLast { it.top < y }
            pinned.visibility = if (current >= 0) View.VISIBLE else View.GONE
            if (current >= 0) {
                pinned.updateText(sections[current].text.toString())
                val nextTop = sections.getOrNull(current + 1)?.top ?: Int.MAX_VALUE
                pinned.translationY = (nextTop - y - sections[current].height).coerceAtMost(0).toFloat()
            }
            if (y > oldY && chrome.body.height - y - chrome.scroller.height < ui(150)) load()
        }
        chrome.scroller.refresh = {
            generation++
            chrome.scroller.setOnScrollChangeListener(null as View.OnScrollChangeListener?)
            show(chrome, client)
        }
        load()
    }

    private fun heading() = with(activity) {
        val count = phoneText("", 15, PHONE_MUTED)
        phoneRow().apply {
            setPadding(ui(24), ui(18), ui(24), ui(18))
            addView(phoneText("观看记录", 28, bold = true), LinearLayout.LayoutParams(0, -2, 1f))
            addView(count, LinearLayout.LayoutParams(-2, -2).apply { rightMargin = ui(18) })
            addView(avatar { settings.show() }, LinearLayout.LayoutParams(ui(44), ui(44)))
        } to count
    }

    private fun append(
        body: LinearLayout, items: List<JSONObject>, api: Api, lastMonth: String, sections: MutableList<TextView>,
    ): String = with(activity) {
        var month = lastMonth
        items.forEach { item ->
            val date = item.getString("watched_on")
            val next = date.take(7)
            if (next != month) {
                month = next
                val monthNumber = month.drop(5).toIntOrNull()
                val heading = monthHeading("${month.take(4)}年${monthNumber ?: month.drop(5)}月")
                body.addView(heading)
                sections.add(heading)
            }
            body.addView(entry(item, api, date))
        }
        month
    }

    private fun MainActivity.monthHeading(title: String) = phoneText(title, 18, PHONE_ACCENT, true).apply {
        setBackgroundColor(PHONE_BG)
        setPadding(ui(24), ui(16), ui(24), ui(16))
        if (android.os.Build.VERSION.SDK_INT >= 28) isAccessibilityHeading = true
    }

    private fun entry(item: JSONObject, api: Api, date: String) = with(activity) {
        val title = item.getString("title")
        phoneRow().apply {
            gravity = Gravity.TOP
            setPadding(ui(24), 0, ui(24), 0)
            addView(column().apply {
                gravity = Gravity.CENTER_HORIZONTAL
                addView(phoneText((date.drop(8).toIntOrNull() ?: date.drop(8)).toString(), 20, bold = true))
                addWithGap(View(this@with).apply { background = panel(PHONE_ACCENT, 4) }, ui(10), ui(7), ui(7))
                addView(View(this@with).apply { setBackgroundColor(0xff333c4c.toInt()) },
                    LinearLayout.LayoutParams(ui(1), 0, 1f).apply { topMargin = ui(6); bottomMargin = ui(6) })
            }, LinearLayout.LayoutParams(ui(32), -1))
            addView(phoneRow().apply {
                gravity = Gravity.TOP
                setPadding(0, 0, 0, ui(28))
                addView(artwork.poster(item, api), LinearLayout.LayoutParams(ui(72), ui(108)))
                addView(column().apply {
                    addView(phoneText(title, 17, bold = true).apply { maxLines = 2 })
                    stars(item)?.let { addWithGap(it, ui(8)) }
                    item.optString("review").takeIf { it.isNotBlank() && it != "null" }?.let {
                        addWithGap(phoneText(it, 14, PHONE_MUTED).apply { maxLines = 2 }, ui(8))
                    }
                    addWithGap(phoneText(source(item), 12, PHONE_MUTED), ui(8))
                }, LinearLayout.LayoutParams(0, -2, 1f).apply { leftMargin = ui(16) })
                contentDescription = "$date $title"
                onTap { detail(item, date) }
            }, LinearLayout.LayoutParams(0, -2, 1f).apply { leftMargin = ui(16) })
        }
    }

    private fun detail(item: JSONObject, date: String) = with(activity) {
        val dialog = Dialog(this)
        val layout = column().apply {
            background = panel(0xff222429.toInt(), 26)
            setPadding(ui(22), ui(16), ui(22), ui(22))
            addView(phoneRow().apply {
                addView(phoneText(date, 12, PHONE_MUTED), LinearLayout.LayoutParams(0, -2, 1f))
                addView(phoneAction("×", { dialog.dismiss() }, size = 28, color = PHONE_MUTED, fill = 0)
                    .apply { contentDescription = "关闭观看记录" })
            })
            addView(ScrollView(this@with).apply {
                addView(column().apply {
                    addView(phoneText(item.getString("title"), 24, bold = true))
                    item.optString("original_title").takeIf { it.isNotBlank() && it != "null" }?.let {
                        addWithGap(phoneText(it, 14, PHONE_MUTED), ui(16))
                    }
                    item.optInt("year").takeIf { it > 0 }?.let { addWithGap(phoneText(it.toString(), 14, PHONE_MUTED), ui(16)) }
                    stars(item)?.let { addWithGap(it, ui(16)) }
                    item.optString("review").takeIf { it.isNotBlank() && it != "null" }?.let {
                        addWithGap(phoneText(it, 16).apply { setTextIsSelectable(true) }, ui(16))
                    }
                    val id = item.optString("source_id")
                    val note = if (source(item) == "豆瓣" && id.matches(Regex("[1-9][0-9]*"))) {
                        phoneAction("豆瓣", {
                            runCatching { startActivity(Intent(Intent.ACTION_VIEW, Uri.parse("https://movie.douban.com/subject/$id/"))) }
                                .onFailure { message("无法打开豆瓣链接") }
                        })
                    } else phoneText(source(item), 12, PHONE_MUTED)
                    addWithGap(note, ui(16))
                })
            }, LinearLayout.LayoutParams(-1, -2))
        }
        dialog.setContentView(layout)
        dialog.window?.setBackgroundDrawableResource(android.R.color.transparent)
        dialog.show()
        layout.measure(View.MeasureSpec.makeMeasureSpec(resources.displayMetrics.widthPixels - ui(48), View.MeasureSpec.EXACTLY),
            View.MeasureSpec.makeMeasureSpec((resources.displayMetrics.heightPixels * .85).toInt(), View.MeasureSpec.AT_MOST))
        dialog.window?.setLayout(resources.displayMetrics.widthPixels - ui(48), layout.measuredHeight)
    }

    private fun stars(item: JSONObject): TextView? = with(activity) {
        val rating = item.optInt("rating").coerceIn(0, 5)
        if (rating == 0) null else phoneText("★".repeat(rating) + "☆".repeat(5 - rating), 15, 0xffe4c67f.toInt())
            .apply { contentDescription = "$rating 星"; letterSpacing = .13f }
    }

    private fun source(item: JSONObject) = when (item.optString("source")) {
        "douban" -> "豆瓣"
        "local" -> "站内"
        else -> "手动记录"
    }
}
