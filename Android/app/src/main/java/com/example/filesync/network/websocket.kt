// network/websocket.kt
package com.example.filesync.network

import android.util.Log
import kotlinx.coroutines.*
import kotlinx.coroutines.channels.Channel
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.receiveAsFlow
import okhttp3.*
import okio.ByteString
import java.util.concurrent.TimeUnit
import java.util.concurrent.atomic.AtomicBoolean

/**
 * WebSocket 连接管理类（单例）
 *
 * 特点：
 * 1. 全局共享一个连接
 * 2. 自动重连（断线后 1.5 秒重试）
 * 3. 自动清理旧连接
 * 4. 支持文本和二进制消息
 * 5. 协程安全
 *
 * 使用示例：
 * ```kotlin
 * // 初始化
 * WebSocket.init()
 *
 * // 连接（可选传入 token）
 * WebSocket.connect(token = "your-token")
 *
 * // 监听消息
 * LaunchedEffect(Unit) {
 *     WebSocket.messageFlow.collect { message ->
 *         when (message) {
 *             is WsMessage.Text -> println("收到文本: ${message.content}")
 *             is WsMessage.Binary -> println("收到二进制: ${message.data.size} 字节")
 *         }
 *     }
 * }
 *
 * // 监听连接状态
 * LaunchedEffect(Unit) {
 *     WebSocket.stateFlow.collect { state ->
 *         when (state) {
 *             WsState.Connected -> println("已连接")
 *             WsState.Disconnected -> println("已断开")
 *             is WsState.Error -> println("错误: ${state.message}")
 *         }
 *     }
 * }
 *
 * // 发送消息
 * WebSocket.send("Hello Server")
 *
 * // 断开连接
 * WebSocket.disconnect()
 * ```
 */
object WebSocket {

    /** WebSocket 服务器地址 */
    var serverUrl = "ws://192.168.31.100:9999/api/ws/connect"
        private set

    /** OkHttp 客户端（专用于 WebSocket） */
    private val client = OkHttpClient.Builder()
        .connectTimeout(10, TimeUnit.SECONDS)
        .readTimeout(0, TimeUnit.SECONDS)  // WebSocket 长连接，不设置读取超时
        .writeTimeout(10, TimeUnit.SECONDS)
        .pingInterval(20, TimeUnit.SECONDS)  // 每 20 秒发送一次 ping
        .build()

    /** 当前 WebSocket 连接 */
    private var webSocket: okhttp3.WebSocket? = null

    /** 连接 token（用于认证） */
    private var currentToken: String? = null

    /** 是否正在连接中 */
    private val isConnecting = AtomicBoolean(false)

    /** 是否应该自动重连 */
    private val shouldReconnect = AtomicBoolean(false)

    /** 重连协程作用域 */
    private val reconnectScope = CoroutineScope(Dispatchers.IO + SupervisorJob())

    /** 重连任务 */
    private var reconnectJob: Job? = null

    /** 消息通道 */
    private val _messageChannel = Channel<WsMessage>(Channel.UNLIMITED)

    /** 消息流（外部订阅） */
    val messageFlow: Flow<WsMessage> = _messageChannel.receiveAsFlow()

    /** 连接状态通道 */
    private val _stateChannel = Channel<WsState>(Channel.CONFLATED)

    /** 连接状态流（外部订阅） */
    val stateFlow: Flow<WsState> = _stateChannel.receiveAsFlow()

    /**
     * 初始化 WebSocket（可选）
     * 可以在这里做一些初始化配置
     */
    fun init() {
        Log.d(TAG, "WebSocket 模块初始化")
    }

    /**
     * 设置服务器地址
     *
     * @param url WebSocket 服务器地址（如 "ws://192.168.1.100:9999/ws"）
     */
    fun setServerUrl(url: String) {
        serverUrl = url
        Log.d(TAG, "服务器地址: $serverUrl")
    }

    /**
     * 连接到服务器
     *
     * @param token 认证令牌（可选）
     */
    fun connect(token: String? = null) {
        if (isConnecting.get()) {
            Log.w(TAG, "正在连接中，忽略重复请求")
            return
        }

        currentToken = token
        shouldReconnect.set(true)
        doConnect()
    }

    /**
     * 执行连接（内部方法）
     */
    private fun doConnect() {
        if (isConnecting.getAndSet(true)) {
            return
        }

        // 清理旧连接
        cleanupOldConnection()

        try {
            // 构建 URL（如果有 token，附加到查询参数）
            val url = if (currentToken != null) {
                "$serverUrl?token=$currentToken"
            } else {
                serverUrl
            }

            Log.d(TAG, "开始连接: $url")

            val request = okhttp3.Request.Builder()
                .url(url)
                .build()

            webSocket = client.newWebSocket(request, object : WebSocketListener() {

                override fun onOpen(webSocket: okhttp3.WebSocket, response: Response) {
                    isConnecting.set(false)
                    Log.i(TAG, "WebSocket 已连接")
                    _stateChannel.trySend(WsState.Connected)

                    // 连接成功，取消重连任务
                    cancelReconnect()
                }

                override fun onMessage(webSocket: okhttp3.WebSocket, text: String) {
                    Log.d(TAG, "收到文本消息: $text")
                    _messageChannel.trySend(WsMessage.Text(text))
                }

                override fun onMessage(webSocket: okhttp3.WebSocket, bytes: ByteString) {
                    Log.d(TAG, "收到二进制消息: ${bytes.size} 字节")
                    _messageChannel.trySend(WsMessage.Binary(bytes.toByteArray()))
                }

                override fun onClosing(webSocket: okhttp3.WebSocket, code: Int, reason: String) {
                    Log.d(TAG, "连接关闭中: $code - $reason")
                }

                override fun onClosed(webSocket: okhttp3.WebSocket, code: Int, reason: String) {
                    isConnecting.set(false)
                    Log.i(TAG, "连接已关闭: $code - $reason")
                    _stateChannel.trySend(WsState.Disconnected)

                    // 如果应该重连，启动重连
                    if (shouldReconnect.get()) {
                        scheduleReconnect()
                    }
                }

                override fun onFailure(
                    webSocket: okhttp3.WebSocket,
                    t: Throwable,
                    response: Response?
                ) {
                    isConnecting.set(false)
                    val errorMsg = t.message ?: "未知错误"
                    Log.e(TAG, "连接失败: $errorMsg", t)
                    _stateChannel.trySend(WsState.Error(errorMsg))

                    // 如果应该重连，启动重连
                    if (shouldReconnect.get()) {
                        scheduleReconnect()
                    }
                }
            })

        } catch (e: Exception) {
            isConnecting.set(false)
            Log.e(TAG, "连接异常: ${e.message}", e)
            _stateChannel.trySend(WsState.Error(e.message ?: "连接异常"))

            if (shouldReconnect.get()) {
                scheduleReconnect()
            }
        }
    }

    /**
     * 安排重连任务
     * 1.5 秒后重试
     */
    private fun scheduleReconnect() {
        // 取消之前的重连任务
        cancelReconnect()

        Log.d(TAG, "将在 1.5 秒后重连...")

        reconnectJob = reconnectScope.launch {
            delay(1500)  // 等待 1.5 秒

            if (shouldReconnect.get()) {
                Log.d(TAG, "开始重连...")
                doConnect()
            }
        }
    }

    /**
     * 取消重连任务
     */
    private fun cancelReconnect() {
        reconnectJob?.cancel()
        reconnectJob = null
    }

    /**
     * 清理旧连接（必须在新连接前调用）
     */
    private fun cleanupOldConnection() {
        webSocket?.let {
            try {
                Log.d(TAG, "清理旧连接")
                it.close(1000, "创建新连接")
            } catch (e: Exception) {
                Log.e(TAG, "清理旧连接失败: ${e.message}")
            }
        }
        webSocket = null
    }

    /**
     * 发送文本消息
     *
     * @param message 文本内容
     * @return true 表示发送成功，false 表示未连接
     */
    fun send(message: String): Boolean {
        val ws = webSocket
        if (ws == null) {
            Log.w(TAG, "未连接，无法发送消息")
            return false
        }

        return try {
            val success = ws.send(message)
            if (success) {
                Log.d(TAG, "发送文本: $message")
            } else {
                Log.w(TAG, "发送失败（队列满或连接关闭）")
            }
            success
        } catch (e: Exception) {
            Log.e(TAG, "发送消息异常: ${e.message}")
            false
        }
    }

    /**
     * 发送二进制消息
     *
     * @param data 二进制数据
     * @return true 表示发送成功
     */
    fun send(data: ByteArray): Boolean {
        val ws = webSocket
        if (ws == null) {
            Log.w(TAG, "未连接，无法发送数据")
            return false
        }

        return try {
            val success = ws.send(ByteString.of(*data))
            if (success) {
                Log.d(TAG, "发送二进制: ${data.size} 字节")
            } else {
                Log.w(TAG, "发送失败（队列满或连接关闭）")
            }
            success
        } catch (e: Exception) {
            Log.e(TAG, "发送数据异常: ${e.message}")
            false
        }
    }

    /**
     * 断开连接
     * 停止自动重连
     */
    fun disconnect() {
        Log.i(TAG, "主动断开连接")

        // 停止自动重连
        shouldReconnect.set(false)
        cancelReconnect()

        // 清理连接
        cleanupOldConnection()

        _stateChannel.trySend(WsState.Disconnected)
    }

    /**
     * 检查是否已连接
     */
    fun isConnected(): Boolean {
        return webSocket != null
    }

    /**
     * 获取当前连接状态
     */
    fun getConnectionState(): WsState {
        return when {
            isConnecting.get() -> WsState.Connecting
            webSocket != null -> WsState.Connected
            else -> WsState.Disconnected
        }
    }

    private const val TAG = "WebSocket"
}

/**
 * WebSocket 消息类型
 */
sealed class WsMessage {
    /** 文本消息 */
    data class Text(val content: String) : WsMessage()

    /** 二进制消息 */
    data class Binary(val data: ByteArray) : WsMessage() {
        override fun equals(other: Any?): Boolean {
            if (this === other) return true
            if (javaClass != other?.javaClass) return false
            other as Binary
            return data.contentEquals(other.data)
        }

        override fun hashCode(): Int {
            return data.contentHashCode()
        }
    }
}

/**
 * WebSocket 连接状态
 */
sealed class WsState {
    /** 连接中 */
    data object Connecting : WsState()

    /** 已连接 */
    data object Connected : WsState()

    /** 已断开 */
    data object Disconnected : WsState()

    /** 错误 */
    data class Error(val message: String) : WsState()
}