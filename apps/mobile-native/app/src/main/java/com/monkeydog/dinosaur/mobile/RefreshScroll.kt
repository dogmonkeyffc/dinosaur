package com.monkeydog.dinosaur.mobile

import android.view.MotionEvent
import android.widget.ScrollView
import kotlin.math.abs

internal class RefreshScroll(private val activity: MainActivity) : ScrollView(activity) {
    var refresh: (() -> Unit)? = null
    private var startY = 0f
    private var startX = 0f
    private var tracking = false

    override fun dispatchTouchEvent(event: MotionEvent): Boolean {
        when (event.actionMasked) {
            MotionEvent.ACTION_DOWN -> {
                tracking = !canScrollVertically(-1)
                startY = event.y
                startX = event.x
            }
            MotionEvent.ACTION_POINTER_DOWN, MotionEvent.ACTION_CANCEL -> tracking = false
            MotionEvent.ACTION_UP -> {
                val distance = event.y - startY
                val perform = tracking && !canScrollVertically(-1) && distance > activity.dp(80) &&
                    distance > abs(event.x - startX) * 1.5f
                tracking = false
                if (perform && refresh != null) {
                    val cancel = MotionEvent.obtain(event)
                    cancel.action = MotionEvent.ACTION_CANCEL
                    super.dispatchTouchEvent(cancel)
                    cancel.recycle()
                    refresh?.invoke()
                    return true
                }
            }
        }
        return super.dispatchTouchEvent(event)
    }
}
