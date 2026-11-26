package com.monkeydog.dinosaur.mobile

import android.view.MotionEvent
import android.view.View
import android.view.ViewConfiguration
import kotlin.math.abs

internal fun MainActivity.videoSwipe(view: View, switch: (Int) -> Unit) {
    val slop = ViewConfiguration.get(this).scaledTouchSlop
    var x = 0f
    var y = 0f
    var active = false
    view.setOnTouchListener { v, event ->
        when (event.actionMasked) {
            MotionEvent.ACTION_DOWN -> {
                x = event.x
                y = event.y
                active = true
            }
            MotionEvent.ACTION_CANCEL, MotionEvent.ACTION_POINTER_DOWN -> active = false
            MotionEvent.ACTION_UP -> if (active) {
                active = false
                val dx = abs(event.x - x)
                val dy = event.y - y
                if (abs(dy) >= dp(64) && abs(dy) > dx * 1.4f) switch(if (dy < 0) 1 else -1)
                else if (dx <= slop && abs(dy) <= slop) v.performClick()
            }
        }
        true
    }
}
