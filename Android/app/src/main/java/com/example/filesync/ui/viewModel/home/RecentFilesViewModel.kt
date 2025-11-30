// RecentFilesViewModel.kt
package com.example.filesync.ui.viewmodel.home

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

class RecentFilesViewModel : ViewModel() {

    private val _files = MutableStateFlow<List<RecentFile>>(emptyList())
    val files: StateFlow<List<RecentFile>> = _files.asStateFlow()

    init {
        loadFiles()
    }

    private fun loadFiles() {
        viewModelScope.launch {
            try {
                // TODO: 从数据库加载最近文件
                _files.value = emptyList()
            } catch (e: Exception) {
                // 处理错误
            }
        }
    }

    fun refresh() {
        loadFiles()
    }
}

data class RecentFile(
    val id: String,
    val name: String,
    val path: String,
    val size: Long,
    val lastModified: Long,
    val fileType: FileType
)

enum class FileType {
    DOCUMENT, IMAGE, VIDEO, AUDIO, ARCHIVE, FOLDER, OTHER
}