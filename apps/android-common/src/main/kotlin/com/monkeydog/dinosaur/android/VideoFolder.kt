package com.monkeydog.dinosaur.android

import java.net.URLEncoder

internal data class VideoFile(val id: Long, val name: String, val preview: String)

internal data class VideoFolder(val path: String, val folders: List<String>, val videos: List<VideoFile>) {
    val total get() = folders.size + videos.size
}

internal fun Api.videoFolder(library: Long, path: String = "", mediaId: Long? = null): VideoFolder {
    val query = if (mediaId == null) "path=" + URLEncoder.encode(path, "UTF-8") else "media_id=$mediaId"
    val result = request("/api/libraries/$library/videos?$query")
    val folders = result.getJSONArray("folders")
    val videos = result.rows("videos").map {
        VideoFile(it.getLong("id"), it.getString("name"), it.optString("poster_url"))
    }
    return VideoFolder(result.getString("path"), List(folders.length(), folders::getString), videos)
}
