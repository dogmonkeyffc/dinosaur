package com.monkeydog.dinosaur.mobile

import android.app.Activity
import android.content.Intent
import android.graphics.Color
import android.net.Uri
import android.os.Bundle
import android.widget.LinearLayout
import java.util.concurrent.Executors

class MainActivity : Activity() {
    internal val worker = Executors.newSingleThreadExecutor()
    internal lateinit var vault: Accounts
    internal var accounts = emptyList<Account>()
    internal var account: Account? = null
    internal val api get() = requireNotNull(account).api
    internal lateinit var content: LinearLayout
    internal var back: (() -> Unit)? = null
    internal var generation = 0
    internal var playback: PlaybackScreen? = null
    internal val catalog by lazy { Catalog(this) }
    internal val settings by lazy { Settings(this) }
    internal val admin by lazy { Admin(this) }

    override fun onCreate(state: Bundle?) {
        super.onCreate(state)
        window.statusBarColor = Color.BLACK
        window.navigationBarColor = Color.BLACK
        vault = Accounts(this)
        accounts = runCatching { vault.load() }.getOrElse {
            message("无法读取已保存的账号，请重新登录")
            emptyList()
        }
        account = accounts.firstOrNull { it.key == vault.active } ?: accounts.firstOrNull()
        if (account == null) showAccounts() else catalog.show()
        handleLink(intent.data)
    }

    override fun onNewIntent(intent: Intent) {
        super.onNewIntent(intent)
        handleLink(intent.data)
    }

    internal fun <T> task(work: () -> T, failed: ((Throwable) -> Unit)? = null, done: (T) -> Unit) {
        val screen = generation
        worker.execute {
            val result = runCatching(work)
            runOnUiThread {
                if (screen == generation && !isFinishing && !isDestroyed) {
                    result.fold(done) { error -> failed?.invoke(error) ?: message(error.message ?: "操作失败") }
                }
            }
        }
    }

    internal fun showAccounts() {
        page("切换账号", account?.let { { catalog.show() } })
        accounts.forEach { item ->
            button(item.user.getString("username") + " · " + item.server.optString("name")) {
                account = item
                vault.active = item.key
                catalog.show()
            }
        }
        account?.let { active ->
            button("退出当前账号") {
                confirm("退出当前账号？") {
                    val remaining = accounts.filter { it.key != active.key }
                    task({
                        vault.save(remaining)
                        runCatching { active.api.request("/api/auth/logout", "POST") }
                    }) {
                        accounts = remaining
                        account = remaining.firstOrNull()
                        vault.active = account?.key
                        showAccounts()
                    }
                }
            }
        }
        button("配对码登录") {
            ServerPicker(this).show({ showAccounts() }) { origin ->
                form("配对码登录", listOf("八位配对码")) { fields -> signIn(origin, fields[0], false) }
            }
        }
        button("二维码登录链接") {
            form("二维码登录链接", listOf("登录链接")) { fields -> handleLink(Uri.parse(fields[0])) }
        }
        button("扫描登录二维码") {
            if (checkSelfPermission(android.Manifest.permission.CAMERA) == android.content.pm.PackageManager.PERMISSION_GRANTED) scanQr()
            else requestPermissions(arrayOf(android.Manifest.permission.CAMERA), 41)
        }
    }

    override fun onRequestPermissionsResult(requestCode: Int, permissions: Array<out String>, grantResults: IntArray) {
        super.onRequestPermissionsResult(requestCode, permissions, grantResults)
        if (requestCode == 41 && grantResults.firstOrNull() == android.content.pm.PackageManager.PERMISSION_GRANTED) scanQr()
    }

    private fun scanQr() {
        page("扫描登录二维码", { showAccounts() })
        val camera = QrCamera(this, { code -> showAccounts(); handleLink(Uri.parse(code)) },
            { showAccounts(); message("无法打开摄像头") })
        content.addView(camera.view, LinearLayout.LayoutParams(-1, dp(420)))
    }

    private fun handleLink(link: Uri?) {
        if (link == null) return
        try {
            require(link.scheme == "dinosaur" && link.authority == "login" &&
                link.path.orEmpty() in listOf("", "/") && link.fragment == null) { "登录链接无效" }
            val server = serverOrigin(link.getQueryParameters("server").single())
            val ticket = link.getQueryParameters("ticket").single()
            require(ticket.matches(Regex("[A-Za-z0-9_-]{43}"))) { "登录链接无效" }
            confirm("登录 $server？") { signIn(server, ticket, true) }
        } catch (_: Exception) {
            message("登录链接无效")
        }
    }

    private fun signIn(raw: String, code: String, ticket: Boolean) {
        val previous = accounts
        task({
            val origin = serverOrigin(raw)
            require(code.matches(if (ticket) Regex("[A-Za-z0-9_-]{43}") else Regex("[0-9]{8}"))) { "登录码无效" }
            val client = Api(origin)
            val server = client.request("/api/server").put("url", origin)
            val serverId = server.getString("id")
            require(serverId.matches(Regex("[a-f0-9]{32}"))) { "服务器身份无效" }
            val mode = if (ticket) "qr" else "login-code"
            val field = if (ticket) "ticket" else "code"
            val session = client.request("/api/devices/$mode/exchange", "POST", json(field to code, "name" to "Dinosaur Android"))
            val item = Account(session.put("server", server).put("key", serverId + "_" + session.getJSONObject("user").getLong("id")))
            val old = previous.firstOrNull { it.key == item.key }
            require(old == null || old.origin == item.origin) { "请先退出原服务器上的同一账号" }
            val updated = listOf(item) + previous.filter { it.key != item.key }
            require(updated.size <= 12) { "最多保存 12 个账号" }
            vault.save(updated)
            old?.let { runCatching { it.api.request("/api/auth/logout", "POST") } }
            updated
        }) {
            accounts = it
            account = it.first()
            vault.active = account?.key
            catalog.show()
        }
    }

    @Deprecated("Android back navigation")
    override fun onBackPressed() { back?.invoke() ?: finish() }
    override fun onStop() { playback?.pause(); super.onStop() }
    override fun onDestroy() { playback?.release(); catalog.dispose(); worker.shutdown(); super.onDestroy() }
}
