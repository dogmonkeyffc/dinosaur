package com.monkeydog.dinosaur.tv

import android.widget.LinearLayout
import com.monkeydog.dinosaur.android.Api
import com.monkeydog.dinosaur.android.VideoFolder
import com.monkeydog.dinosaur.android.videoFolder
import org.json.JSONObject

internal class TvVideoFolders(private val activity: TvActivity, private val api: Api) {
    fun library(library: JSONObject) = show(library, "")

    fun play(page: TvPage, item: JSONObject) = with(activity) {
        val id = item.getLong("id")
        task(page, { api.videoFolder(item.getLong("library_id"), mediaId = id) }) { folder ->
            play(folder, id)
        }
    }

    private fun play(folder: VideoFolder, id: Long) {
        TvPlayback(activity, api, folder.videos.map { it.id }).show(id)
    }

    private fun show(library: JSONObject, path: String): Unit = with(activity) {
        val title = if (path.isEmpty()) library.getString("name") else path.substringAfterLast('/')
        val (page, content) = panelPage(title)
        val back = control("返回", ::back)
        val listing = column()
        content.addControl(back)
        content.addView(listing)
        page.focus = back
        show(page)

        fun load() {
            listing.removeAllViews()
            val status = label("正在加载", color = muted)
            listing.addView(status)
            task(page, { api.videoFolder(library.getLong("id"), path) }, {
                status.text = it.message ?: "无法加载目录"
                listing.addControl(control("重试", ::load))
            }) { render(library, listing, it, 1) }
        }
        load()
    }

    private fun render(library: JSONObject, listing: LinearLayout, folder: VideoFolder, page: Int): Unit = with(activity) {
        listing.removeAllViews()
        if (folder.total == 0) listing.addView(label("此目录为空", color = muted))
        val first = (page - 1) * 24
        for (index in first until minOf(first + 24, folder.total)) {
            if (index < folder.folders.size) {
                val name = folder.folders[index]
                val path = if (folder.path.isEmpty()) name else "${folder.path}/$name"
                listing.addControl(control(name) { show(library, path) }.apply {
                    contentDescription = "打开文件夹 $name"
                    setCompoundDrawablesRelativeWithIntrinsicBounds(R.drawable.ic_tv_folder, 0, 0, 0)
                    compoundDrawablePadding = dp(12)
                })
            } else {
                val video = folder.videos[index - folder.folders.size]
                listing.addControl(control(video.name) { play(folder, video.id) }.apply {
                    contentDescription = "播放 ${video.name}"
                    setCompoundDrawablesRelativeWithIntrinsicBounds(android.R.drawable.ic_media_play, 0, 0, 0)
                    compoundDrawablePadding = dp(12)
                })
            }
        }
        val pages = (folder.total + 23) / 24
        if (pages > 1) listing.addView(row().apply {
            fun go(number: Int) {
                render(library, listing, folder, number)
                listing.getChildAt(0)?.requestFocus()
            }
            addControl(control("上一页") { go(page - 1) }.apply { isEnabled = page > 1 }, horizontal = true)
            addControl(control("下一页") { go(page + 1) }.apply { isEnabled = page < pages }, horizontal = true)
            addView(label("$page / $pages", color = muted))
        })
    }
}
