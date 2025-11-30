// DevicesViewModel.kt
package com.example.filesync.ui.viewmodel.home

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

class DevicesViewModel : ViewModel() {

    private val _devices = MutableStateFlow<List<ConnectedDevice>>(emptyList())
    val devices: StateFlow<List<ConnectedDevice>> = _devices.asStateFlow()

    init {
        loadDevices()
    }

    private fun loadDevices() {
        viewModelScope.launch {
            try {
                // TODO: 从网络加载设备列表
                _devices.value = emptyList()
            } catch (e: Exception) {
                // 处理错误
            }
        }
    }

    fun refresh() {
        loadDevices()
    }
}

data class ConnectedDevice(
    val id: String,
    val name: String,
    val ip: String,
    val deviceType: DeviceType,
    val isOnline: Boolean,
    val lastSeen: Long
)

enum class DeviceType {
    COMPUTER, SMARTPHONE, TABLET, SERVER, OTHER
}