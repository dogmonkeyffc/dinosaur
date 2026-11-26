package com.monkeydog.dinosaur.mobile

import com.monkeydog.dinosaur.android.Api
import com.monkeydog.dinosaur.android.rows
import org.json.JSONObject
import java.net.URLEncoder

internal data class CatalogPage(
    val libraries: List<JSONObject>, val selected: JSONObject?, val data: JSONObject,
    val recent: List<JSONObject>, val kind: String, val folder: VideoFolder?,
)

internal fun Api.catalogPage(
    browse: Boolean, search: Boolean, wanted: Long, typed: String, page: Int, directory: String,
    hidden: (Long) -> Boolean,
): CatalogPage {
    val libraries = request("/api/libraries").rows("libraries")
    val visible = if (browse || search) libraries else libraries.filterNot { hidden(it.getLong("id")) }
    val selected = visible.firstOrNull { it.optLong("id") == wanted } ?: visible.firstOrNull()
        ?: return CatalogPage(visible, null, JSONObject(), emptyList(), "", null)
    val id = selected.getLong("id")
    val kind = selected.getString("type")
    if (browse && kind == "video") {
        val path = if (id == wanted) directory else ""
        return CatalogPage(visible, selected, JSONObject(), emptyList(), kind, videoFolder(id, path))
    }
    val route = if (kind == "tv") "series" else "media"
    val encoded = URLEncoder.encode(typed, "UTF-8")
    val data = request("/api/$route?type=$kind&library_id=$id&page=$page&page_size=24&q=$encoded")
    val recent = if (browse || search) emptyList() else {
        val ids = visible.map { it.getLong("id") }.toSet()
        runCatching { request("/api/playback/continue?type=$kind").rows("items") }
            .getOrDefault(emptyList()).filter { it.optLong("library_id") in ids }
    }
    return CatalogPage(visible, selected, data, recent, kind, null)
}
