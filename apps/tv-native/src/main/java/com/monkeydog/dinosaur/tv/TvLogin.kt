package com.monkeydog.dinosaur.tv

import android.widget.TextView
import com.monkeydog.dinosaur.android.Account
import com.monkeydog.dinosaur.android.Api
import com.monkeydog.dinosaur.android.json
import com.monkeydog.dinosaur.android.serverOrigin
import org.json.JSONObject

internal class TvLogin(private val activity: TvActivity) {
    fun servers() = with(activity) {
        val (page, content) = panelPage("选择服务器")
        val status = label("正在查找局域网服务器", color = muted)
        val results = column()
        val enter = control("手动输入地址") {
            input("服务器地址", "", address = true) { address ->
                task(page, { probeTvServer(address) }, { status.text = it.message ?: "无法连接服务器" }) {
                    pair(page, it)
                }
            }
        }
        content.addControl(enter)
        val discovered = mutableSetOf<String>()
        val discovery = TvDiscovery(this, { candidates, advertisedId ->
            task(page, {
                candidates.firstNotNullOfOrNull { address ->
                    runCatching { probeTvServer(address, advertisedId) }.getOrNull()
                }
            }, {}) { server ->
                if (server != null && discovered.add(server.getString("id"))) {
                    status.text = ""
                    results.addControl(control(server.getString("name")) { pair(page, server) })
                }
            }
        }, { if (discovered.isEmpty()) status.text = "未发现服务器" })
        content.addControl(control("重新扫描") {
            discovered.clear()
            results.removeAllViews()
            status.text = "正在查找局域网服务器"
            discovery.start()
        })
        content.addView(status)
        content.addView(results)
        page.focus = enter
        page.leave = discovery::stop
        page.resume = { if (discovered.isEmpty()) discovery.start() }
        show(page, reset = true)
        discovery.start()
    }

    private fun pair(previous: TvPage, server: JSONObject) = with(activity) {
        val (page, content) = panelPage("登录电视")
        val code = label("", 64f, accent, bold = true).apply { letterSpacing = .15f }
        val status = label("正在创建配对码", color = muted)
        val pairing = TvPairing(this, page, server, code, status)
        content.addView(label("在手机「登录电视」或网页「账户安全」中输入配对码"))
        content.addView(code)
        content.addView(status)
        val controls = row()
        val refresh = control("刷新配对码", pairing::start)
        controls.addControl(refresh, horizontal = true)
        controls.addControl(control("返回服务器列表", ::back), horizontal = true)
        content.addView(controls)
        page.focus = refresh
        page.leave = pairing::stop
        if (current === previous) {
            show(page)
            pairing.start()
        }
    }
}

internal fun probeTvServer(address: String, expectedId: String? = null): JSONObject {
    val origin = serverOrigin(address)
    val info = Api(origin).request("/api/server")
    val id = info.getString("id")
    val capabilities = info.getJSONArray("capabilities")
    require(id.matches(Regex("[a-f0-9]{32}")) && (expectedId == null || id == expectedId) &&
        info.optInt("api_version") == 1 && info.optString("name").isNotBlank() &&
        (0 until capabilities.length()).any { capabilities.optString(it) == "tv_pairing" }) {
        "服务器不支持电视配对"
    }
    return info.put("url", origin)
}

private class TvPairing(
    private val activity: TvActivity,
    private val page: TvPage,
    private val server: JSONObject,
    private val code: TextView,
    private val status: TextView,
) {
    private val api = Api(server.getString("url"))
    private var revision = 0
    private var secret: String? = null
    private var expires = 0L
    private val poll = Runnable { poll(revision) }
    private val clock = object : Runnable {
        override fun run() {
            val remaining = expires - System.currentTimeMillis() / 1000
            if (remaining <= 0) {
                stop()
                code.text = ""
                status.text = "配对码已过期"
            } else {
                status.text = "${remaining} 秒后过期"
                activity.main.postDelayed(this, 1000)
            }
        }
    }

    fun start() {
        stop()
        val request = revision
        code.text = ""
        status.text = "正在创建配对码"
        activity.worker.execute {
            val result = runCatching { api.request("/api/devices/pairings", "POST", json("name" to "Dinosaur TV")) }
            activity.main.post {
                result.onSuccess { pairing ->
                    val issued = pairing.getString("device_code")
                    if (request != revision) cancel(issued) else {
                        secret = issued
                        code.text = pairing.getString("code")
                        expires = pairing.getLong("expires_at")
                        clock.run()
                        schedulePoll(pairing)
                    }
                }.onFailure { if (request == revision) status.text = it.message ?: "无法创建配对码" }
            }
        }
    }

    private fun poll(request: Int) {
        val issued = secret ?: return
        activity.worker.execute {
            val result = runCatching { api.request("/api/devices/pairings/poll", "POST", json("device_code" to issued)) }
            activity.main.post {
                result.onSuccess { reply ->
                    if (reply.has("token")) {
                        val account = runCatching {
                            val key = server.getString("id") + "_" + reply.getJSONObject("user").getLong("id")
                            Account(reply.put("key", key).put("server", server))
                        }
                        account.onSuccess {
                            if (request == revision) {
                                secret = null
                                stop()
                                code.text = ""
                                status.text = "正在登录"
                                activity.openSession(page, it)
                            } else if (!activity.worker.isShutdown) {
                                activity.worker.execute { runCatching { it.api.request("/api/auth/logout", "POST") } }
                            }
                        }.onFailure { if (request == revision) failed("服务器返回了无效的账号") }
                    } else if (request == revision) schedulePoll(reply)
                }.onFailure { if (request == revision) failed(it.message ?: "配对请求失败") }
            }
        }
    }

    private fun schedulePoll(reply: JSONObject) {
        if (secret != null) activity.main.postDelayed(poll, reply.optLong("interval", 3).coerceIn(3, 15) * 1000)
    }

    private fun failed(message: String) {
        stop()
        code.text = ""
        status.text = message
    }

    private fun cancel(issued: String) {
        if (!activity.worker.isShutdown) activity.worker.execute {
            runCatching { api.request("/api/devices/pairings/cancel", "POST", json("device_code" to issued)) }
        }
    }

    fun stop() {
        revision++
        activity.main.removeCallbacks(poll)
        activity.main.removeCallbacks(clock)
        secret?.let(::cancel)
        secret = null
    }
}
