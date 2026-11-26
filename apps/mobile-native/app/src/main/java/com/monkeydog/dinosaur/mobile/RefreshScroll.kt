package com.monkeydog.dinosaur.mobile

import android.graphics.Rect
import android.view.MotionEvent
import android.view.View
import android.view.ViewGroup
import android.widget.HorizontalScrollView
import android.widget.ScrollView
import kotlin.math.abs

internal class RefreshScroll(private val activity: MainActivity) : ScrollView(activity) {
    var refresh: (() -> Unit)? = null
    var swipeLibrary: ((Int) -> Unit)? = null
    private var startY = 0f
    private var startX = 0f
    private var tracking = false
    private var trackingLibrary = false

    override fun dispatchTouchEvent(event: MotionEvent): Boolean {
        when (event.actionMasked) {
            MotionEvent.ACTION_DOWN -> {
                tracking = !canScrollVertically(-1)
                trackingLibrary = swipeLibrary != null && !touchesHorizontalList(this, event.rawX.toInt(), event.rawY.toInt())
                startY = event.y
                startX = event.x
            }
            MotionEvent.ACTION_POINTER_DOWN, MotionEvent.ACTION_CANCEL -> {
                tracking = false
                trackingLibrary = false
            }
            MotionEvent.ACTION_UP -> {
                val distance = event.y - startY
                val sideways = event.x - startX
                val switch = swipeLibrary.takeIf {
                    trackingLibrary && abs(sideways) > activity.dp(64) && abs(sideways) > abs(distance) * 1.5f
                }
                val perform = tracking && !canScrollVertically(-1) && distance > activity.dp(80) &&
                    distance > abs(sideways) * 1.5f
                tracking = false
                trackingLibrary = false
                if (switch != null || perform && refresh != null) {
                    val cancel = MotionEvent.obtain(event)
                    cancel.action = MotionEvent.ACTION_CANCEL
                    super.dispatchTouchEvent(cancel)
                    cancel.recycle()
                    if (switch != null) switch(if (sideways < 0) 1 else -1) else refresh?.invoke()
                    return true
                }
            }
        }
        return super.dispatchTouchEvent(event)
    }

    private fun touchesHorizontalList(view: View, x: Int, y: Int): Boolean {
        val bounds = Rect()
        if (!view.getGlobalVisibleRect(bounds) || !bounds.contains(x, y)) return false
        if (view is HorizontalScrollView) return true
        return view is ViewGroup && (0 until view.childCount).any { touchesHorizontalList(view.getChildAt(it), x, y) }
    }
}
