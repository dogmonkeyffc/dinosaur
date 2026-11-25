package com.monkeydog.dinosaur.mobile

import com.monkeydog.dinosaur.android.serverOrigin
import android.content.Context
import android.net.nsd.NsdManager
import android.net.nsd.NsdServiceInfo
import android.os.Handler
import android.os.Looper

@Suppress("DEPRECATION")
internal class ServerPicker(private val activity: MainActivity) {
    private val manager = activity.getSystemService(Context.NSD_SERVICE) as NsdManager
    private val handler = Handler(Looper.getMainLooper())

    fun show(close: () -> Unit, select: (String) -> Unit) = with(activity) {
        val found = mutableSetOf<String>()
        val queue = ArrayDeque<NsdServiceInfo>()
        var resolving = false
        lateinit var listener: NsdManager.DiscoveryListener
        var running = true
        fun stop() {
            if (running) {
                running = false
                runCatching { manager.stopServiceDiscovery(listener) }
            }
        }
        page("选择服务器", { stop(); close() })
        val screen = generation
        button("手动输入地址") {
            form("服务器地址", listOf("HTTP(S) 地址")) { values ->
                runCatching { serverOrigin(values.first()) }.onSuccess { stop(); select(it) }
                    .onFailure { message(it.message ?: "地址无效") }
            }
        }
        fun advance() {
            if (resolving || queue.isEmpty() || !running || generation != screen) return
            resolving = true
            val entry = queue.removeFirst()
            val resolve = object : NsdManager.ResolveListener {
                override fun onResolveFailed(serviceInfo: NsdServiceInfo, errorCode: Int) {
                    handler.post { resolving = false; advance() }
                }
                override fun onServiceResolved(serviceInfo: NsdServiceInfo) {
                    handler.post {
                        resolving = false
                        val host = serviceInfo.host?.hostAddress
                        if (running && generation == screen && host != null) {
                            val address = runCatching { serverOrigin("http://$host:${serviceInfo.port}") }.getOrNull()
                            if (address != null && found.add(address)) {
                                button(serviceInfo.serviceName) { stop(); select(address) }
                            }
                        }
                        advance()
                    }
                }
            }
            runCatching { manager.resolveService(entry, resolve) }.onFailure { resolving = false }
        }
        listener = object : NsdManager.DiscoveryListener {
            override fun onDiscoveryStarted(type: String) = Unit
            override fun onDiscoveryStopped(type: String) = Unit
            override fun onStartDiscoveryFailed(type: String, errorCode: Int) { handler.post { stop() } }
            override fun onStopDiscoveryFailed(type: String, errorCode: Int) = Unit
            override fun onServiceLost(service: NsdServiceInfo) = Unit
            override fun onServiceFound(service: NsdServiceInfo) {
                handler.post { queue.add(service); advance() }
            }
        }
        runCatching { manager.discoverServices("_dinosaur._tcp.", NsdManager.PROTOCOL_DNS_SD, listener) }
        handler.postDelayed({ stop() }, 12000)
    }
}
