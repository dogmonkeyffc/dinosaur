package com.monkeydog.dinosaur.android

import android.content.Context
import android.security.keystore.KeyGenParameterSpec
import android.security.keystore.KeyProperties
import android.util.Base64
import org.json.JSONArray
import org.json.JSONObject
import java.security.KeyStore
import javax.crypto.Cipher
import javax.crypto.KeyGenerator
import javax.crypto.SecretKey
import javax.crypto.spec.GCMParameterSpec

internal data class Account(val data: JSONObject) {
    val key = data.getString("key")
    val server = data.getJSONObject("server")
    val user = data.getJSONObject("user")
    val origin = serverOrigin(server.getString("url"))
    val token = data.getString("token").also { require(it.matches(Regex("[A-Za-z0-9_-]{43}"))) }
    val expired get() = data.optLong("expires_at") <= System.currentTimeMillis() / 1000
    val api get() = Api(origin, token)
}

internal class Accounts(context: Context) {
    private val prefs = context.getSharedPreferences("native-accounts", Context.MODE_PRIVATE)
    private val legacy = context.getSharedPreferences("SecureStore", Context.MODE_PRIVATE)
    private val store = KeyStore.getInstance("AndroidKeyStore").apply { load(null) }
    private val alias = "dinosaur.native.accounts"
    var active: String?
        get() = prefs.getString("active", null)
        set(value) { prefs.edit().putString("active", value).apply() }

    private fun decode(value: String, alias: String): String {
        val entry = JSONObject(value)
        val cipher = Cipher.getInstance("AES/GCM/NoPadding")
        cipher.init(Cipher.DECRYPT_MODE, store.getKey(alias, null) as SecretKey,
            GCMParameterSpec(entry.getInt("tlen"), Base64.decode(entry.getString("iv"), Base64.DEFAULT)))
        return cipher.doFinal(Base64.decode(entry.getString("ct"), Base64.DEFAULT)).toString(Charsets.UTF_8)
    }

    fun legacyValue(name: String): String? = runCatching {
        val raw = legacy.getString("key_v1-$name", null) ?: legacy.getString(name, null) ?: return null
        val suffix = if (JSONObject(raw).optBoolean("usesKeystoreSuffix")) ":keystoreUnauthenticated" else ""
        decode(raw, "AES/GCM/NoPadding:key_v1$suffix")
    }.getOrNull()

    fun load(): List<Account> {
        val saved = prefs.getString("accounts", null)
        val persisted = saved?.let { JSONArray(decode(it, alias)) }
        val ids = if (persisted == null) legacyValue("dinosaur.accounts")?.let(::JSONArray) else null
        val result = (0 until (persisted?.length() ?: ids?.length() ?: 0).coerceAtMost(12)).mapNotNull { index ->
            runCatching {
                val data = persisted?.getJSONObject(index) ?: run {
                    val key = ids!!.getString(index)
                    require(key.matches(Regex("[a-f0-9]{32}_[0-9]+")))
                    JSONObject(legacyValue("dinosaur.account.$key") ?: error("No account"))
                }
                Account(data).takeUnless { it.expired }
            }.getOrNull()
        }
        if (persisted == null && result.isNotEmpty()) save(result)
        return result
    }

    fun save(accounts: List<Account>) {
        if (!store.containsAlias(alias)) {
            KeyGenerator.getInstance(KeyProperties.KEY_ALGORITHM_AES, "AndroidKeyStore").apply {
                init(KeyGenParameterSpec.Builder(alias, KeyProperties.PURPOSE_ENCRYPT or KeyProperties.PURPOSE_DECRYPT)
                    .setBlockModes(KeyProperties.BLOCK_MODE_GCM)
                    .setEncryptionPaddings(KeyProperties.ENCRYPTION_PADDING_NONE).setKeySize(256).build())
            }.generateKey()
        }
        val cipher = Cipher.getInstance("AES/GCM/NoPadding")
        cipher.init(Cipher.ENCRYPT_MODE, store.getKey(alias, null) as SecretKey)
        val encoded = cipher.doFinal(JSONArray(accounts.map { it.data }).toString().toByteArray(Charsets.UTF_8))
        val record = json("ct" to Base64.encodeToString(encoded, Base64.NO_WRAP),
            "iv" to Base64.encodeToString(cipher.iv, Base64.NO_WRAP), "tlen" to 128)
        check(prefs.edit().putString("accounts", record.toString()).commit()) { "无法保存账号" }
    }
}
