package com.monkeydog.dinosaur.android

internal fun Api.historyPage(page: Int) = request("/api/history?page=$page&page_size=24")

internal fun Api.historySettings() = request("/api/history/settings").rows("libraries")

internal fun Api.setHistoryRecording(library: Long, record: Boolean) =
    request("/api/history/settings/$library", "PUT", json("record_history" to record))
