package com.monkeydog.dinosaur.mobile

import android.graphics.Bitmap
import android.graphics.BitmapFactory
import android.os.Build
import android.graphics.RenderEffect
import android.graphics.Shader
import android.util.LruCache
import android.view.Gravity
import android.widget.FrameLayout
import android.widget.ImageView
import com.monkeydog.dinosaur.android.Api
import org.json.JSONObject
import java.net.HttpURLConnection
import java.net.URL
import java.util.concurrent.Executors

internal class PhoneArtwork(private val activity: MainActivity) {
    private val workers = Executors.newFixedThreadPool(3)
    private val cache = object : LruCache<String, Bitmap>(20 * 1024 * 1024) {
        override fun sizeOf(key: String, value: Bitmap) = value.byteCount
    }

    fun poster(item: JSONObject, client: Api, radius: Int = 10) = with(activity) {
        val fallback = phoneText(item.optString("title").take(1), 32, PHONE_ACCENT, true).apply {
            gravity = Gravity.CENTER
        }
        FrameLayout(this).apply {
            background = panel(PHONE_SURFACE, radius)
            clipToOutline = radius > 0
            addView(fallback, FrameLayout.LayoutParams(-1, -1))
            addView(image(item, client), FrameLayout.LayoutParams(-1, -1))
        }
    }

    fun image(item: JSONObject, client: Api, blur: Boolean = false) = ImageView(activity).apply {
        scaleType = ImageView.ScaleType.CENTER_CROP
        importantForAccessibility = ImageView.IMPORTANT_FOR_ACCESSIBILITY_NO
        if (blur && Build.VERSION.SDK_INT >= 31) {
            val radius = activity.ui(18).toFloat()
            setRenderEffect(RenderEffect.createBlurEffect(radius, radius, Shader.TileMode.CLAMP))
        }
        val path = item.optString("poster_url").takeIf { it.isNotBlank() && it != "null" } ?: return@apply
        val account = activity.account?.key
        val url = runCatching { URL(client.resource(path).toString()) }.getOrNull() ?: return@apply
        val key = "$account:$url"
        val existing = cache.get(key)
        if (existing != null) {
            setImageBitmap(existing)
        } else {
            workers.execute {
                val bitmap = cache.get(key) ?: runCatching {
                    val connection = url.openConnection() as HttpURLConnection
                    try {
                        connection.instanceFollowRedirects = false
                        connection.connectTimeout = 5000
                        connection.readTimeout = 5000
                        client.headers.forEach(connection::setRequestProperty)
                        if (connection.responseCode == 200) connection.inputStream.use {
                            BitmapFactory.decodeStream(it, null, BitmapFactory.Options().apply { inSampleSize = 2 })
                        } else null
                    } finally {
                        connection.disconnect()
                    }
                }.getOrNull()?.also { cache.put(key, it) }
                if (bitmap != null) activity.runOnUiThread { setImageBitmap(bitmap) }
            }
        }
    }

    fun close() = workers.shutdownNow()
}
