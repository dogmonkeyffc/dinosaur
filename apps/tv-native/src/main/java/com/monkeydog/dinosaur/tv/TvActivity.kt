package com.monkeydog.dinosaur.tv

import android.app.Activity
import android.app.AlertDialog
import android.os.Build
import android.os.Bundle
import android.os.Handler
import android.os.Looper
import android.view.KeyEvent
import android.view.View
import android.view.WindowInsets
import android.view.WindowInsetsController
import android.widget.FrameLayout
import android.widget.Toast
import com.monkeydog.dinosaur.android.Account
import com.monkeydog.dinosaur.android.Accounts
import com.monkeydog.dinosaur.android.ApiFailure
import java.util.concurrent.Executors

internal class TvPage(val view: View) {
    var focus: View? = null
    var leave: () -> Unit = {}
    var resume: () -> Unit = {}
}

class TvActivity : Activity() {
    internal val worker = Executors.newSingleThreadExecutor()
    internal val main = Handler(Looper.getMainLooper())
    private val stack = ArrayDeque<TvPage>()
    private lateinit var host: FrameLayout
    private lateinit var vault: Accounts
    private val login by lazy { TvLogin(this) }
    internal val artwork by lazy { TvArtwork(this) }
    internal var player: TvPlayback? = null
    internal var started = false
        private set
    internal val current get() = stack.lastOrNull()

    override fun onCreate(state: Bundle?) {
        super.onCreate(state)
        host = FrameLayout(this).apply { setBackgroundColor(surface) }
        setContentView(host)
        immersive()
        val loading = panelPage("Dinosaur TV").first
        show(loading, reset = true)
        restore(loading)
    }

    private fun restore(page: TvPage) {
        task(page, {
            vault = Accounts(this)
            vault.load().firstOrNull()?.also { account ->
                try {
                    account.api.request("/api/auth/me")
                } catch (failure: ApiFailure) {
                    if (failure.status != 401) throw failure
                    vault.save(emptyList())
                    return@task null
                }
            }
        }, { failure ->
            val retry = panelPage("连接服务器")
            retry.second.addView(label(failure.message ?: "无法恢复登录", color = muted))
            retry.second.addControl(control("重试") { restore(retry.first) })
            retry.second.addControl(control("选择服务器") { login.servers() })
            show(retry.first, reset = true)
        }) { account ->
            if (account == null) login.servers() else TvCatalog(this, account).home()
        }
    }

    internal fun openSession(page: TvPage, account: Account) {
        task(page, {
            if (!::vault.isInitialized) vault = Accounts(this)
            val previous = runCatching { vault.load() }.getOrDefault(emptyList())
            vault.save(listOf(account))
            previous.filter { it.token != account.token }.forEach {
                runCatching { it.api.request("/api/auth/logout", "POST") }
            }
        }) { TvCatalog(this, account).home() }
    }

    internal fun logout(page: TvPage, account: Account) {
        AlertDialog.Builder(this).setTitle("退出登录？").setNegativeButton("取消", null)
            .setPositiveButton("退出") { _, _ ->
                task(page, { vault.save(emptyList()) }) {
                    login.servers()
                    worker.execute { runCatching { account.api.request("/api/auth/logout", "POST") } }
                }
            }.show()
    }

    internal fun panelPage(title: String): Pair<TvPage, android.widget.LinearLayout> {
        val content = column().apply {
            setPadding(dp(52), dp(30), dp(52), dp(32))
            addView(label(title, 32f, bold = true).apply { setPadding(0, 0, 0, dp(18)) })
        }
        return TvPage(scroll(content)) to content
    }

    internal fun show(page: TvPage, reset: Boolean = false) {
        current?.focus = host.findFocus() ?: current?.focus
        if (reset) {
            while (stack.isNotEmpty()) stack.removeLast().leave()
        }
        stack.addLast(page)
        attach(page)
    }

    private fun attach(page: TvPage) {
        host.removeAllViews()
        host.addView(page.view, FrameLayout.LayoutParams(-1, -1))
        page.view.post {
            if (current === page) {
                if (page.focus?.requestFocus() != true) page.view.focusSearch(View.FOCUS_FORWARD)?.requestFocus()
            }
        }
    }

    internal fun back() {
        if (stack.size <= 1) {
            finish()
        } else {
            stack.removeLast().leave()
            val page = stack.last()
            attach(page)
            page.resume()
        }
    }

    @Suppress("DEPRECATION")
    override fun onBackPressed() = back()

    override fun dispatchKeyEvent(event: KeyEvent): Boolean {
        if (player?.key(event) == true) return true
        return super.dispatchKeyEvent(event)
    }

    internal fun <T> task(page: TvPage, job: () -> T, error: (Throwable) -> Unit = { message(it.message ?: "请求失败") },
                          done: (T) -> Unit) {
        worker.execute {
            val result = runCatching(job)
            main.post { if (current === page && !isFinishing) result.fold(done, error) }
        }
    }

    internal fun message(value: String) = Toast.makeText(this, value, Toast.LENGTH_LONG).show()

    @Suppress("DEPRECATION")
    private fun immersive() {
        if (Build.VERSION.SDK_INT >= 30) {
            window.insetsController?.apply {
                systemBarsBehavior = WindowInsetsController.BEHAVIOR_SHOW_TRANSIENT_BARS_BY_SWIPE
                hide(WindowInsets.Type.systemBars())
            }
        } else {
            window.decorView.systemUiVisibility = View.SYSTEM_UI_FLAG_IMMERSIVE_STICKY or
                View.SYSTEM_UI_FLAG_FULLSCREEN or View.SYSTEM_UI_FLAG_HIDE_NAVIGATION
        }
    }

    override fun onWindowFocusChanged(hasFocus: Boolean) {
        super.onWindowFocusChanged(hasFocus)
        if (hasFocus) immersive()
    }

    override fun onStart() {
        super.onStart()
        started = true
    }

    override fun onStop() {
        started = false
        player?.pause()
        super.onStop()
    }

    override fun onDestroy() {
        while (stack.isNotEmpty()) stack.removeLast().leave()
        artwork.close()
        worker.shutdown()
        super.onDestroy()
    }
}
