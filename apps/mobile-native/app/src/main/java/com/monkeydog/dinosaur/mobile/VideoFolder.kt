package com.monkeydog.dinosaur.mobile

import android.text.TextUtils
import android.widget.LinearLayout
import com.monkeydog.dinosaur.android.Api
import com.monkeydog.dinosaur.android.rows
import org.json.JSONObject
import java.net.URLEncoder

internal data class VideoFile(val id: Long, val name: String)

internal data class VideoFolder(val path: String, val folders: List<String>, val videos: List<VideoFile>) {
    val total get() = folders.size + videos.size
}

internal fun Api.videoFolder(library: Long, path: String = "", mediaId: Long? = null): VideoFolder {
    val query = if (mediaId == null) "path=" + URLEncoder.encode(path, "UTF-8") else "media_id=$mediaId"
    val result = request("/api/libraries/$library/videos?$query")
    val folders = result.getJSONArray("folders")
    val videos = result.rows("videos").map { VideoFile(it.getLong("id"), it.getString("name")) }
    return VideoFolder(result.getString("path"), List(folders.length(), folders::getString), videos)
}

internal fun MainActivity.renderVideoFolder(
    body: LinearLayout, libraries: List<JSONObject>, selected: JSONObject, folder: VideoFolder,
    page: Int, inset: Int, select: (Long) -> Unit, settings: () -> Unit, directory: (String) -> Unit, play: (Long) -> Unit,
) {
    val title = selected.getString("name")
    body.addView(phoneRow().apply {
        setPadding(ui(24), inset + ui(12), ui(24), ui(10))
        addView(phoneText(title, 24, bold = true), LinearLayout.LayoutParams(0, -2, 1f))
        addView(avatar(settings), LinearLayout.LayoutParams(ui(44), ui(44)))
    })
    body.addView(libraryTabs(libraries, selected.getLong("id"), select))
    body.addView(column().apply {
        setPadding(ui(24), ui(16), ui(24), 0)
        if (folder.path.isNotEmpty()) addView(phoneRow().apply {
            val parent = folder.path.substringBeforeLast('/', "")
            addView(phoneAction("‹", { directory(parent) }, fill = PHONE_SELECTED).apply { contentDescription = "上一级目录" })
            addView(phoneText(folder.path, 15, PHONE_MUTED).apply {
                maxLines = 2
                ellipsize = TextUtils.TruncateAt.START
            }, LinearLayout.LayoutParams(0, -2, 1f).apply { leftMargin = ui(12) })
        })
        val first = ((page - 1) * 24).coerceIn(0, folder.total)
        for (index in first until minOf(first + 24, folder.total)) {
            if (index < folder.folders.size) {
                val name = folder.folders[index]
                val path = listOf(folder.path, name).filter(String::isNotEmpty).joinToString("/")
                addWithGap(directoryRow("folder", name) { directory(path) }, ui(10))
            } else {
                val video = folder.videos[index - folder.folders.size]
                addWithGap(directoryRow("video", video.name) { play(video.id) }, ui(10))
            }
        }
        if (folder.total == 0) addWithGap(status("此目录为空"), ui(10))
    })
}

private fun MainActivity.directoryRow(icon: String, title: String, open: () -> Unit) = phoneRow().apply {
    minimumHeight = ui(58)
    setPadding(ui(16), ui(10), ui(16), ui(10))
    background = panel(PHONE_SURFACE, 18)
    addView(CatalogIcon(this@directoryRow, icon, PHONE_ACCENT), LinearLayout.LayoutParams(ui(24), ui(24)))
    addView(phoneText(title, 16).apply { maxLines = 2 }, LinearLayout.LayoutParams(0, -2, 1f)
        .apply { leftMargin = ui(14) })
    val action = if (icon == "folder") "打开文件夹" else "播放"
    contentDescription = "$action $title"
    onTap(open)
}
