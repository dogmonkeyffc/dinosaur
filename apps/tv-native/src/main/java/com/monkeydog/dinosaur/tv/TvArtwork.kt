package com.monkeydog.dinosaur.tv

import android.graphics.Bitmap
import android.graphics.BitmapFactory
import android.text.TextUtils
import android.util.LruCache
import android.widget.ImageView
import android.widget.LinearLayout
import com.monkeydog.dinosaur.android.Api
import java.util.concurrent.Executors
import okhttp3.Request
import org.json.JSONObject

internal class TvArtwork(private val activity: TvActivity) {
    private val worker = Executors.newFixedThreadPool(3)
    private val cache = object : LruCache<String, Bitmap>(12 * 1024 * 1024) {
        override fun sizeOf(key: String, value: Bitmap) = value.byteCount
    }

    fun tile(item: JSONObject, api: Api, activate: () -> Unit): LinearLayout = with(activity) {
        val title = item.optString("title").ifBlank { item.optString("name") }
        val image = ImageView(this).apply {
            scaleType = ImageView.ScaleType.CENTER_CROP
            setBackgroundColor(surface)
        }
        val path = item.optString("poster_url")
        if (path.isNotBlank()) load(api, path, image)
        column().apply {
            background = focusSurface()
            isFocusable = true
            isFocusableInTouchMode = true
            contentDescription = title
            setPadding(dp(4), dp(4), dp(4), dp(8))
            addView(image, LinearLayout.LayoutParams(-1, dp(156)))
            addView(label(title, 15f, bold = true).apply {
                importantForAccessibility = android.view.View.IMPORTANT_FOR_ACCESSIBILITY_NO
                setPadding(dp(4), dp(7), dp(4), 0)
                maxLines = 2
                ellipsize = TextUtils.TruncateAt.END
                minHeight = dp(44)
            })
            setOnClickListener { activate() }
        }
    }

    private fun load(api: Api, path: String, view: ImageView) {
        val key = api.origin + path
        cache.get(key)?.let { view.setImageBitmap(it); return }
        worker.execute {
            val bitmap = runCatching {
                cache.get(key) ?: read(api, path)?.also { cache.put(key, it) }
            }.getOrNull()
            if (bitmap != null) activity.main.post { view.setImageBitmap(bitmap) }
        }
    }

    private fun read(api: Api, path: String): Bitmap? {
        val request = Request.Builder().url(api.resource(path).toString())
        api.headers.forEach { (name, value) -> request.header(name, value) }
        return Api.http.newCall(request.build()).execute().use { response ->
            val body = response.body ?: return null
            if (!response.isSuccessful) return null
            val bytes = body.source().apply { request(6L * 1024 * 1024 + 1) }
            if (bytes.buffer.size > 6L * 1024 * 1024) return null
            val data = bytes.readByteArray()
            val dimensions = BitmapFactory.Options().apply { inJustDecodeBounds = true }
            BitmapFactory.decodeByteArray(data, 0, data.size, dimensions)
            var scale = 1
            while (maxOf(dimensions.outWidth, dimensions.outHeight) / scale > 560) scale *= 2
            BitmapFactory.decodeByteArray(data, 0, data.size, BitmapFactory.Options().apply { inSampleSize = scale })
        }
    }

    fun close() {
        worker.shutdownNow()
        cache.evictAll()
    }
}
