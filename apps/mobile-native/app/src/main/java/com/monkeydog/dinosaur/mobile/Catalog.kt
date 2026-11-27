package com.monkeydog.dinosaur.mobile

import android.os.Handler
import android.os.Looper
import android.view.inputmethod.InputMethodManager
import android.widget.LinearLayout
import com.monkeydog.dinosaur.android.VideoFile
import com.monkeydog.dinosaur.android.videoFolder
import com.monkeydog.dinosaur.android.Api
import com.monkeydog.dinosaur.android.rows
import org.json.JSONObject

internal class Catalog(private val activity: MainActivity) {
    private val artwork = PhoneArtwork(activity)
    private val feed = CatalogFeed(activity, artwork)
    private val timeline = HistoryTimeline(activity, artwork)
    private val details = MediaDetails(activity, artwork) { id, close -> play(id, close, landscape = true) }
    private val handler = Handler(Looper.getMainLooper())
    private var pendingSearch: Runnable? = null
    private var owner: String? = null
    private var tab = 0
    private var query = ""
    private var searching = false
    private var pageNumber = 1
    private var savedScroll = 0
    private var directory = ""
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
        directory = ""
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
        chrome.scroller.swipeLibrary = null
        body.removeAllViews()
        body.addView(loading())
        val path = directory
        task({
            client.catalogPage(destination == 1, search, wanted, typed, page, path, settings::hidden)
        }, { error ->
            if (revision == request) feedback(body, error.message ?: "无法加载片库") { catalog(chrome, toolbar, client) }
        }) { result ->
            if (revision != request) return@task
            val id = result.selected?.optLong("id") ?: 0L
            if (id > 0) selectedIds[selection] = id
            val select = { next: Long ->
                if (next != id) {
                    selectedIds[selection] = next
                    pageNumber = 1
                    savedScroll = 0
                    directory = ""
                    catalog(chrome, toolbar, client)
                }
            }
            val index = result.libraries.indexOfFirst { it.optLong("id") == id }
            if (!search && index >= 0) chrome.scroller.swipeLibrary = { offset ->
                result.libraries.getOrNull(index + offset)?.let { select(it.getLong("id")) }
            }
            val open: (JSONObject) -> Unit = { item ->
                savedScroll = chrome.scroller.scrollY
                getSystemService(InputMethodManager::class.java).hideSoftInputFromWindow(chrome.scroller.windowToken, 0)
                if (result.kind == "video") playVideo(item, client) else details.open(item, result.kind) { showChrome() }
            }
            toolbar.removeAllViews()
            body.removeAllViews()
            body.setPadding(0, 0, 0, ui(28))
            val items = result.data.rows("items")
            val folder = result.folder
            if (folder != null && result.selected != null) {
                directory = folder.path
                val openFolder = { next: String ->
                    directory = next
                    pageNumber = 1
                    savedScroll = 0
                    catalog(chrome, toolbar, client)
                }
                back = if (folder.path.isEmpty()) ({ show(0) }) else ({ openFolder(folder.path.substringBeforeLast('/', "")) })
                renderVideoFolder(body, result.libraries, result.selected, folder, page, chrome.topInset, select, {
                    savedScroll = chrome.scroller.scrollY
                    settings.show()
                }, openFolder) { mediaId ->
                    savedScroll = chrome.scroller.scrollY
                    play(mediaId, { showChrome() }, videos = folder.videos.map(VideoFile::id))
                }
            } else if (destination == 1) {
                back = { show(0) }
                feed.library(body, result.selected, result.libraries, items, client, chrome.topInset, select, {
                    savedScroll = chrome.scroller.scrollY
                    settings.show()
                }, open)
            } else {
                toolbar.addView(libraryTabs(result.libraries, id, select))
                if (search) feed.search(body, items, client, typed.isNotEmpty(), open)
                else if (result.selected != null) feed.home(body, result.selected.optString("name"), items, result.recent, client, open, {
                    savedScroll = chrome.scroller.scrollY
                    if (result.kind == "video") playVideo(it, client)
                    else play(it.getLong("id"), { showChrome() }, landscape = true)
                })
            }
            if (items.isEmpty() && folder == null) body.addView(status(when {
                result.selected == null && destination == 0 && !search -> "暂无首页媒体库"
                result.selected == null -> "暂无媒体库"
                typed.isNotEmpty() -> "没有找到相关媒体"
                else -> "暂无媒体"
            }))
            if (destination == 1 || search) feed.pagination(body, folder?.total ?: result.data.optInt("total"), page) {
                pageNumber = it
                savedScroll = 0
                catalog(chrome, toolbar, client)
            }
            chrome.scroller.post { chrome.scroller.scrollTo(0, savedScroll) }
        }
    }

    private fun playVideo(item: JSONObject, client: Api) = with(activity) {
        val id = item.getLong("id")
        val library = item.getLong("library_id")
        task({ client.videoFolder(library, mediaId = id).videos.map(VideoFile::id) }) { videos ->
            play(id, { showChrome() }, videos = videos)
        }
    }

    private fun play(id: Long, close: () -> Unit, landscape: Boolean = false, videos: List<Long> = emptyList()) {
        activity.playback?.release()
        activity.playback = PlaybackScreen(activity, activity.api, landscape, videos, close).apply { open(id) }
    }

    fun dispose() {
        pendingSearch?.let(handler::removeCallbacks)
        artwork.close()
    }

}
