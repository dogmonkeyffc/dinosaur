package com.monkeydog.dinosaur.mobile

import java.net.URLEncoder

internal class DirectoryPicker(private val activity: MainActivity) {
    fun show(path: String, after: String = "", previous: () -> Unit, select: (String) -> Unit): Unit = with(activity) {
        page("服务器目录", previous)
        val client = api
        val query = "path=${URLEncoder.encode(path, "UTF-8")}&after=${URLEncoder.encode(after, "UTF-8")}"
        task({ client.request("/api/admin/libraries/directories?$query") }) { data ->
            val current = data.getString("path")
            val parent = data.optString("parent")
            label(current)
            button("使用这个目录") { select(current) }
            if (parent.isNotEmpty()) button("上一级") { show(parent, previous = previous, select = select) }
            data.rows("folders").forEach { folder ->
                button(folder.getString("name")) { show(folder.getString("path"), previous = previous, select = select) }
            }
            data.optString("next_after").takeIf(String::isNotBlank)?.let { next ->
                button("下一页") { show(current, next, previous, select) }
            }
            if (after.isNotEmpty()) button("回到第一页") { show(current, previous = previous, select = select) }
        }
    }
}
