package com.monkeydog.dinosaur.mobile

import android.text.InputFilter
import android.text.InputType
import android.view.Gravity
import android.view.View
import android.view.inputmethod.InputMethodManager
import android.widget.EditText
import android.widget.LinearLayout
import android.widget.TextView
import com.monkeydog.dinosaur.android.json
import org.json.JSONObject

internal class PairTv(private val activity: MainActivity, private val close: () -> Unit) {
    private var code = ""
    private var pairing: JSONObject? = null
    private var approved = false
    private var accountKey: String? = null

    fun show(): Unit = with(activity) {
        val account = requireNotNull(account)
        if (account.key != accountKey) {
            accountKey = account.key
            code = ""
            pairing = null
            approved = false
        }
        accountPage("登录电视", close)
        content.addView(phoneCard().apply {
            addView(accountRow(account))
            addWithGap(phoneAction("切换账号", { showAccounts { this@PairTv.show() } }, color = PHONE_TEXT), ui(20))
        })
        val error = phoneText("", 14, 0xffffb4ab.toInt()).apply { visibility = View.GONE }
        content.addWithGap(phoneCard().apply {
            addView(phoneText(if (approved) "✓" else "▣", 54, PHONE_ACCENT).apply {
                gravity = Gravity.CENTER
                setPadding(0, ui(18), 0, ui(18))
                importantForAccessibility = View.IMPORTANT_FOR_ACCESSIBILITY_NO
            })
            val device = pairing
            if (device == null) codeForm(this, error) else deviceActions(this, error, device)
            addWithGap(error, ui(20))
        }, ui(22))
    }

    private fun codeForm(card: LinearLayout, error: TextView) = with(activity) {
        val input = EditText(this).apply {
            minHeight = ui(74)
            background = panel(0xff111215.toInt(), 20)
            setTextColor(PHONE_TEXT)
            setHintTextColor(PHONE_MUTED)
            inputType = InputType.TYPE_CLASS_NUMBER
            gravity = Gravity.CENTER
            textSize = 32 * .85f
            letterSpacing = .27f
            filters = arrayOf(InputFilter.LengthFilter(6))
            contentDescription = "电视六位配对码"
            hint = "000000"
            setText(code)
        }
        val client = api
        lateinit var lookup: TextView
        lookup = primary("查找电视") {
            code = input.text.toString()
            getSystemService(InputMethodManager::class.java).hideSoftInputFromWindow(input.windowToken, 0)
            lookup.isEnabled = false
            input.isEnabled = false
            task({ client.request("/api/devices/pairings/lookup", "POST", json("code" to code)) }, {
                lookup.isEnabled = true
                input.isEnabled = true
                showError(error, it)
            }) { pairing = it; show() }
        }
        fun enabled(value: Boolean) {
            lookup.isEnabled = value
            lookup.alpha = if (value) 1f else .5f
        }
        enabled(code.length == 6)
        input.addTextChangedListener(object : android.text.TextWatcher {
            override fun beforeTextChanged(s: CharSequence?, start: Int, count: Int, after: Int) = Unit
            override fun afterTextChanged(s: android.text.Editable?) = Unit
            override fun onTextChanged(s: CharSequence?, start: Int, before: Int, count: Int) {
                enabled(s?.length == 6)
                error.visibility = View.GONE
            }
        })
        card.addWithGap(input, ui(20))
        card.addWithGap(lookup, ui(20))
    }

    private fun deviceActions(card: LinearLayout, error: TextView, device: JSONObject) = with(activity) {
        val name = device.getString("name")
        card.addWithGap(phoneText(if (approved) "已授权 $name" else name, 20).apply { gravity = Gravity.CENTER }, ui(20))
        if (approved) {
            card.addWithGap(primary("完成", close), ui(20))
            return@with
        }
        val client = api
        lateinit var confirm: TextView
        val retry = phoneAction("重新输入", { pairing = null; show() }, color = PHONE_TEXT)
        confirm = primary("确认授权") {
            confirm.isEnabled = false
            retry.isEnabled = false
            task({ client.request("/api/devices/pairings/approve", "POST",
                json("pairing_id" to device.getString("pairing_id"))) }, {
                confirm.isEnabled = true
                retry.isEnabled = true
                showError(error, it)
            }) { approved = true; show() }
        }
        card.addWithGap(confirm, ui(20))
        card.addWithGap(retry, ui(20))
    }

    private fun showError(error: TextView, failure: Throwable) {
        error.text = failure.message ?: "操作失败"
        error.visibility = View.VISIBLE
    }

    private fun MainActivity.primary(title: String, run: () -> Unit) =
        phoneAction(title, run, color = 0xff14233b.toInt(), fill = PHONE_ACCENT)
}
