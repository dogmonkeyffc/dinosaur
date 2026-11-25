package com.monkeydog.dinosaur.tv

import android.content.Context
import android.net.nsd.NsdManager
import android.net.nsd.NsdServiceInfo
import android.net.wifi.WifiManager

@Suppress("DEPRECATION")
internal class TvDiscovery(
    activity: TvActivity,
    private val found: (List<String>, String?) -> Unit,
    private val finished: () -> Unit,
) {
    private val main = activity.main
    private val manager = activity.getSystemService(Context.NSD_SERVICE) as NsdManager
    private val lock = activity.applicationContext.getSystemService(WifiManager::class.java)
        ?.createMulticastLock("dinosaur-tv")
    private val pending = ArrayDeque<NsdServiceInfo>()
    private val finish = Runnable { stop(); finished() }
    private var listener: NsdManager.DiscoveryListener? = null
    private var resolving = false

    fun start() {
        stop()
        val scan = object : NsdManager.DiscoveryListener {
            override fun onDiscoveryStarted(type: String) = Unit
            override fun onDiscoveryStopped(type: String) = Unit
            override fun onStopDiscoveryFailed(type: String, error: Int) = Unit
            override fun onStartDiscoveryFailed(type: String, error: Int) {
                main.post { if (listener === this) finish.run() }
            }
            override fun onServiceLost(info: NsdServiceInfo) = Unit
            override fun onServiceFound(info: NsdServiceInfo) {
                main.post {
                    if (listener === this) {
                        pending.addLast(info)
                        advance()
                    }
                }
            }
        }
        listener = scan
        runCatching {
            lock?.acquire()
            manager.discoverServices("_dinosaur._tcp.", NsdManager.PROTOCOL_DNS_SD, scan)
        }.onFailure { finish.run() }
        if (listener === scan) main.postDelayed(finish, 15000)
    }

    private fun advance() {
        val scan = listener ?: return
        if (resolving) return
        val service = pending.removeFirstOrNull() ?: return
        val resolver = object : NsdManager.ResolveListener {
            override fun onResolveFailed(info: NsdServiceInfo, error: Int) {
                main.post { resolving = false; advance() }
            }
            override fun onServiceResolved(info: NsdServiceInfo) {
                main.post {
                    resolving = false
                    if (listener === scan) {
                        val publicUrl = info.attributes["url"]?.decodeToString()
                        val advertisedId = info.attributes["id"]?.decodeToString()
                        val address = info.host?.hostAddress?.let { host -> "http://$host:${info.port}" }
                        found(listOfNotNull(publicUrl, address).distinct(), advertisedId)
                    }
                    advance()
                }
            }
        }
        resolving = true
        runCatching { manager.resolveService(service, resolver) }.onFailure {
            resolving = false
            main.post { advance() }
        }
    }

    fun stop() {
        main.removeCallbacks(finish)
        listener?.let { runCatching { manager.stopServiceDiscovery(it) } }
        listener = null
        pending.clear()
        if (lock?.isHeld == true) lock.release()
    }
}
