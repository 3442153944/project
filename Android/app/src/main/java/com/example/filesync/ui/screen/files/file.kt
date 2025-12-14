// ui/screen/files/file.kt
package com.example.filesync.ui.screen.files

import android.os.Environment
import androidx.activity.compose.BackHandler
import androidx.compose.foundation.ExperimentalFoundationApi
import androidx.compose.foundation.combinedClickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.lifecycle.viewmodel.compose.viewModel
import com.example.filesync.ui.components.files.*
import com.example.filesync.ui.viewModel.files.ActiveDiskViewModel
import com.example.filesync.ui.viewModel.files.FileListViewModel
import com.example.filesync.ui.viewModel.transmission.DownloadListViewModel
import com.example.filesync.util.RootHelper
import java.io.File

@OptIn(ExperimentalFoundationApi::class)
@Composable
fun FileScreen(
    modifier: Modifier = Modifier
) {
    val context = LocalContext.current
    val diskViewModel = viewModel<ActiveDiskViewModel>()
    val fileListViewModel = viewModel<FileListViewModel>()
    val downloadViewModel = viewModel<DownloadListViewModel>()

    val diskData by diskViewModel.diskData.collectAsState()
    val diskLoading by diskViewModel.loading.collectAsState()
    val diskError by diskViewModel.error.collectAsState()

    val fileData by fileListViewModel.fileData.collectAsState()
    val fileLoading by fileListViewModel.loading.collectAsState()
    val fileError by fileListViewModel.error.collectAsState()
    val pathStack by fileListViewModel.pathStack.collectAsState()

    var currentDiskPath by remember { mutableStateOf<String?>(null) }
    var showDownloadDialog by remember { mutableStateOf(false) }
    var selectedFileForDownload by remember { mutableStateOf<com.example.filesync.ui.viewModel.files.FileItem?>(null) }

    // 判断是否在磁盘根目录
    val isAtDiskRoot = remember(fileData, currentDiskPath) {
        val parentPath = fileData?.parentPath
        parentPath.isNullOrEmpty() || parentPath == currentDiskPath
    }

    // 系统返回键处理
    BackHandler(enabled = currentDiskPath != null) {
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
    }

    // 下载确认对话框
    if (showDownloadDialog && selectedFileForDownload != null) {
        DownloadConfirmDialog(
            fileItem = selectedFileForDownload!!,
            onConfirm = { savePath ->
                val item = selectedFileForDownload!!
                // 获取文件所在目录（去掉文件名）
                val parentPath = item.path.substringBeforeLast(File.separator)

                downloadViewModel.addDownload(
                    path = parentPath,
                    name = item.name,
                    saveDir = File(savePath),
                    deviceId = null
                )

                showDownloadDialog = false
                selectedFileForDownload = null
            },
            onDismiss = {
                showDownloadDialog = false
                selectedFileForDownload = null
            }
        )
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
                            // 目录：进入
                            fileListViewModel.navigateTo(item.path)
                        } else {
                            // 文件：下载
                            selectedFileForDownload = item
                            showDownloadDialog = true
                        }
                    },
                    onLongClick = if (item.isDir) {
                        {
                            // 长按目录：下载整个文件夹
                            selectedFileForDownload = item
                            showDownloadDialog = true
                        }
                    } else null
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

/**
 * 下载确认对话框
 */
@Composable
fun DownloadConfirmDialog(
    fileItem: com.example.filesync.ui.viewModel.files.FileItem,
    onConfirm: (savePath: String) -> Unit,
    onDismiss: () -> Unit
) {
    val context = LocalContext.current

    // 使用 LaunchedEffect 在协程中检查 Root 状态
    var isRooted by remember { mutableStateOf(false) }

    LaunchedEffect(Unit) {
        isRooted = RootHelper.isDeviceRooted()
    }

    // 根据 Root 状态决定默认下载路径
    val defaultDownloadPath = remember(isRooted) {
        if (isRooted) {
            "/sdcard/Download" // Root 模式：从根目录起
        } else {
            Environment.getExternalStoragePublicDirectory(Environment.DIRECTORY_DOWNLOADS).absolutePath
        }
    }

    var savePath by remember(defaultDownloadPath) { mutableStateOf(defaultDownloadPath) }

    AlertDialog(
        onDismissRequest = onDismiss,
        title = {
            Text(if (fileItem.isDir) "下载文件夹" else "下载文件")
        },
        text = {
            Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
                Text("文件名: ${fileItem.name}")
                if (!fileItem.isDir) {
                    Text("大小: ${formatFileSize(fileItem.size)}")
                }

                Spacer(modifier = Modifier.height(8.dp))

                OutlinedTextField(
                    value = savePath,
                    onValueChange = { savePath = it },
                    label = { Text("保存路径") },
                    modifier = Modifier.fillMaxWidth(),
                    singleLine = true
                )

                if (isRooted) {
                    Text(
                        "Root 模式：可以访问系统根目录",
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.primary
                    )
                }
            }
        },
        confirmButton = {
            TextButton(
                onClick = {
                    if (savePath.isNotBlank()) {
                        onConfirm(savePath)
                    }
                }
            ) {
                Text("开始下载")
            }
        },
        dismissButton = {
            TextButton(onClick = onDismiss) {
                Text("取消")
            }
        }
    )
}

/**
 * 格式化文件大小
 */
fun formatFileSize(bytes: Long): String {
    return when {
        bytes < 1024 -> "$bytes B"
        bytes < 1024 * 1024 -> "${bytes / 1024} KB"
        bytes < 1024 * 1024 * 1024 -> "${bytes / (1024 * 1024)} MB"
        else -> "${bytes / (1024 * 1024 * 1024)} GB"
    }
}