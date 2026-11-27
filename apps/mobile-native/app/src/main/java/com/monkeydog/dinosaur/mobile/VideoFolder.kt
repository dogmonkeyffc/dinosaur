package com.monkeydog.dinosaur.mobile

import android.text.TextUtils
import android.widget.LinearLayout
import com.monkeydog.dinosaur.android.Api
import com.monkeydog.dinosaur.android.VideoFile
import com.monkeydog.dinosaur.android.VideoFolder
import org.json.JSONObject

internal fun MainActivity.renderVideoFolder(
    body: LinearLayout, libraries: List<JSONObject>, selected: JSONObject, folder: VideoFolder,
    page: Int, inset: Int, artwork: PhoneArtwork, client: Api, select: (Long) -> Unit, settings: () -> Unit, directory: (String) -> Unit, play: (Long) -> Unit,
) {
    if (folder.path.isEmpty()) {
        body.addView(phoneRow().apply {
            setPadding(ui(24), inset + ui(12), ui(24), ui(10))
            addView(phoneText(selected.getString("name"), 24, bold = true), LinearLayout.LayoutParams(0, -2, 1f))
            addView(avatar(settings), LinearLayout.LayoutParams(ui(44), ui(44)))
        })
        body.addView(libraryTabs(libraries, selected.getLong("id"), select))
    }
    body.addView(column().apply {
        setPadding(ui(24), ui(16), ui(24), 0)
        val first = ((page - 1) * 24).coerceIn(0, folder.total)
        for (index in first until minOf(first + 24, folder.total)) {
            if (index < folder.folders.size) {
                val name = folder.folders[index]
                val path = listOf(folder.path, name).filter(String::isNotEmpty).joinToString("/")
                addWithGap(folderRow(name) { directory(path) }, ui(10))
            } else {
                val video = folder.videos[index - folder.folders.size]
                addWithGap(videoCard(video, artwork, client) { play(video.id) }, ui(10))
            }
        }
        if (folder.total == 0) addWithGap(status("此目录为空"), ui(10))
    })
}

internal fun MainActivity.videoFolderHeading(path: String, inset: Int, up: () -> Unit) = phoneRow().apply {
    setPadding(ui(12), inset + ui(8), ui(24), ui(8))
    addView(PlayerAction(this@videoFolderHeading, PlayerIcon.BACK, "上一级目录", up),
        LinearLayout.LayoutParams(ui(48), ui(48)))
    addView(phoneText(path.substringAfterLast('/'), 24, bold = true).apply {
        maxLines = 1
        ellipsize = TextUtils.TruncateAt.END
    }, LinearLayout.LayoutParams(0, -2, 1f).apply { leftMargin = ui(8) })
}

private fun MainActivity.folderRow(title: String, open: () -> Unit) = phoneRow().apply {
    minimumHeight = ui(58)
    setPadding(ui(16), ui(10), ui(16), ui(10))
    background = panel(PHONE_SURFACE, 18)
    addView(CatalogIcon(this@folderRow, "folder", PHONE_ACCENT), LinearLayout.LayoutParams(ui(24), ui(24)))
    addView(phoneText(title, 16).apply { maxLines = 2 }, LinearLayout.LayoutParams(0, -2, 1f)
        .apply { leftMargin = ui(14) })
    contentDescription = "打开文件夹 $title"
    onTap(open)
}

private fun MainActivity.videoCard(video: VideoFile, artwork: PhoneArtwork, client: Api, play: () -> Unit) = phoneRow().apply {
    background = panel(PHONE_SURFACE, 18)
    setPadding(ui(8), ui(8), ui(16), ui(8))
    addView(artwork.preview(video.preview, client), LinearLayout.LayoutParams(ui(140), ui(79)))
    addView(phoneText(video.name, 16).apply {
        maxLines = 3
        ellipsize = TextUtils.TruncateAt.END
    }, LinearLayout.LayoutParams(0, -2, 1f).apply { leftMargin = ui(14) })
    contentDescription = "播放 ${video.name}"
    onTap(play)
}
