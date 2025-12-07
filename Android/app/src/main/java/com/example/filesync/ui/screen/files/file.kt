// ui/screen/files/file.kt
package com.example.filesync.ui.screen.files

import androidx.activity.compose.BackHandler
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.Text
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.lifecycle.viewmodel.compose.viewModel
import com.example.filesync.ui.components.files.*
import com.example.filesync.ui.viewModel.files.ActiveDiskViewModel
import com.example.filesync.ui.viewModel.files.FileListViewModel

@Composable
fun FileScreen(
    modifier: Modifier = Modifier
) {
    val diskViewModel = viewModel<ActiveDiskViewModel>()
    val fileListViewModel = viewModel<FileListViewModel>()

    val diskData by diskViewModel.diskData.collectAsState()
    val diskLoading by diskViewModel.loading.collectAsState()
    val diskError by diskViewModel.error.collectAsState()

    val fileData by fileListViewModel.fileData.collectAsState()
    val fileLoading by fileListViewModel.loading.collectAsState()
    val fileError by fileListViewModel.error.collectAsState()
    val pathStack by fileListViewModel.pathStack.collectAsState()

    var currentDiskPath by remember { mutableStateOf<String?>(null) }

    // 判断是否在磁盘根目录
    val isAtDiskRoot = remember(fileData, currentDiskPath) {
        val parentPath = fileData?.parentPath
        parentPath.isNullOrEmpty() || parentPath == currentDiskPath
    }

    // 系统返回键处理
    BackHandler(enabled = currentDiskPath != null) {
        when {
            pathStack.isNotEmpty() -> {
                // 有历史记录，返回上一个目录
                fileListViewModel.navigateBack()
            }
            !isAtDiskRoot && fileData?.parentPath?.isNotEmpty() == true -> {
                // 不在根目录，返回上级
                fileListViewModel.navigateToParent()
            }
            else -> {
                // 在根目录，返回磁盘列表
                currentDiskPath = null
                fileListViewModel.clearState()
            }
        }
    }

    if (currentDiskPath != null) {
        LaunchedEffect(currentDiskPath) {
            fileListViewModel.loadDirectory(currentDiskPath!!)
        }

        LazyColumn(
            modifier = modifier
                .fillMaxSize()
                .padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(8.dp)
        ) {
            item {
                FileListHeader(
                    currentPath = fileData?.currentPath ?: currentDiskPath!!,
                    canGoBack = true,
                    loading = fileLoading,
                    onBack = {
                        when {
                            pathStack.isNotEmpty() -> {
                                fileListViewModel.navigateBack()
                            }
                            !isAtDiskRoot && fileData?.parentPath?.isNotEmpty() == true -> {
                                fileListViewModel.navigateToParent()
                            }
                            else -> {
                                currentDiskPath = null
                                fileListViewModel.clearState()
                            }
                        }
                    },
                    onRefresh = { fileListViewModel.refresh() }
                )
            }

            fileData?.let { data ->
                item {
                    FileListStats(
                        totalCount = data.totalCount,
                        dirCount = data.dirCount,
                        fileCount = data.fileCount
                    )
                }
            }

            if (fileLoading) {
                item { LoadingIndicator() }
            }

            if (fileError != null) {
                item { ErrorCard(message = fileError!!) }
            }

            val fileItems = fileData?.items ?: emptyList()
            val sortedItems = fileItems.sortedWith(compareBy({ !it.isDir }, { it.name.lowercase() }))

            items(sortedItems, key = { it.path }) { item ->
                FileItemCard(
                    item = item,
                    onClick = {
                        if (item.isDir) {
                            fileListViewModel.navigateTo(item.path)
                        }
                    }
                )
            }
        }
    } else {
        LazyColumn(
            modifier = modifier
                .fillMaxSize()
                .padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(12.dp)
        ) {
            item {
                DiskListHeader(
                    title = "可用磁盘 (${diskData?.allowedCount ?: 0})",
                    loading = diskLoading,
                    onRefresh = { diskViewModel.loadDisks() }
                )
            }

            if (diskLoading) {
                item { LoadingIndicator() }
            }

            if (diskError != null) {
                item { ErrorCard(message = diskError!!) }
            }

            diskItems(
                disks = diskData?.allowedDisks ?: emptyList(),
                onDiskClick = { disk ->
                    currentDiskPath = disk.path
                }
            )

            val disabledDisks = diskData?.allDisks?.filter { !it.isAllowed } ?: emptyList()
            if (disabledDisks.isNotEmpty()) {
                item {
                    Text(
                        text = "其他磁盘",
                        fontSize = 16.sp,
                        fontWeight = FontWeight.Medium,
                        modifier = Modifier.padding(top = 8.dp)
                    )
                }

                diskItems(
                    disks = disabledDisks,
                    onDiskClick = null
                )
            }
        }
    }
}