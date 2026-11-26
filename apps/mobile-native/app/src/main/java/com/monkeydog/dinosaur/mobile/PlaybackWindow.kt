package com.monkeydog.dinosaur.mobile

import android.content.pm.ActivityInfo
import android.graphics.Color
import android.os.Build
import android.view.View
import android.view.WindowInsets
import android.view.WindowInsetsController
import android.widget.FrameLayout

internal fun MainActivity.playbackRoot(fullscreen: () -> Boolean) = FrameLayout(this).apply {
    setBackgroundColor(Color.BLACK)
    fitsSystemWindows = Build.VERSION.SDK_INT < 30 && !fullscreen()
    if (Build.VERSION.SDK_INT >= 30) setOnApplyWindowInsetsListener { view, insets ->
        val bars = insets.getInsets(WindowInsets.Type.statusBars() or WindowInsets.Type.displayCutout())
        if (fullscreen()) view.setPadding(0, 0, 0, 0) else view.setPadding(bars.left, bars.top, bars.right, 0)
        WindowInsets.CONSUMED
    }
}

internal fun MainActivity.setPlaybackFullscreen(fullscreen: Boolean, previousOrientation: Int) {
    requestedOrientation = if (fullscreen) ActivityInfo.SCREEN_ORIENTATION_SENSOR_LANDSCAPE
    else previousOrientation
    if (Build.VERSION.SDK_INT >= 30) {
        window.insetsController?.apply {
            systemBarsBehavior = WindowInsetsController.BEHAVIOR_SHOW_TRANSIENT_BARS_BY_SWIPE
            if (fullscreen) hide(WindowInsets.Type.systemBars()) else show(WindowInsets.Type.systemBars())
        }
    } else {
        @Suppress("DEPRECATION")
        window.decorView.systemUiVisibility = if (fullscreen)
            View.SYSTEM_UI_FLAG_FULLSCREEN or View.SYSTEM_UI_FLAG_HIDE_NAVIGATION or View.SYSTEM_UI_FLAG_IMMERSIVE_STICKY
        else View.SYSTEM_UI_FLAG_VISIBLE
    }
    findViewById<FrameLayout>(android.R.id.content).getChildAt(0)?.apply {
        if (Build.VERSION.SDK_INT < 30) fitsSystemWindows = !fullscreen
        if (fullscreen) setPadding(0, 0, 0, 0)
        requestApplyInsets()
    }
}
