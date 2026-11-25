package com.monkeydog.dinosaur.mobile

import android.app.Activity
import android.hardware.Camera
import android.view.SurfaceHolder
import android.view.SurfaceView
import com.google.zxing.BinaryBitmap
import com.google.zxing.PlanarYUVLuminanceSource
import com.google.zxing.qrcode.QRCodeReader
import com.google.zxing.common.HybridBinarizer

@Suppress("DEPRECATION")
internal class QrCamera(private val activity: Activity, private val result: (String) -> Unit, private val failed: () -> Unit) : SurfaceHolder.Callback {
    val view = SurfaceView(activity).apply { holder.addCallback(this@QrCamera) }
    private var camera: Camera? = null
    private var lastFrame = 0L

    override fun surfaceCreated(holder: SurfaceHolder) {
        try {
            val source = Camera.open()
            camera = source
            source.setDisplayOrientation(90)
            source.setPreviewDisplay(holder)
            val size = source.parameters.previewSize
            source.setPreviewCallback { data, _ ->
                val now = android.os.SystemClock.elapsedRealtime()
                if (now - lastFrame < 180) return@setPreviewCallback
                lastFrame = now
                val value = runCatching {
                    val image = PlanarYUVLuminanceSource(data, size.width, size.height, 0, 0, size.width, size.height, false)
                    QRCodeReader().decode(BinaryBitmap(HybridBinarizer(image))).text
                }.getOrNull()
                if (value != null) {
                    stop()
                    result(value)
                }
            }
            source.startPreview()
        } catch (_: Exception) {
            stop()
            failed()
        }
    }

    override fun surfaceChanged(holder: SurfaceHolder, format: Int, width: Int, height: Int) = Unit
    override fun surfaceDestroyed(holder: SurfaceHolder) = stop()
    fun stop() {
        camera?.apply {
            setPreviewCallback(null)
            stopPreview()
            release()
        }
        camera = null
    }
}
