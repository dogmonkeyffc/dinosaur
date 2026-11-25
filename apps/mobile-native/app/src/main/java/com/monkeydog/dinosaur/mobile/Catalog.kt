package com.monkeydog.dinosaur.mobile

import android.os.Handler
import android.os.Looper
import android.view.inputmethod.InputMethodManager
import android.widget.LinearLayout
import com.monkeydog.dinosaur.android.Api
import com.monkeydog.dinosaur.android.rows
import org.json.JSONObject
import java.net.URLEncoder

internal class Catalog(private val activity: MainActivity) {
    private val artwork = PhoneArtwork(activity)
    private val feed = CatalogFeed(activity, artwork)
    private val timeline = HistoryTimeline(activity, artwork)
    private val details = MediaDetails(activity, artwork, ::play)
    private val handler = Handler(Looper.getMainLooper())
    private var pendingSearch: Runnable? = null
    private var owner: String? = null
    private var tab = 0
    private var query = ""
    private var searching = false
    private var pageNumber = 1
    private var savedScroll = 0
    private val selectedIds = LongArray(3)
    private var request = 0

    fun browseLibrary(id: Long) {
        selectedIds[1] = id
        tab = 1
        reset()
        show()
    }

    fun show(destination: Int = tab): Unit = with(activity) {
        val accountKey = requireNotNull(account).key
        if (accountKey != owner) {
            owner = accountKey
            tab = 0
            selectedIds.fill(0)
            reset()
            showChrome()
            return@with
        }
        if (destination != tab) {
            tab = destination
            reset()
        }
        showChrome()
    }

    private fun reset() {
        query = ""
        searching = false
        pageNumber = 1
        savedScroll = 0
    }

    private fun showChrome(): Unit = with(activity) {
        pendingSearch?.let(handler::removeCallbacks)
        val chrome = CatalogChrome(this, tab) { destination ->
            getSystemService(InputMethodManager::class.java).hideSoftInputFromWindow(currentFocus?.windowToken, 0)
            tab = destination
            reset()
            showChrome()
        }
        val client = api
        if (tab == 2) {
            timeline.show(chrome, client)
            return@with
        }
        val toolbar = column()
        if (tab == 0) {
            val (heading, input) = searchHeading(query, {
                savedScroll = chrome.scroller.scrollY
                settings.show()
            }) { value ->
                query = value
                pageNumber = 1
                savedScroll = 0
                val screen = generation
                request++
                pendingSearch?.let(handler::removeCallbacks)
                pendingSearch = Runnable { if (screen == generation) catalog(chrome, toolbar, client) }
                    .also { handler.postDelayed(it, 250) }
            }
            input.setOnFocusChangeListener { _, active ->
                if (active && !searching) {
                    searching = true
                    pageNumber = 1
                    savedScroll = 0
                    back = { reset(); showChrome() }
                    catalog(chrome, toolbar, client)
                }
            }
            input.setOnEditorActionListener { field, _, _ ->
                getSystemService(InputMethodManager::class.java).hideSoftInputFromWindow(field.windowToken, 0)
                true
            }
            chrome.header.addView(heading)
            chrome.header.addView(toolbar)
            if (searching) back = { reset(); showChrome() }
        }
        chrome.scroller.refresh = { savedScroll = 0; catalog(chrome, toolbar, client) }
        catalog(chrome, toolbar, client)
    }

    private fun catalog(chrome: CatalogChrome, toolbar: LinearLayout, client: Api): Unit = with(activity) {
        val revision = ++request
        val destination = tab
        val search = searching
        val selection = if (search) 2 else destination
        val wanted = selectedIds[selection].takeIf { it > 0 } ?: selectedIds[0]
        val typed = query.trim()
        val page = pageNumber
        val body = chrome.body
        body.removeAllViews()
        body.addView(loading())
        task({
            val libraries = client.request("/api/libraries").rows("libraries")
            val visible = if (destination == 0 && !search) libraries.filterNot { settings.hidden(it.getLong("id")) } else libraries
            val selected = visible.firstOrNull { it.optLong("id") == wanted } ?: visible.firstOrNull()
            val id = selected?.getLong("id") ?: 0
            val kind = selected?.optString("type").orEmpty()
            val route = if (kind == "tv") "series" else "media"
            val encoded = URLEncoder.encode(typed, "UTF-8")
            val data = if (selected == null) JSONObject() else
                client.request("/api/$route?type=$kind&library_id=$id&page=$page&page_size=24&q=$encoded")
            val recent = if (destination == 0 && !search && selected != null) {
                val ids = visible.map { it.getLong("id") }.toSet()
                runCatching { client.request("/api/playback/continue?type=$kind").rows("items") }
                    .getOrDefault(emptyList()).filter { it.optLong("library_id") in ids }
            } else emptyList()
            CatalogPage(visible, selected, data, recent, kind)
        }, { error ->
            if (revision == request) feedback(body, error.message ?: "无法加载片库") { catalog(chrome, toolbar, client) }
        }) { result ->
            if (revision != request) return@task
            val id = result.selected?.optLong("id") ?: 0L
            if (id > 0) selectedIds[selection] = id
            val select = { next: Long ->
                selectedIds[selection] = next
                pageNumber = 1
                savedScroll = 0
                catalog(chrome, toolbar, client)
            }
            val open: (JSONObject) -> Unit = { item ->
                savedScroll = chrome.scroller.scrollY
                getSystemService(InputMethodManager::class.java).hideSoftInputFromWindow(chrome.scroller.windowToken, 0)
                details.open(item, result.kind) { showChrome() }
            }
            toolbar.removeAllViews()
            body.removeAllViews()
            body.setPadding(0, 0, 0, ui(28))
            val items = result.data.rows("items")
            if (destination == 1) {
                feed.library(body, result.selected, result.libraries, items, client, chrome.topInset, select, {
                    savedScroll = chrome.scroller.scrollY
                    settings.show()
                }, open)
            } else {
                toolbar.addView(libraryTabs(result.libraries, id, select))
                if (search) feed.search(body, items, client, typed.isNotEmpty(), open)
                else if (result.selected != null) feed.home(body, result.selected.optString("name"), items, result.recent, client, {
                    tab = 1
                    reset()
                    selectedIds[1] = id
                    showChrome()
                }, open, {
                    savedScroll = chrome.scroller.scrollY
                    play(it.getLong("id")) { showChrome() }
                })
            }
            if (items.isEmpty()) body.addView(status(when {
                result.selected == null && destination == 0 && !search -> "暂无首页媒体库"
                result.selected == null -> "暂无媒体库"
                typed.isNotEmpty() -> "没有找到相关媒体"
                else -> "暂无媒体"
            }))
            if (destination == 1 || search) feed.pagination(body, result.data.optInt("total"), page) {
                pageNumber = it
                savedScroll = 0
                catalog(chrome, toolbar, client)
            }
            chrome.scroller.post { chrome.scroller.scrollTo(0, savedScroll) }
        }
    }

    private fun play(id: Long, close: () -> Unit) {
        activity.playback?.release()
        activity.playback = PlaybackScreen(activity, activity.api, close).apply { open(id) }
    }

    fun dispose() {
        pendingSearch?.let(handler::removeCallbacks)
        artwork.close()
    }

    private data class CatalogPage(
        val libraries: List<JSONObject>, val selected: JSONObject?, val data: JSONObject,
        val recent: List<JSONObject>, val kind: String,
    )
}
