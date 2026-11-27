package com.monkeydog.dinosaur.mobile

import android.text.InputType
import android.view.View
import android.widget.EditText
import com.monkeydog.dinosaur.android.json
import org.json.JSONObject

internal class LibraryEditor(
    private val activity: MainActivity,
    item: JSONObject?,
    private val cancel: () -> Unit,
    private val onSaved: (JSONObject) -> Unit,
) {
    private val id = item?.getLong("id")
    private val types = linkedMapOf("movie" to "电影", "tv" to "电视剧", "video" to "普通视频")
    private var type = item?.getString("type")
    private var name = item?.getString("name").orEmpty()
    private var root = item?.getString("root_path").orEmpty()
    private var saving = false

    fun show(): Unit = with(activity) {
        page(if (id == null) "添加媒体库" else "编辑媒体库", { if (!saving) cancel() })
        val nameInput = field("名称", name, InputType.TYPE_CLASS_TEXT)
        fun typeTitle() = type?.let(types::get)?.let { "媒体类型 · $it" } ?: "选择媒体类型"
        val typeInput = phoneAction(typeTitle(), {}).apply {
            setOnClickListener {
                val keys = types.keys.toList()
                choose("媒体类型", types.values.toList(), keys.indexOf(type)) { index ->
                    type = keys[index]
                    text = typeTitle()
                }
            }
        }
        content.addWithGap(typeInput, ui(16))
        val pathInput = field("服务器媒体目录", root, InputType.TYPE_CLASS_TEXT or InputType.TYPE_TEXT_FLAG_NO_SUGGESTIONS)
        val browse = phoneAction("浏览目录", {
            name = nameInput.text.toString()
            root = pathInput.text.toString()
            DirectoryPicker(activity).show(root.trim(), previous = ::show) { selected ->
                root = selected
                show()
            }
        })
        content.addWithGap(browse, ui(8))

        val status = phoneText("", 15, PHONE_MUTED).apply {
            accessibilityLiveRegion = View.ACCESSIBILITY_LIVE_REGION_POLITE
            visibility = View.GONE
        }
        val save = phoneAction("保存", {}, color = PHONE_SELECTED, fill = PHONE_ACCENT)
        content.addWithGap(save, ui(20))
        content.addWithGap(status, ui(10))
        fun feedback(text: String) {
            status.text = text
            status.visibility = View.VISIBLE
        }
        fun setSaving(value: Boolean) {
            saving = value
            listOf(nameInput, typeInput, pathInput, browse, save).forEach { it.isEnabled = !value }
        }
        save.setOnClickListener {
            val selected = type
            if (selected == null) {
                feedback("请选择媒体类型")
                return@setOnClickListener
            }
            name = nameInput.text.toString()
            root = pathInput.text.toString()
            val path = "/api/admin/libraries" + (id?.let { "/$it" } ?: "")
            val body = json("name" to name, "root_path" to root, "type" to selected)
            val client = api
            setSaving(true)
            feedback("正在保存")
            task({ client.request(path, if (id == null) "POST" else "PUT", body).getJSONObject("library") }, {
                setSaving(false)
                feedback(it.message ?: "保存失败")
            }) { saved ->
                setSaving(false)
                onSaved(saved)
            }
        }
    }

    private fun MainActivity.field(title: String, value: String, flags: Int): EditText {
        val input = EditText(this).apply {
            id = View.generateViewId()
            inputType = flags
            setSingleLine()
            setTextColor(PHONE_TEXT)
            setText(value)
            setPadding(ui(16), ui(12), ui(16), ui(12))
            background = panel(PHONE_SURFACE, 14)
        }
        content.addWithGap(phoneText(title, 16, PHONE_MUTED).apply { labelFor = input.id }, ui(16))
        content.addWithGap(input, ui(8))
        return input
    }
}
