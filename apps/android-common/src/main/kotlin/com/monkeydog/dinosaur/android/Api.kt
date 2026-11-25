package com.monkeydog.dinosaur.android

import android.net.Uri
import org.json.JSONArray
import org.json.JSONObject
import java.net.URI
import java.util.concurrent.TimeUnit
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.RequestBody.Companion.toRequestBody

internal fun json(vararg pairs: Pair<String, Any?>) = JSONObject().apply {
    pairs.forEach { (name, value) -> put(name, value ?: JSONObject.NULL) }
}

internal fun JSONObject.rows(name: String): List<JSONObject> {
    val array = optJSONArray(name) ?: JSONArray()
    return (0 until array.length()).mapNotNull(array::optJSONObject)
}

internal fun serverOrigin(raw: String): String {
    val url = URI(raw.trim())
    val host = url.host?.lowercase() ?: error("请输入服务器地址")
    val scheme = url.scheme?.lowercase()
    require(scheme in listOf("http", "https") && url.userInfo == null &&
        url.rawQuery == null && url.rawFragment == null && url.path in listOf("", "/") &&
        (url.port == -1 || url.port in 1..65535)) { "服务器地址无效" }
    if (scheme == "http") {
        val parts = host.split('.')
        val ip = parts.mapNotNull(String::toIntOrNull)
        val local = parts.size == 4 && ip.size == 4 && ip.all { it in 0..255 } &&
            (ip[0] in listOf(10, 127) || ip[0] == 192 && ip[1] == 168 || ip[0] == 172 && ip[1] in 16..31)
        require(local || host == "localhost" || host.endsWith(".local")) { "公网服务器请使用 HTTPS" }
    }
    return URI(scheme, null, host, url.port, null, null, null).toString()
}

internal class ApiFailure(val status: Int, message: String) : IllegalStateException(message)

internal class Api(val origin: String, private val token: String = "") {
    companion object {
        val http = OkHttpClient.Builder().followRedirects(false).followSslRedirects(false)
            .connectTimeout(8, TimeUnit.SECONDS).readTimeout(16, TimeUnit.SECONDS).build()
        private val jsonType = "application/json".toMediaType()
    }
    val headers: Map<String, String> get() =
        if (token.isEmpty()) emptyMap() else mapOf("Authorization" to "Bearer $token")

    fun resource(path: String): Uri {
        val base = URI(origin)
        val target = base.resolve(path).normalize()
        require(target.scheme == base.scheme && target.rawAuthority == base.rawAuthority &&
            target.path.startsWith("/api/") && target.rawFragment == null) { "资源地址无效" }
        return Uri.parse(target.toString())
    }

    fun request(path: String, method: String = "GET", body: JSONObject? = null): JSONObject {
        val result = exchange(path, method, body)
        return if (result.isEmpty()) JSONObject() else JSONObject(result)
    }

    fun text(path: String): String = exchange(path, "GET", null)

    fun stream(path: String, offset: Long): Uri {
        val uri = resource(path)
        val target = uri.buildUpon().clearQuery()
        uri.queryParameterNames.filter { it != "start" }.forEach { name ->
            uri.getQueryParameters(name).forEach { target.appendQueryParameter(name, it) }
        }
        return target.appendQueryParameter("start", (offset / 1000.0).toString()).build()
    }

    private fun exchange(path: String, method: String, body: JSONObject?): String {
        val payload = if (body != null || method in listOf("POST", "PUT", "PATCH"))
            (body?.toString() ?: "{}").toRequestBody(jsonType) else null
        val request = Request.Builder().url(resource(path).toString()).method(method, payload)
        headers.forEach { (name, value) -> request.header(name, value) }
        return http.newCall(request.build()).execute().use { response ->
            val status = response.code
            if (status == 401) throw ApiFailure(status, "登录已过期，请重新登录")
            if (status == 403) throw ApiFailure(status, "当前账号没有操作权限")
            if (status !in 200..299) {
                val message = runCatching { JSONObject(response.body?.string().orEmpty()).optJSONObject("error")?.optString("message") }.getOrNull()
                throw ApiFailure(status, message?.takeIf(String::isNotBlank) ?: "请求失败（$status）")
            }
            if (status == 204) return ""
            response.body?.string().orEmpty()
        }
    }
}
