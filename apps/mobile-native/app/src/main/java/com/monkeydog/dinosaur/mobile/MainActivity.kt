package com.monkeydog.dinosaur.mobile

import com.monkeydog.dinosaur.android.Account
import com.monkeydog.dinosaur.android.Accounts
import com.monkeydog.dinosaur.android.Api
import com.monkeydog.dinosaur.android.json
import com.monkeydog.dinosaur.android.serverOrigin
import android.app.Activity
import android.content.Intent
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
    private var accountReturn: (() -> Unit)? = null
    private var signingOut = false
    internal val catalog by lazy { Catalog(this) }
    internal val settings by lazy { Settings(this) }
    internal val admin by lazy { Admin(this) }

    override fun onCreate(state: Bundle?) {
        super.onCreate(state)
        window.statusBarColor = PHONE_BG
        window.navigationBarColor = PHONE_BG
        if (android.os.Build.VERSION.SDK_INT >= 29) window.isNavigationBarContrastEnforced = false
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

    internal fun showAccounts(returnTo: (() -> Unit)? = accountReturn) {
        accountReturn = returnTo
        accountPage(if (accounts.isEmpty()) "登录 Dinosaur" else "切换账号", account?.let { { leaveAccounts() } })
        accounts.forEach { item ->
            content.addWithGap(phoneCard().apply {
                addView(accountRow(item, selected = item.key == account?.key).apply {
                    onTap {
                        if (!signingOut) {
                            account = item
                            vault.active = item.key
                            leaveAccounts()
                        }
                    }
                })
                addWithGap(phoneAction("退出账号", { signOut(item) }, color = PHONE_TEXT), ui(20), -2)
            }, ui(22))
        }
        content.addWithGap(phoneCard().apply {
            addView(phoneAction("配对码登录", {
                if (!signingOut) ServerPicker(this@MainActivity).show({ showAccounts() }) { origin ->
                    form("配对码登录", listOf("八位配对码")) { fields -> signIn(origin, fields[0], false) }
                }
            }, color = 0xff14233b.toInt(), fill = PHONE_ACCENT))
            addWithGap(phoneAction("扫描 Web 登录二维码", {
                if (!signingOut) {
                    if (checkSelfPermission(android.Manifest.permission.CAMERA) == android.content.pm.PackageManager.PERMISSION_GRANTED) scanQr()
                    else requestPermissions(arrayOf(android.Manifest.permission.CAMERA), 41)
                }
            }, color = PHONE_TEXT), ui(20))
            addWithGap(phoneAction("二维码登录链接", {
                if (!signingOut) form("二维码登录链接", listOf("登录链接")) { fields -> handleLink(Uri.parse(fields[0])) }
            }, color = PHONE_TEXT), ui(20))
        }, ui(22))
    }

    private fun leaveAccounts() {
        if (signingOut) return
        val destination = accountReturn
        accountReturn = null
        if (destination != null) destination() else catalog.show()
    }

    private fun signOut(item: Account) {
        if (signingOut) return
        confirm("退出 ${item.user.getString("username")}？") {
            val remaining = accounts.filter { it.key != item.key }
            val active = account?.takeUnless { it.key == item.key } ?: remaining.firstOrNull()
            signingOut = true
            task({
                vault.save(remaining)
                runCatching { item.api.request("/api/auth/logout", "POST") }
            }, {
                signingOut = false
                message(it.message ?: "退出失败")
            }) {
                signingOut = false
                accounts = remaining
                account = active
                vault.active = active?.key
                if (active == null) accountReturn = null
                showAccounts()
            }
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
        if (link == null || signingOut) return
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
            accountReturn = null
            catalog.show()
        }
    }

    @Deprecated("Android back navigation")
    override fun onBackPressed() { back?.invoke() ?: finish() }
    override fun onStop() { playback?.pause(); super.onStop() }
    override fun onDestroy() { playback?.release(); catalog.dispose(); worker.shutdown(); super.onDestroy() }
}
