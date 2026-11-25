package com.monkeydog.dinosaur.mobile

import android.os.Build
import android.view.Gravity
import android.view.View
import android.view.WindowInsets
import android.widget.FrameLayout
import android.widget.ImageView
import android.widget.LinearLayout
import android.widget.ScrollView
import com.monkeydog.dinosaur.android.json
import com.monkeydog.dinosaur.android.rows
import org.json.JSONArray

internal class Settings(private val activity: MainActivity) {
    private val prefs = activity.getSharedPreferences("native-home", 0)
    private fun visibilityKey() = "dinosaur.home-libraries." + requireNotNull(activity.account).key
    private fun hiddenIds(): Set<Long> {
        val key = visibilityKey()
        val raw = prefs.getString(key, null) ?: activity.vault.legacyValue(key) ?: "[]"
        val array = runCatching { JSONArray(raw) }.getOrElse { JSONArray() }
        return (0 until array.length()).map { array.optLong(it) }.toSet()
    }
    fun hidden(id: Long) = id in hiddenIds()

    fun show(): Unit = with(activity) {
        val current = requireNotNull(account)
        val username = current.user.getString("username")
        settingsPanel(false)
        content.addView(column().apply {
            gravity = Gravity.CENTER_HORIZONTAL
            setPadding(0, ui(16), 0, ui(16))
            addView(FrameLayout(this@with).apply {
                background = panel(0xff202125.toInt(), 52, PHONE_ACCENT)
                addView(phoneText(username.take(1).uppercase(), 38, 0xffc4ead1.toInt(), true).apply {
                    gravity = Gravity.CENTER
                    background = panel(0xff354a40.toInt(), 41)
                }, FrameLayout.LayoutParams(ui(82), ui(82), Gravity.CENTER))
            }, LinearLayout.LayoutParams(ui(96), ui(96)))
            addWithGap(phoneText(username, 25), ui(18), -2)
        })
        content.addWithGap(group().apply {
            addView(settingRow(R.drawable.settings_switch_account, "切换账号") { showAccounts(null) })
        }, ui(20))
        val links = group()
        renderLinks(links, canManage = false)
        content.addWithGap(links, ui(20))
        val version = packageManager.getPackageInfo(packageName, 0).versionName
        content.addWithGap(phoneText("Dinosaur $version", 13, PHONE_MUTED).apply { gravity = Gravity.CENTER }, ui(32))
        val client = api
        task({ client.request("/api/auth/me").getJSONObject("user") }, { }) { user ->
            if (user.getString("role") == "admin") renderLinks(links, canManage = true)
        }
    }

    private fun MainActivity.renderLinks(links: LinearLayout, canManage: Boolean) {
        links.removeAllViews()
        val rows = buildList {
            add(settingRow(R.drawable.settings_tv, "登录电视") { PairTv(activity) { catalog.show() }.show() })
            if (canManage) add(settingRow(R.drawable.settings_manage, "管理") { admin.show() })
            add(settingRow(R.drawable.settings_home, "首页显示") { libraries(false) })
            add(settingRow(R.drawable.settings_history, "记录观看") { libraries(true) })
        }
        rows.forEachIndexed { index, row ->
            if (index > 0) links.separator()
            links.addView(row)
        }
    }

    private fun MainActivity.settingsPanel(nested: Boolean) {
        generation++
        back = if (nested) ({ show() }) else ({ catalog.show() })
        val toolbar = phoneRow().apply {
            setPadding(ui(14), ui(8), ui(14), 0)
            val previous = if (nested) phoneAction("←", { show() }, size = 25, fill = 0).apply {
                contentDescription = "返回账号设置"
            } else android.view.View(this@settingsPanel)
            addView(previous, LinearLayout.LayoutParams(ui(48), ui(48)))
            addView(phoneText(account?.server?.optString("name").orEmpty(), 14).apply {
                gravity = Gravity.CENTER
                maxLines = 1
            }, LinearLayout.LayoutParams(0, -2, 1f))
            addView(phoneAction("×", { catalog.show() }, size = 28, color = PHONE_MUTED, fill = 0).apply {
                contentDescription = "关闭设置"
            }, LinearLayout.LayoutParams(ui(48), ui(48)))
        }
        val settingsContent = column().apply { setPadding(ui(18), 0, ui(18), ui(40)) }
        content = settingsContent
        setContentView(column().apply {
            fitsSystemWindows = Build.VERSION.SDK_INT < 30
            setBackgroundColor(PHONE_BG)
            setPadding(0, ui(10), 0, 0)
            if (Build.VERSION.SDK_INT >= 30) setOnApplyWindowInsetsListener { view, insets ->
                val bars = insets.getInsets(WindowInsets.Type.systemBars() or WindowInsets.Type.displayCutout())
                view.setPadding(bars.left, bars.top + ui(10), bars.right, 0)
                settingsContent.setPadding(ui(18), 0, ui(18), maxOf(ui(40), bars.bottom))
                WindowInsets.CONSUMED
            }
            addView(column().apply {
                background = panel(0xff202125.toInt(), 28).apply {
                    val corner = ui(28).toFloat()
                    cornerRadii = floatArrayOf(corner, corner, corner, corner, 0f, 0f, 0f, 0f)
                }
                addView(toolbar)
                addView(ScrollView(this@settingsPanel).apply { addView(settingsContent) }, LinearLayout.LayoutParams(-1, 0, 1f))
            }, LinearLayout.LayoutParams(-1, 0, 1f))
        })
    }

    private fun MainActivity.group() = column().apply {
        background = panel(0xff111215.toInt(), 26)
        clipToOutline = true
    }

    private fun MainActivity.divider() = View(this).apply {
        setBackgroundColor(0xff292a2e.toInt())
        layoutParams = LinearLayout.LayoutParams(-1, ui(1))
    }

    private fun LinearLayout.separator() = addView(activity.divider())

    private fun MainActivity.settingRow(icon: Int, title: String, action: () -> Unit) =
        phoneRow().apply {
            minimumHeight = ui(66)
            setPadding(ui(20), ui(12), ui(20), ui(12))
            addView(ImageView(this@settingRow).apply {
                setImageResource(icon)
                setColorFilter(PHONE_MUTED)
                scaleType = ImageView.ScaleType.FIT_CENTER
                importantForAccessibility = View.IMPORTANT_FOR_ACCESSIBILITY_NO
            }, LinearLayout.LayoutParams(ui(24), ui(24)))
            addView(phoneText(title, 16), LinearLayout.LayoutParams(0, -2, 1f).apply { leftMargin = ui(18) })
            addView(phoneText("›", 22, PHONE_MUTED))
            contentDescription = title
            onTap(action)
        }

    private fun libraries(record: Boolean) = with(activity) {
        settingsPanel(true)
        content.addView(phoneText(if (record) "记录观看" else "首页显示", 24).apply { setPadding(ui(8), ui(14), 0, ui(14)) })
        val group = group()
        content.addView(group)
        val client = api
        task({ client.request(if (record) "/api/history/settings" else "/api/libraries").rows("libraries") }) { items ->
            if (items.isEmpty()) label("暂无媒体库")
            items.forEachIndexed { index, item ->
                if (index > 0) group.separator()
                val id = item.getLong("id")
                group.addView(switchControl(item.getString("name"), if (record) item.optBoolean("record_history") else !hidden(id)) { value, control ->
                    if (record) {
                        control.isChecked = !value
                        val save = {
                            control.isEnabled = false
                            task({ client.request("/api/history/settings/$id", "PUT", json("record_history" to value)) }, {
                                control.isEnabled = true
                                message(it.message ?: "保存失败")
                            }) {
                                control.isEnabled = true
                                control.isChecked = value
                            }
                        }
                        if (value) save() else confirm("关闭记录并清理该媒体库的进度和自动观看记录？", save)
                    } else {
                        val current = hiddenIds()
                        val next = if (value) current - id else current + id
                        prefs.edit().putString(visibilityKey(), JSONArray(next.toList()).toString()).apply()
                    }
                })
            }
        }
    }
}
