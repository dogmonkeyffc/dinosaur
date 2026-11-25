package com.monkeydog.dinosaur.mobile

import android.graphics.Canvas
import android.graphics.Paint
import android.graphics.Path
import android.graphics.RectF
import android.os.Build
import android.text.InputType
import android.view.Gravity
import android.view.View
import android.view.WindowInsets
import android.view.inputmethod.EditorInfo
import android.widget.EditText
import android.widget.FrameLayout
import android.widget.HorizontalScrollView
import android.widget.LinearLayout
import org.json.JSONObject

internal class CatalogChrome(private val activity: MainActivity, tab: Int, navigate: (Int) -> Unit) {
    val header = activity.column()
    val body = activity.column()
    var topInset = 0
        private set
    val scroller = RefreshScroll(activity).apply {
        isFillViewport = true
        isVerticalScrollBarEnabled = false
        addView(body)
    }
    val stage = FrameLayout(activity).apply { addView(scroller, FrameLayout.LayoutParams(-1, -1)) }

    init {
        with(activity) {
            generation++
            back = if (tab == 0) null else ({ navigate(0) })
            val bar = navigation(tab, navigate)
            val root = column().apply {
                setBackgroundColor(PHONE_BG)
                fitsSystemWindows = Build.VERSION.SDK_INT < 30
                isFocusableInTouchMode = true
                addView(header)
                addView(stage, LinearLayout.LayoutParams(-1, 0, 1f))
                addView(bar)
            }
            if (Build.VERSION.SDK_INT >= 30) root.setOnApplyWindowInsetsListener { view, insets ->
                val bars = insets.getInsets(WindowInsets.Type.systemBars() or WindowInsets.Type.displayCutout())
                val keyboard = insets.getInsets(WindowInsets.Type.ime())
                topInset = if (tab == 1) bars.top else 0
                view.setPadding(bars.left, if (tab == 1) 0 else bars.top, bars.right, keyboard.bottom)
                bar.setPadding(0, ui(10), 0, if (keyboard.bottom > 0) ui(10) else maxOf(ui(28), bars.bottom))
                WindowInsets.CONSUMED
            }
            setContentView(root)
            root.requestFocus()
        }
    }
}

internal fun MainActivity.searchHeading(query: String, settings: () -> Unit, change: (String) -> Unit): Pair<View, EditText> {
    val input = EditText(this).apply {
        setSingleLine()
        inputType = InputType.TYPE_CLASS_TEXT or InputType.TYPE_TEXT_FLAG_NO_SUGGESTIONS
        imeOptions = EditorInfo.IME_ACTION_SEARCH
        textSize = 17 * .85f
        setTextColor(PHONE_TEXT)
        setHintTextColor(PHONE_MUTED)
        hint = "搜索媒体"
        background = null
        setPadding(ui(14), 0, 0, 0)
        setText(query)
    }
    val clear = phoneText("×", 26, PHONE_MUTED).apply {
        contentDescription = "清除搜索"
        gravity = Gravity.CENTER
        visibility = if (query.isEmpty()) View.GONE else View.VISIBLE
        onTap { input.text.clear() }
    }
    input.addTextChangedListener(object : android.text.TextWatcher {
        override fun beforeTextChanged(value: CharSequence?, start: Int, count: Int, after: Int) = Unit
        override fun onTextChanged(value: CharSequence?, start: Int, before: Int, count: Int) {
            clear.visibility = if (value.isNullOrEmpty()) View.GONE else View.VISIBLE
            change(value.toString())
        }
        override fun afterTextChanged(value: android.text.Editable?) = Unit
    })
    val row = phoneRow().apply {
        setPadding(ui(24), ui(12), ui(24), ui(10))
        addView(phoneRow().apply {
            background = panel(PHONE_SURFACE, 30)
            setPadding(ui(18), 0, ui(6), 0)
            addView(CatalogIcon(this@searchHeading, "search", PHONE_MUTED), LinearLayout.LayoutParams(ui(24), ui(24)))
            addView(input, LinearLayout.LayoutParams(0, -1, 1f))
            addView(clear, LinearLayout.LayoutParams(ui(36), -1))
        }, LinearLayout.LayoutParams(0, ui(56), 1f))
        addView(avatar(settings), LinearLayout.LayoutParams(ui(44), ui(44)).apply { leftMargin = ui(14) })
    }
    return row to input
}

internal fun MainActivity.libraryTabs(
    libraries: List<JSONObject>, selected: Long, select: (Long) -> Unit,
) = column().apply {
    val tray = phoneRow().apply { setPadding(ui(24), 0, ui(24), 0) }
    var active: View? = null
    libraries.forEach { item ->
        val id = item.getLong("id")
        val picked = id == selected
        val tab = FrameLayout(this@libraryTabs).apply {
            isSelected = picked
            addView(phoneText(item.getString("name"), 15, if (picked) PHONE_ACCENT else PHONE_MUTED, true),
                FrameLayout.LayoutParams(-2, -2, Gravity.CENTER))
            if (picked) addView(View(this@libraryTabs).apply { background = panel(PHONE_ACCENT, 3) },
                FrameLayout.LayoutParams(-1, ui(3), Gravity.BOTTOM))
            onTap { select(id) }
        }
        tray.addView(tab, LinearLayout.LayoutParams(-2, ui(52)).apply { rightMargin = ui(30) })
        if (picked) active = tab
    }
    addView(HorizontalScrollView(this@libraryTabs).apply {
        isHorizontalScrollBarEnabled = false
        addView(tray)
        post { active?.let { scrollTo((it.right + ui(24) - width).coerceAtLeast(0), 0) } }
    })
    addView(View(this@libraryTabs).apply { setBackgroundColor(PHONE_BORDER) }, LinearLayout.LayoutParams(-1, ui(1)))
}

private fun MainActivity.navigation(tab: Int, navigate: (Int) -> Unit) = phoneRow().apply {
    setBackgroundColor(0xff1d1e21.toInt())
    setPadding(0, ui(10), 0, ui(28))
    listOf("首页" to "home", "片库" to "library", "记录" to "history").forEachIndexed { index, (label, icon) ->
        val selected = index == tab
        val color = if (selected) PHONE_ACCENT else PHONE_MUTED
        addView(column().apply {
            gravity = Gravity.CENTER_HORIZONTAL
            isSelected = selected
            addView(FrameLayout(this@navigation).apply {
                if (selected) background = panel(PHONE_SELECTED, 18)
                addView(CatalogIcon(this@navigation, icon, color), FrameLayout.LayoutParams(ui(24), ui(24), Gravity.CENTER))
            }, LinearLayout.LayoutParams(ui(64), ui(32)))
            addWithGap(phoneText(label, 12, color, true), ui(5), -2)
            onTap { navigate(index) }
        }, LinearLayout.LayoutParams(0, -2, 1f))
    }
}

internal class CatalogIcon(activity: MainActivity, private val name: String, color: Int) : View(activity) {
    private val pen = Paint(Paint.ANTI_ALIAS_FLAG).apply {
        this.color = color
        style = Paint.Style.STROKE
        strokeWidth = 1.8f
        strokeJoin = Paint.Join.ROUND
        strokeCap = Paint.Cap.ROUND
    }

    override fun onDraw(canvas: Canvas) {
        canvas.save()
        canvas.scale(width / 24f, height / 24f)
        when (name) {
            "search" -> {
                canvas.drawCircle(10f, 10f, 7f, pen)
                canvas.drawLine(15f, 15f, 22f, 22f, pen)
            }
            "history" -> {
                canvas.drawCircle(12f, 12f, 10f, pen)
                canvas.drawLine(12f, 6f, 12f, 12f, pen)
                canvas.drawLine(12f, 12f, 17f, 12f, pen)
            }
            "library" -> listOf(2f, 14f).forEach { x ->
                listOf(2f, 14f).forEach { y -> canvas.drawRoundRect(RectF(x, y, x + 8, y + 8), 1f, 1f, pen) }
            }
            else -> canvas.drawPath(Path().apply {
                moveTo(2f, 11f); lineTo(12f, 2f); lineTo(22f, 11f)
                moveTo(4f, 9f); lineTo(4f, 22f); lineTo(20f, 22f); lineTo(20f, 9f)
            }, pen)
        }
        canvas.restore()
    }
}
