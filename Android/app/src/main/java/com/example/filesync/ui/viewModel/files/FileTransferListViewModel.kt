// ui/viewModel/files/FileTransferListViewModel.kt
package com.example.filesync.ui.viewModel.files

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.example.filesync.ui.viewModel.transmission.FileTransferStatus
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.launch

class FileTransferListViewModel : ViewModel() {

    private val _transferItems = MutableStateFlow<List<FileTransferItem>>(emptyList())
    val transferItems: StateFlow<List<FileTransferItem>> = _transferItems

    private val _filterStatus = MutableStateFlow<FileTransferStatus?>(null)
    val filterStatus: StateFlow<FileTransferStatus?> = _filterStatus

    private val _sortBy = MutableStateFlow(SortBy.TIME_DESC)
    val sortBy: StateFlow<SortBy> = _sortBy

    init {
        loadTransferHistory()
    }

    fun loadTransferHistory() {
        viewModelScope.launch {
            // TODO: 从数据库或 API 加载传输历史
            // 这里先用模拟数据
            _transferItems.value = emptyList()
        }
    }

    fun addTransferItem(item: FileTransferItem) {
        _transferItems.value = listOf(item) + _transferItems.value
    }

    fun updateTransferProgress(id: Int, progress: Float, speed: Long) {
        _transferItems.value = _transferItems.value.map { item ->
            if (item.id == id) {
                item.copy(progress = progress, speed = speed)
            } else {
                item
            }
        }
    }

    fun updateTransferStatus(id: Int, status: FileTransferStatus) {
        _transferItems.value = _transferItems.value.map { item ->
            if (item.id == id) {
                item.copy(status = status)
            } else {
                item
            }
        }
    }

    fun removeTransferItem(id: Int) {
        _transferItems.value = _transferItems.value.filter { it.id != id }
    }

    fun clearCompleted() {
        _transferItems.value = _transferItems.value.filter {
            it.status != FileTransferStatus.COMPLETED
        }
    }

    fun clearAll() {
        _transferItems.value = emptyList()
    }

    fun setFilter(status: FileTransferStatus?) {
        _filterStatus.value = status
    }

    fun setSortBy(sortBy: SortBy) {
        _sortBy.value = sortBy
    }

    fun getFilteredAndSortedItems(): List<FileTransferItem> {
        var items = _transferItems.value

        // 应用过滤
        _filterStatus.value?.let { status ->
            items = items.filter { it.status == status }
        }

        // 应用排序
        items = when (_sortBy.value) {
            SortBy.TIME_DESC -> items
            SortBy.TIME_ASC -> items.reversed()
            SortBy.NAME_ASC -> items.sortedBy { it.name }
            SortBy.NAME_DESC -> items.sortedByDescending { it.name }
            SortBy.SIZE_ASC -> items.sortedBy { it.size }
            SortBy.SIZE_DESC -> items.sortedByDescending { it.size }
        }

        return items
    }

    fun retryTransfer(id: Int) {
        // TODO: 重试传输
    }

    fun cancelTransfer(id: Int) {
        updateTransferStatus(id, FileTransferStatus.CANCELLED)
    }
    fun pauseTransfer(id: Int) {
        updateTransferStatus(id, FileTransferStatus.PAUSED)
    }
    fun resumeTransfer(id: Int) {
        updateTransferStatus(id, FileTransferStatus.WAITING)
    }
}

data class FileTransferItem(
    val id: Int,
    val name: String,
    val size: Long,
    val isDir: Boolean,
    val childrenCount: Int,
    val progress: Float,
    val speed: Long,
    val status: FileTransferStatus,
    val sourcePath: String = "",
    val targetPath: String = "",
    val startTime: Long = System.currentTimeMillis(),
    val endTime: Long? = null,
    val errorMessage: String? = null
)

enum class SortBy {
    TIME_DESC,
    TIME_ASC,
    NAME_ASC,
    NAME_DESC,
    SIZE_ASC,
    SIZE_DESC
}