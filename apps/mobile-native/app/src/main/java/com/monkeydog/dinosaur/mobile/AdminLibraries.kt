package com.monkeydog.dinosaur.mobile

import android.view.Gravity
import android.view.View
import android.widget.PopupMenu
import android.widget.ScrollView
import com.monkeydog.dinosaur.android.json
import com.monkeydog.dinosaur.android.rows
import org.json.JSONArray
import org.json.JSONObject

private const val LIBRARIES_PATH = "/api/admin/libraries"

internal class AdminLibraries(private val activity: MainActivity) {
    private lateinit var list: AdminLibraryList
    private lateinit var scroller: ScrollView
    private val pendingScans = mutableSetOf<Long>()

    fun show() = showList("", "", 0)

    private fun showList(type: String, search: String, position: Int): Unit = with(activity) {
        page("媒体库设置", { settings.show() })
        content.addView(phoneRow().apply {
            gravity = Gravity.END or Gravity.CENTER_VERTICAL
            addView(phoneAction("添加媒体库", { edit(null) }, color = PHONE_SELECTED, fill = PHONE_ACCENT))
        })
        scroller = content.parent as ScrollView
        list = AdminLibraryList(activity, type, search, ::menu)
        pendingScans.clear()
        refresh(position)
    }

    private fun returnToList(): () -> Unit {
        val type = list.type
        val search = list.search
        val position = scroller.scrollY
        return { showList(type, search, position) }
    }

    private fun refresh(position: Int = scroller.scrollY) = with(activity) {
        val client = api
        val view = scroller
        task({ client.request(LIBRARIES_PATH).rows("libraries") }) { rows ->
            list.setRows(rows)
            view.post { view.scrollTo(0, position) }
        }
    }

    private fun menu(anchor: View, item: JSONObject) = with(activity) {
        val id = item.getLong("id")
        val popup = PopupMenu(this, anchor, Gravity.END)
        fun option(title: String, enabled: Boolean = true, action: () -> Unit) {
            popup.menu.add(title).apply {
                isEnabled = enabled
                setOnMenuItemClickListener {
                    action()
                    true
                }
            }
        }
        val active = item.scanActive() || id in pendingScans
        option("编辑") { edit(item) }
        option("访问权限") { access(id, returnToList()) }
        option(if (active) "正在扫描…" else "重新扫描", !active) { scan(id) }
        option("刷新扫描状态") { refresh() }
        option("移除媒体库") { remove(item) }
        popup.show()
    }

    private fun edit(item: JSONObject?) {
        val previous = returnToList()
        LibraryEditor(activity, item, previous) { previous() }.show()
    }

    private fun scan(id: Long): Unit = with(activity) {
        if (!pendingScans.add(id)) return@with
        val client = api
        task({ client.request("$LIBRARIES_PATH/$id/scan", "POST", JSONObject()).getJSONObject("scan") }, {
            pendingScans.remove(id)
            message(it.message ?: "无法开始扫描")
            refresh()
        }) { job ->
            list.setScan(id, job)
            pendingScans.remove(id)
            message("已开始扫描")
            refresh()
        }
    }

    private fun remove(item: JSONObject) = with(activity) {
        val id = item.getLong("id")
        confirm("移除媒体库 ${item.getString("name")}？") {
            val client = api
            task({ client.request("$LIBRARIES_PATH/$id", "DELETE") }) { refresh() }
        }
    }

    private fun access(id: Long, previous: () -> Unit): Unit = with(activity) {
        page("访问权限", previous)
        val client = api
        val path = "$LIBRARIES_PATH/$id/access"
        task({ client.request(path) }) { data ->
            val current = data.getJSONObject("access")
            var restricted = current.getBoolean("restricted")
            val ids = current.optJSONArray("user_ids") ?: JSONArray()
            val selected = (0 until ids.length()).map(ids::getLong).toMutableSet()
            rowSwitch("限制访问", restricted) { value, _ -> restricted = value }
            data.rows("accounts").forEach { item ->
                val userId = item.getLong("id")
                rowSwitch(item.getString("username"), userId in selected) { enabled, _ ->
                    if (enabled) selected.add(userId) else selected.remove(userId)
                }
            }
            button("保存") {
                val body = json("restricted" to restricted, "user_ids" to JSONArray(selected.toList()))
                task({ client.request(path, "PUT", body) }) { previous() }
            }
        }
    }
}
