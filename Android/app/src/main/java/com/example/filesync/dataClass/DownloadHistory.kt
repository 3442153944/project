// data/model/DownloadHistory.kt
package com.example.filesync.dataClass

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

/**
 * 下载历史 API 响应
 */
@Serializable
data class DownloadHistoryResponse(
    val code: Int = 0,
    val message: String = "",
    val data: DownloadHistoryData? = null
)

@Serializable
data class DownloadHistoryData(
    val list: List<DownloadHistoryItem> = emptyList(),
    val total: Long = 0,
    val pageNum: Int = 0,
    val pageSize: Int = 0
)

@Serializable
data class DownloadHistoryItem(
    val id: Int = 0,
    @SerialName("user_id") val userId: Int = 0,
    @SerialName("device_id") val deviceId: Int = 0,
    @SerialName("file_id") val fileId: Int? = null,
    @SerialName("file_name") val fileName: String = "",
    @SerialName("file_size") val fileSize: Long = 0,
    @SerialName("download_status") val downloadStatus: String = "",
    @SerialName("download_speed") val downloadSpeed: Long = 0,
    @SerialName("ip_address") val ipAddress: String = "",
    @SerialName("started_at") val startedAt: String? = null,
    @SerialName("completed_at") val completedAt: String? = null,
    @SerialName("created_at") val createdAt: String = ""
)

/**
 * 下载历史请求参数
 */
@Serializable
data class DownloadHistoryRequest(
    val pageNum: Int = 0,
    val pageSize: Int = 0
)