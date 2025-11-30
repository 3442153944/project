// StorageViewModel.kt
package com.example.filesync.ui.viewmodel.home

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

class StorageViewModel : ViewModel() {

    private val _info = MutableStateFlow(StorageInfo())
    val info: StateFlow<StorageInfo> = _info.asStateFlow()

    init {
        loadStorage()
    }

    private fun loadStorage() {
        viewModelScope.launch {
            try {
                // TODO: 获取存储空间信息
                _info.value = StorageInfo(
                    totalBytes = 0L,
                    usedBytes = 0L
                )
            } catch (e: Exception) {
                // 处理错误
            }
        }
    }

    fun refresh() {
        loadStorage()
    }
}

data class StorageInfo(
    val totalBytes: Long = 0L,
    val usedBytes: Long = 0L
) {
    val usedPercentage: Float
        get() = if (totalBytes > 0) usedBytes.toFloat() / totalBytes else 0f

    fun formatSize(bytes: Long): String {
        return when {
            bytes < 1024 -> "$bytes B"
            bytes < 1024 * 1024 -> "%.1f KB".format(bytes / 1024.0)
            bytes < 1024 * 1024 * 1024 -> "%.1f MB".format(bytes / (1024.0 * 1024))
            else -> "%.1f GB".format(bytes / (1024.0 * 1024 * 1024))
        }
    }
}