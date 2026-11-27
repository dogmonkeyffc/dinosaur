package com.monkeydog.dinosaur.mobile

import android.text.InputType
import android.view.Gravity
import android.view.View
import android.widget.EditText
import android.widget.LinearLayout
import android.widget.PopupMenu
import android.widget.Switch
import com.monkeydog.dinosaur.android.json
import com.monkeydog.dinosaur.android.rows
import org.json.JSONArray
import org.json.JSONObject

private const val USERS_PATH = "/api/admin/users"

internal class AdminUsers(private val activity: MainActivity) {
    fun show(): Unit = with(activity) {
        page("用户管理", { settings.show() })
        val client = api
        button("添加账号") {
            form("添加账号", listOf("用户名", "密码")) { input ->
                val body = json("username" to input[0], "password" to input[1], "role" to "user")
                task({ client.request(USERS_PATH, "POST", body) }) { show() }
            }
        }
        task({ client.request(USERS_PATH).rows("users") }) { users ->
            users.forEach { content.addWithGap(userCard(it), ui(12)) }
        }
    }

    private fun userCard(user: JSONObject) = with(activity) {
        val name = user.getString("username")
        val role = if (user.getString("role") == "admin") "管理员" else "普通账号"
        val state = if (user.getBoolean("disabled")) "已停用" else "已启用"
        phoneCard().apply {
            addView(phoneRow().apply {
                addView(column().apply {
                    addView(phoneText(name, 21, bold = true))
                    addWithGap(phoneText("$role · $state", 14, PHONE_MUTED), ui(8))
                }, LinearLayout.LayoutParams(0, -2, 1f))
                if (user.getLong("id") != account?.user?.optLong("id")) {
                    addView(phoneText("⋯", 24, PHONE_ACCENT).apply {
                        gravity = Gravity.CENTER
                        contentDescription = "管理 $name"
                        onTap { menu(this, user) }
                    }, LinearLayout.LayoutParams(ui(48), ui(48)))
                }
            })
        }
    }

    private fun menu(anchor: View, user: JSONObject) = with(activity) {
        val id = user.getLong("id")
        val name = user.getString("username")
        val disabled = user.getBoolean("disabled")
        val popup = PopupMenu(this, anchor, Gravity.END)
        fun option(title: String, action: () -> Unit) {
            popup.menu.add(title).setOnMenuItemClickListener {
                action()
                true
            }
        }
        option("重置密码") { password(id, name) }
        if (user.getString("role") == "user") option("媒体库权限") { libraries(id, name) }
        val state = if (disabled) "启用账号" else "停用账号"
        option(state) {
            confirm("${if (disabled) "启用" else "停用"} $name？") {
                val client = api
                task({ client.request("$USERS_PATH/$id", "PATCH", json("disabled" to !disabled)) }) { show() }
            }
        }
        popup.show()
    }

    private fun password(id: Long, name: String): Unit = with(activity) {
        var saving = false
        page("重置密码", { if (!saving) show() })
        label(name)
        val first = passwordField("新密码")
        val second = passwordField("再次输入新密码")
        val status = phoneText("", 15, PHONE_MUTED)
        val save = phoneAction("重置密码", {}, color = PHONE_SELECTED, fill = PHONE_ACCENT)
        content.addWithGap(save, ui(16))
        content.addWithGap(status, ui(8))
        save.setOnClickListener {
            val value = first.text.toString()
            if (value.toByteArray(Charsets.UTF_8).size !in 12..1024) {
                status.text = "密码需为 12 至 1024 字节"
                return@setOnClickListener
            }
            if (value != second.text.toString()) {
                status.text = "两次输入的密码不一致"
                return@setOnClickListener
            }
            confirm("重置 $name 的密码并退出其已登录设备？") {
                val client = api
                saving = true
                listOf(first, second, save).forEach { it.isEnabled = false }
                status.text = "正在保存"
                task({ client.request("$USERS_PATH/$id/password", "PUT", json("new_password" to value)) }, {
                    saving = false
                    listOf(first, second, save).forEach { it.isEnabled = true }
                    status.text = it.message ?: "重置密码失败"
                }) {
                    message("密码已重置")
                    show()
                }
            }
        }
    }

    private fun MainActivity.passwordField(title: String) = EditText(this).apply {
        setSingleLine()
        inputType = InputType.TYPE_CLASS_TEXT or InputType.TYPE_TEXT_VARIATION_PASSWORD
        if (android.os.Build.VERSION.SDK_INT >= 26) importantForAutofill = View.IMPORTANT_FOR_AUTOFILL_NO
        hint = title
        contentDescription = title
        setPadding(ui(16), ui(12), ui(16), ui(12))
        background = panel(PHONE_SURFACE, 14)
        content.addWithGap(this, ui(12))
    }

    private fun libraries(id: Long, name: String): Unit = with(activity) {
        var saving = false
        page("媒体库权限", { if (!saving) show() })
        label(name)
        val client = api
        val path = "$USERS_PATH/$id/libraries"
        task({ client.request(path).rows("libraries") }) { rows ->
            val changes = linkedMapOf<Long, Boolean>()
            val controls = mutableListOf<Switch>()
            val save = phoneAction("保存", {}, color = PHONE_SELECTED, fill = PHONE_ACCENT)
            fun updateSave() {
                save.isEnabled = changes.isNotEmpty() && !saving
                save.alpha = if (save.isEnabled) 1f else .5f
            }
            if (rows.isEmpty()) label("暂无媒体库")
            rows.forEach { item ->
                val title = item.getString("name")
                if (item.getBoolean("restricted")) {
                    val library = item.getLong("id")
                    val granted = item.getBoolean("granted")
                    val switch = switchControl(title, granted) { value, _ ->
                        if (value == granted) changes.remove(library) else changes[library] = value
                        updateSave()
                    }
                    controls.add(switch)
                    content.addWithGap(switch, ui(8))
                } else {
                    label("$title · 公开")
                }
            }
            if (controls.isNotEmpty()) {
                content.addWithGap(save, ui(16))
                updateSave()
                save.setOnClickListener {
                    val updates = changes.map { (library, granted) -> json("library_id" to library, "granted" to granted) }
                    val body = json("changes" to JSONArray(updates))
                    saving = true
                    controls.forEach { it.isEnabled = false }
                    updateSave()
                    task({ client.request(path, "PATCH", body) }, {
                        saving = false
                        controls.forEach { it.isEnabled = true }
                        updateSave()
                        message(it.message ?: "保存权限失败")
                    }) {
                        message("媒体库权限已保存")
                        show()
                    }
                }
            }
        }
    }
}
