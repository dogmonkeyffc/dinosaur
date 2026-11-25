package com.monkeydog.dinosaur.mobile

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
        page("账号与设置", { catalog.show() })
        label(requireNotNull(account).user.getString("username"))
        button("切换账号") { showAccounts() }
        button("登录电视") { pairTv() }
        button("首页显示") { libraries(false) }
        button("记录观看") { libraries(true) }
        val client = api
        task({ client.request("/api/auth/me").getJSONObject("user") }, { }) { user ->
            if (user.getString("role") == "admin") button("管理") { admin.show() }
        }
    }

    private fun pairTv() = with(activity) {
        form("登录电视", listOf("电视六位配对码")) { values ->
            val client = api
            task({ client.request("/api/devices/pairings/lookup", "POST", json("code" to values[0])) }) { pairing ->
                confirm("授权 ${pairing.getString("name")} 登录？") {
                    task({ client.request("/api/devices/pairings/approve", "POST", json("pairing_id" to pairing.getString("pairing_id"))) }) {
                        message("已授权")
                    }
                }
            }
        }
    }

    private fun libraries(record: Boolean) = with(activity) {
        page(if (record) "记录观看" else "首页显示", { show() })
        val client = api
        task({ client.request(if (record) "/api/history/settings" else "/api/libraries").rows("libraries") }) { items ->
            if (items.isEmpty()) label("暂无媒体库")
            items.forEach { item ->
                val id = item.getLong("id")
                rowSwitch(item.getString("name"), if (record) item.optBoolean("record_history") else !hidden(id)) { value, control ->
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
                }
            }
        }
    }
}
