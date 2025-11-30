// SyncStatusViewModel.kt
package com.example.filesync.ui.viewmodel.home

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

class SyncStatusViewModel : ViewModel() {

    private val _status = MutableStateFlow<SyncStatus>(SyncStatus.Idle)
    val status: StateFlow<SyncStatus> = _status.asStateFlow()

    fun startSync() {
        viewModelScope.launch {
            try {
                _status.value = SyncStatus.Syncing(
                    progress = 0,
                    uploadSpeed = 0.0,
                    downloadSpeed = 0.0,
                    activeTaskCount = 0
                )
                // TODO: 实现同步逻辑
            } catch (e: Exception) {
                _status.value = SyncStatus.Failed(e.message ?: "同步失败")
            }
        }
    }

    fun stopSync() {
        viewModelScope.launch {
            _status.value = SyncStatus.Idle
            // TODO: 停止同步
        }
    }
}

sealed class SyncStatus {
    object Idle : SyncStatus()
    data class Syncing(
        val progress: Int,
        val uploadSpeed: Double,
        val downloadSpeed: Double,
        val activeTaskCount: Int
    ) : SyncStatus()
    object Success : SyncStatus()
    data class Failed(val error: String) : SyncStatus()
}