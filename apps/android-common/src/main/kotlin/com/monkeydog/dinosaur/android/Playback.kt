package com.monkeydog.dinosaur.android

internal fun Api.playback(mediaId: Long, query: String = "") =
    request("/api/playback/$mediaId" + if (query.isEmpty()) "" else "?$query")

internal fun Api.continueWatching(type: String) = request("/api/playback/continue?type=$type").rows("items")

internal fun Api.savePlaybackProgress(mediaId: Long, seconds: Double, watched: Boolean) {
    val body = json("position" to seconds)
    if (watched) body.put("watched", true)
    request("/api/playback/$mediaId/progress", "PUT", body)
}
