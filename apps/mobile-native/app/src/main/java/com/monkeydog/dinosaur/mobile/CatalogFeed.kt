package com.monkeydog.dinosaur.mobile

import android.graphics.Color
import android.text.TextUtils
import android.view.Gravity
import android.view.View
import android.widget.FrameLayout
import android.widget.HorizontalScrollView
import android.widget.LinearLayout
import com.monkeydog.dinosaur.android.Api
import org.json.JSONObject
import kotlin.math.roundToInt

internal class CatalogFeed(private val activity: MainActivity, private val artwork: PhoneArtwork) {
    fun home(
        body: LinearLayout, title: String, items: List<JSONObject>, recent: List<JSONObject>, api: Api,
        browse: () -> Unit, open: (JSONObject) -> Unit, play: (JSONObject) -> Unit,
    ) = with(activity) {
        body.setPadding(0, ui(22), 0, ui(22))
        if (recent.isNotEmpty()) body.addWithGap(shelf("继续观看", recent.take(8), api, null, play), ui(28))
        if (items.isNotEmpty()) body.addWithGap(shelf(title, items.take(10), api, browse, open), ui(28))
        items.firstOrNull { it.optString("poster_url").let { url -> url.isNotBlank() && url != "null" } }?.let { item ->
            body.addView(feature(item, api, open), LinearLayout.LayoutParams(-1, -2).apply {
                setMargins(ui(24), ui(28), ui(24), 0)
            })
        }
        if (items.size > 10) body.addWithGap(shelf("更多内容", items.drop(10), api, browse, open), ui(28))
    }

    private fun shelf(
        title: String, items: List<JSONObject>, api: Api, more: (() -> Unit)?, open: (JSONObject) -> Unit,
    ) = with(activity) {
        column().apply {
            addView(phoneRow().apply {
                setPadding(ui(24), 0, ui(24), 0)
                addView(phoneText(title, 21, bold = true), LinearLayout.LayoutParams(0, -2, 1f))
                more?.let {
                    val arrow = phoneText("→", 24).apply {
                        gravity = Gravity.CENTER
                        background = panel(0xff232427.toInt(), 20)
                        contentDescription = "查看全部$title"
                        onTap(it)
                    }
                    addView(arrow, LinearLayout.LayoutParams(ui(40), ui(40)))
                }
            })
            val width = ui(((resources.configuration.screenWidthDp - 64) / 3.1).coerceIn(100.0, 150.0))
            addWithGap(HorizontalScrollView(this@with).apply {
                isHorizontalScrollBarEnabled = false
                addView(phoneRow().apply {
                    gravity = Gravity.TOP
                    setPadding(ui(24), 0, ui(24), 0)
                    items.forEachIndexed { index, item ->
                        addView(card(item, api, width, open), LinearLayout.LayoutParams(width, -2).apply {
                            if (index > 0) leftMargin = ui(12)
                        })
                    }
                })
            }, ui(16))
        }
    }

    fun search(body: LinearLayout, items: List<JSONObject>, api: Api, queried: Boolean, open: (JSONObject) -> Unit) =
        with(activity) {
            body.setPadding(ui(24), ui(22), ui(24), ui(22))
            body.addView(phoneText(if (queried) "搜索结果" else "浏览片库", 21, bold = true))
            val dpWidth = resources.configuration.screenWidthDp
            val columns = if (dpWidth > 700) 5 else if (dpWidth >= 360) 3 else 2
            val width = (dp(dpWidth) - ui(48) - ui(14) * (columns - 1)) / columns
            items.chunked(columns).forEachIndexed { index, group ->
                body.addWithGap(phoneRow().apply {
                    gravity = Gravity.TOP
                    group.forEachIndexed { column, item ->
                        addView(card(item, api, width, open), LinearLayout.LayoutParams(width, -2).apply {
                            if (column > 0) leftMargin = ui(14)
                        })
                    }
                }, ui(if (index == 0) 28 else 14))
            }
        }

    private fun card(item: JSONObject, api: Api, width: Int, open: (JSONObject) -> Unit) = with(activity) {
        val title = item.getString("title")
        column().apply {
            addView(artwork.poster(item, api), LinearLayout.LayoutParams(width, (width * 1.5).roundToInt()))
            addWithGap(phoneText(title, 14).apply {
                setSingleLine(true)
                ellipsize = TextUtils.TruncateAt.END
            }, ui(9))
            contentDescription = title
            onTap { open(item) }
        }
    }

    private fun feature(item: JSONObject, api: Api, open: (JSONObject) -> Unit) = with(activity) {
        val title = item.getString("title")
        column().apply {
            background = panel(0xff202b46.toInt(), 22)
            clipToOutline = true
            addView(FrameLayout(this@with).apply {
                setBackgroundColor(0xff25394b.toInt())
                addView(artwork.image(item, api, blur = true).apply { alpha = .45f }, FrameLayout.LayoutParams(-1, -1))
                addView(artwork.poster(item, api, 8), FrameLayout.LayoutParams(ui(134), ui(200), Gravity.TOP or Gravity.CENTER_HORIZONTAL)
                    .apply { topMargin = ui(10) })
            }, LinearLayout.LayoutParams(-1, ui(220)))
            addView(phoneRow().apply {
                setPadding(ui(18), ui(18), ui(18), ui(18))
                addView(phoneText(title, 20, 0xffe0e7ff.toInt(), true).apply { maxLines = 2 },
                    LinearLayout.LayoutParams(0, -2, 1f))
                addView(phoneText("→", 24).apply {
                    background = panel(0xff3f517a.toInt(), 22)
                    gravity = Gravity.CENTER
                }, LinearLayout.LayoutParams(ui(44), ui(44)).apply { leftMargin = ui(14) })
            })
            onTap { open(item) }
            contentDescription = title
        }
    }

    fun library(
        body: LinearLayout, library: JSONObject?, options: List<JSONObject>, items: List<JSONObject>,
        api: Api, topInset: Int, select: (Long) -> Unit, settings: () -> Unit, open: (JSONObject) -> Unit,
    ) = with(activity) {
        val featured = items.firstOrNull()
        val heroHeight = ui((resources.configuration.screenHeightDp * .48).coerceIn(300.0, 430.0))
        body.addView(FrameLayout(this).apply {
            background = panel(0xff243c39.toInt(), 26).apply {
                val bottom = ui(26).toFloat()
                cornerRadii = floatArrayOf(0f, 0f, 0f, 0f, bottom, bottom, bottom, bottom)
            }
            clipToOutline = true
            if (featured != null) addView(artwork.image(featured, api).apply { alpha = .65f }, FrameLayout.LayoutParams(-1, -1))
            addView(View(this@with).apply { setBackgroundColor(0x8c081413.toInt()) },
                FrameLayout.LayoutParams(-1, (heroHeight * .68).roundToInt(), Gravity.BOTTOM))
            addView(phoneRow().apply {
                setPadding(ui(24), topInset + ui(12), ui(24), 0)
                addView(phoneText("D.", 30, 0xffd2eedb.toInt(), true).apply { contentDescription = "Dinosaur" },
                    LinearLayout.LayoutParams(0, -2, 1f))
                addView(avatar(settings), LinearLayout.LayoutParams(ui(44), ui(44)))
            }, FrameLayout.LayoutParams(-1, -2, Gravity.TOP))
            addView(column().apply {
                setPadding(ui(24), ui(24), ui(24), ui(24))
                addView(phoneText(featured?.getString("title") ?: "片库", 32, Color.WHITE, true).apply { maxLines = 3 })
                if (featured != null) addWithGap(phoneRow().apply {
                    addView(phoneRow().apply {
                        items.drop(1).take(3).forEach { item ->
                            addView(artwork.poster(item, api, 5), LinearLayout.LayoutParams(ui(36), ui(50)).apply {
                                if (childCount > 0) leftMargin = ui(8)
                            })
                        }
                    }, LinearLayout.LayoutParams(0, -2, 1f))
                    addView(phoneAction("查看", { open(featured) }, color = 0xff17213a.toInt(), fill = 0xffe3e8fc.toInt())
                        .apply { minWidth = ui(90) })
                }, ui(22))
            }, FrameLayout.LayoutParams(-1, -2, Gravity.BOTTOM))
        }, LinearLayout.LayoutParams(-1, heroHeight))
        body.addWithGap(libraryTabs(options, library?.optLong("id") ?: 0L, select), ui(24))
        if (items.size > 1) body.addWithGap(column().apply {
            setPadding(ui(24), 0, ui(24), ui(8))
            addView(phoneText(library?.optString("name").orEmpty(), 21, bold = true))
            items.drop(1).forEachIndexed { index, item ->
                addWithGap(phoneRow().apply {
                    minimumHeight = ui(122)
                    addView(artwork.poster(item, api), LinearLayout.LayoutParams(ui(68), ui(102)))
                    addView(phoneText(item.getString("title"), 17).apply { maxLines = 2 },
                        LinearLayout.LayoutParams(0, -2, 1f).apply { leftMargin = ui(18) })
                    addView(phoneText("›", 25, PHONE_MUTED))
                    onTap { open(item) }
                }, ui(if (index == 0) 20 else 8))
            }
        }, ui(24))
    }

    fun pagination(body: LinearLayout, total: Int, page: Int, move: (Int) -> Unit) = with(activity) {
        val pages = (total + 23) / 24
        if (pages <= 1) return@with
        body.addWithGap(phoneRow().apply {
            gravity = Gravity.CENTER
            addView(phoneAction("←", { move(page - 1) }).apply {
                contentDescription = "上一页"
                isEnabled = page > 1
                if (!isEnabled) alpha = .35f
            })
            addView(phoneText("$page / $pages", 14).apply { setPadding(ui(22), 0, ui(22), 0) })
            addView(phoneAction("→", { move(page + 1) }).apply {
                contentDescription = "下一页"
                isEnabled = page < pages
                if (!isEnabled) alpha = .35f
            })
        }, ui(22))
    }
}
