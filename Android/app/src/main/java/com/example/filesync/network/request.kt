// network/request.kt
package com.example.filesync.network

import android.content.Context
import android.util.Log
import androidx.datastore.core.DataStore
import androidx.datastore.preferences.core.Preferences
import androidx.datastore.preferences.core.edit
import androidx.datastore.preferences.core.stringPreferencesKey
import androidx.datastore.preferences.core.booleanPreferencesKey
import androidx.datastore.preferences.preferencesDataStore
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.withContext
import kotlinx.serialization.json.Json
import kotlinx.serialization.encodeToString
import kotlinx.serialization.SerializationStrategy
import kotlinx.serialization.serializer
import okhttp3.*
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.RequestBody.Companion.toRequestBody
import java.util.concurrent.TimeUnit

private val Context.dataStore: DataStore<Preferences> by preferencesDataStore(name = "secure_prefs")

object Request {

    var baseUrl = "http://192.168.31.100:9999/api"
        set(value) {
            field = value.trimEnd('/')
            Log.d(TAG, "基础 URL: $field")
        }

    private var appContext: Context? = null

    private val TOKEN_KEY = stringPreferencesKey("token")
    private val USERNAME_KEY = stringPreferencesKey("saved_username")
    private val PASSWORD_KEY = stringPreferencesKey("saved_password")
    private val REMEMBER_PASSWORD_KEY = booleanPreferencesKey("remember_password")

    val client = OkHttpClient.Builder()
        .connectTimeout(30, TimeUnit.SECONDS)
        .readTimeout(30, TimeUnit.SECONDS)
        .writeTimeout(30, TimeUnit.SECONDS)
        .retryOnConnectionFailure(true)
        .build()

    val json = Json {
        ignoreUnknownKeys = true
        isLenient = true
        prettyPrint = false
    }

    fun init(context: Context) {
        appContext = context.applicationContext
        Log.d(TAG, "Request 初始化成功")
    }

    // Token 管理
    suspend fun saveToken(token: String) {
        appContext?.dataStore?.edit { preferences ->
            preferences[TOKEN_KEY] = token
        }
        Log.d(TAG, "Token 已保存")
    }

    suspend fun getToken(): String? {
        return appContext?.dataStore?.data?.map { preferences ->
            preferences[TOKEN_KEY]
        }?.first()
    }

    suspend fun clearToken() {
        appContext?.dataStore?.edit { preferences ->
            preferences.remove(TOKEN_KEY)
        }
        Log.d(TAG, "Token 已清除")
    }

    suspend fun hasToken(): Boolean {
        return getToken() != null
    }

    // 记住密码功能
    suspend fun saveCredentials(username: String, password: String, remember: Boolean) {
        appContext?.dataStore?.edit { preferences ->
            preferences[REMEMBER_PASSWORD_KEY] = remember
            if (remember) {
                preferences[USERNAME_KEY] = username
                preferences[PASSWORD_KEY] = password
            } else {
                preferences.remove(USERNAME_KEY)
                preferences.remove(PASSWORD_KEY)
            }
        }
        Log.d(TAG, "凭据已保存: remember=$remember")
    }

    suspend fun getSavedCredentials(): Triple<String, String, Boolean>? {
        return appContext?.dataStore?.data?.map { preferences ->
            val remember = preferences[REMEMBER_PASSWORD_KEY] ?: false
            val username = preferences[USERNAME_KEY] ?: ""
            val password = preferences[PASSWORD_KEY] ?: ""
            Triple(username, password, remember)
        }?.first()
    }

    suspend fun clearCredentials() {
        appContext?.dataStore?.edit { preferences ->
            preferences.remove(USERNAME_KEY)
            preferences.remove(PASSWORD_KEY)
            preferences.remove(REMEMBER_PASSWORD_KEY)
        }
        Log.d(TAG, "凭据已清除")
    }

    /**
     * POST 请求（主要请求方法）
     */
    suspend inline fun <reified T, reified B> post(
        endpoint: String,
        body: B? = null,
        noinline onResult: (Result<T>) -> Unit = {}
    ) {
        request("POST", endpoint, body, json.serializersModule.serializer<B>(), onResult)
    }

    /**
     * POST 请求（无请求体）
     */
    suspend inline fun <reified T> post(
        endpoint: String,
        noinline onResult: (Result<T>) -> Unit = {}
    ) {
        request<T, Unit>("POST", endpoint, null, null, onResult)
    }

    /**
     * 通用请求方法
     */
    suspend inline fun <reified T, B> request(
        method: String,
        endpoint: String,
        body: B?,
        serializer: SerializationStrategy<B>?,
        noinline onResult: (Result<T>) -> Unit = {}
    ) = withContext(Dispatchers.IO) {
        try {
            val url = "$baseUrl$endpoint"
            Log.d(TAG, "$method $url")

            val token = getToken()

            val requestBuilder = okhttp3.Request.Builder()
                .url(url)
                .apply {
                    token?.let { header("token", it) }
                }

            val requestBody = if (body != null && serializer != null) {
                json.encodeToString(serializer, body).toRequestBody("application/json".toMediaType())
            } else {
                "{}".toRequestBody("application/json".toMediaType())
            }

            when (method.uppercase()) {
                "POST" -> requestBuilder.post(requestBody)
                "PUT" -> requestBuilder.put(requestBody)
                "DELETE" -> requestBuilder.delete(requestBody)
                "GET" -> requestBuilder.get()
            }

            val response = client.newCall(requestBuilder.build()).execute()

            if (response.isSuccessful) {
                val responseBody = response.body?.string()

                if (responseBody != null) {
                    try {
                        val result = json.decodeFromString<T>(responseBody)

                        // 自动提取保存 token
                        try {
                            val dataField = result!!::class.java.getDeclaredField("data")
                            dataField.isAccessible = true
                            val dataValue = dataField.get(result)
                            if (dataValue != null) {
                                val tokenField = dataValue::class.java.getDeclaredField("token")
                                tokenField.isAccessible = true
                                val tokenValue = tokenField.get(dataValue) as? String
                                tokenValue?.let { saveToken(it) }
                            }
                        } catch (_: Exception) {}

                        onResult(Result.success(result))
                    } catch (e: Exception) {
                        Log.e(TAG, "JSON 解析失败: ${e.message}")
                        onResult(Result.failure(Exception("JSON 解析失败: ${e.message}")))
                    }
                } else {
                    onResult(Result.failure(Exception("响应体为空")))
                }
            } else {
                if (response.code == 401) {
                    Log.w(TAG, "认证失败，清除 token")
                    clearToken()
                }
                onResult(Result.failure(Exception("HTTP ${response.code}: ${response.message}")))
            }

        } catch (e: Exception) {
            Log.e(TAG, "请求失败: ${e.message}")
            onResult(Result.failure(e))
        }
    }

    const val TAG = "Request"
}