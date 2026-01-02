// ui/screen/files/FileTransferListScreen.kt
package com.example.filesync.ui.screen.files

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.lifecycle.viewmodel.compose.viewModel
import androidx.navigation.NavController
import com.example.filesync.ui.viewModel.files.*
import com.example.filesync.ui.viewModel.transmission.FileTransferStatus

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun FileTransferListScreen(
    onBackClick: () -> Unit,
    modifier: Modifier = Modifier,
    navController: NavController,
    viewModel: FileTransferListViewModel = viewModel()
) {
    val transferItems by viewModel.transferItems.collectAsState()
    val filterStatus by viewModel.filterStatus.collectAsState()
    val sortBy by viewModel.sortBy.collectAsState()

    var showFilterMenu by remember { mutableStateOf(false) }
    var showSortMenu by remember { mutableStateOf(false) }

    val filteredItems = remember(transferItems, filterStatus, sortBy) {
        viewModel.getFilteredAndSortedItems()
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("传输列表") },
                navigationIcon = {
                    IconButton(onClick = onBackClick) {
                        Icon(Icons.Default.ArrowBack, "返回")
                    }
                },
                actions = {
                    // 筛选按钮
                    Box {
                        IconButton(onClick = { showFilterMenu = true }) {
                            Icon(Icons.Default.FilterList, "筛选")
                        }
                        DropdownMenu(
                            expanded = showFilterMenu,
                            onDismissRequest = { showFilterMenu = false }
                        ) {
                            DropdownMenuItem(
                                text = { Text("全部") },
                                onClick = {
                                    viewModel.setFilter(null)
                                    showFilterMenu = false
                                },
                                leadingIcon = {
                                    if (filterStatus == null) {
                                        Icon(Icons.Default.Check, null)
                                    }
                                }
                            )
                            FileTransferStatus.entries.forEach { status ->
                                DropdownMenuItem(
                                    text = { Text(status.displayName) },
                                    onClick = {
                                        viewModel.setFilter(status)
                                        showFilterMenu = false
                                    },
                                    leadingIcon = {
                                        if (filterStatus == status) {
                                            Icon(Icons.Default.Check, null)
                                        }
                                    }
                                )
                            }
                        }
                    }

                    // 排序按钮
                    Box {
                        IconButton(onClick = { showSortMenu = true }) {
                            Icon(Icons.Default.Sort, "排序")
                        }
                        DropdownMenu(
                            expanded = showSortMenu,
                            onDismissRequest = { showSortMenu = false }
                        ) {
                            DropdownMenuItem(
                                text = { Text("时间 ↓") },
                                onClick = {
                                    viewModel.setSortBy(SortBy.TIME_DESC)
                                    showSortMenu = false
                                }
                            )
                            DropdownMenuItem(
                                text = { Text("时间 ↑") },
                                onClick = {
                                    viewModel.setSortBy(SortBy.TIME_ASC)
                                    showSortMenu = false
                                }
                            )
                            DropdownMenuItem(
                                text = { Text("名称 ↑") },
                                onClick = {
                                    viewModel.setSortBy(SortBy.NAME_ASC)
                                    showSortMenu = false
                                }
                            )
                            DropdownMenuItem(
                                text = { Text("名称 ↓") },
                                onClick = {
                                    viewModel.setSortBy(SortBy.NAME_DESC)
                                    showSortMenu = false
                                }
                            )
                            DropdownMenuItem(
                                text = { Text("大小 ↑") },
                                onClick = {
                                    viewModel.setSortBy(SortBy.SIZE_ASC)
                                    showSortMenu = false
                                }
                            )
                            DropdownMenuItem(
                                text = { Text("大小 ↓") },
                                onClick = {
                                    viewModel.setSortBy(SortBy.SIZE_DESC)
                                    showSortMenu = false
                                }
                            )
                        }
                    }

                    // 更多操作
                    Box {
                        var showClearMenu by remember { mutableStateOf(false) }
                        IconButton(onClick = { showClearMenu = true }) {
                            Icon(Icons.Default.MoreVert, "更多")
                        }
                        DropdownMenu(
                            expanded = showClearMenu,
                            onDismissRequest = { showClearMenu = false }
                        ) {
                            DropdownMenuItem(
                                text = { Text("清空已完成") },
                                onClick = {
                                    viewModel.clearCompleted()
                                    showClearMenu = false
                                }
                            )
                            DropdownMenuItem(
                                text = { Text("清空全部") },
                                onClick = {
                                    viewModel.clearAll()
                                    showClearMenu = false
                                }
                            )
                        }
                    }
                }
            )
        }
    ) { padding ->
        if (filteredItems.isEmpty()) {
            Box(
                modifier = modifier
                    .fillMaxSize()
                    .padding(padding),
                contentAlignment = Alignment.Center
            ) {
                Column(
                    horizontalAlignment = Alignment.CenterHorizontally,
                    verticalArrangement = Arrangement.spacedBy(8.dp)
                ) {
                    Icon(
                        Icons.Default.CloudQueue,
                        contentDescription = null,
                        modifier = Modifier.size(64.dp),
                        tint = MaterialTheme.colorScheme.outline
                    )
                    Text(
                        "暂无传输记录",
                        color = MaterialTheme.colorScheme.outline
                    )
                }
            }
        } else {
            LazyColumn(
                modifier = modifier
                    .fillMaxSize()
                    .padding(padding),
                contentPadding = PaddingValues(16.dp),
                verticalArrangement = Arrangement.spacedBy(8.dp)
            ) {
                items(filteredItems, key = { it.id }) { item ->
                    TransferItemCard(
                        item = item,
                        onRetry = { viewModel.retryTransfer(item.id) },
                        onCancel = { viewModel.cancelTransfer(item.id) },
                        onPause = { viewModel.pauseTransfer(item.id) },
                        onResume = { viewModel.resumeTransfer(item.id) },
                        onRemove = { viewModel.removeTransferItem(item.id) }
                    )
                }
            }
        }
    }
}

@Composable
fun TransferItemCard(
    item: FileTransferItem,
    onRetry: () -> Unit,
    onCancel: () -> Unit,
    onPause: () -> Unit,
    onResume: () -> Unit,
    onRemove: () -> Unit
) {
    Card(
        modifier = Modifier.fillMaxWidth()
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(8.dp)
        ) {
            // 文件信息行
            Row(
                modifier = Modifier.fillMaxWidth(),
                verticalAlignment = Alignment.CenterVertically
            ) {
                Icon(
                    if (item.isDir) Icons.Default.Folder else Icons.Default.InsertDriveFile,
                    contentDescription = null,
                    modifier = Modifier.size(40.dp),
                    tint = when (item.status) {
                        FileTransferStatus.COMPLETED -> MaterialTheme.colorScheme.tertiary
                        FileTransferStatus.FAILED -> MaterialTheme.colorScheme.error
                        else -> MaterialTheme.colorScheme.primary
                    }
                )
                Spacer(Modifier.width(12.dp))
                Column(modifier = Modifier.weight(1f)) {
                    Text(
                        item.name,
                        style = MaterialTheme.typography.bodyLarge,
                        fontWeight = FontWeight.Medium
                    )
                    Row(
                        horizontalArrangement = Arrangement.spacedBy(8.dp)
                    ) {
                        Text(
                            formatFileSize(item.size),
                            style = MaterialTheme.typography.bodySmall,
                            color = MaterialTheme.colorScheme.onSurfaceVariant
                        )
                        if (item.isDir && item.childrenCount > 0) {
                            Text(
                                "• ${item.childrenCount} 项",
                                style = MaterialTheme.typography.bodySmall,
                                color = MaterialTheme.colorScheme.onSurfaceVariant
                            )
                        }
                    }
                }

                // 状态标签
                TransferStatusChip(item.status)
            }

            // 进度条（传输中或暂停时显示）
            if (item.status.showProgress) {
                Column(verticalArrangement = Arrangement.spacedBy(4.dp)) {
                    Row(
                        modifier = Modifier.fillMaxWidth(),
                        horizontalArrangement = Arrangement.SpaceBetween
                    ) {
                        Text(
                            "${(item.progress * 100).toInt()}%",
                            style = MaterialTheme.typography.labelSmall
                        )
                        if (item.status == FileTransferStatus.TRANSFERRING) {
                            Text(
                                formatSpeed(item.speed),
                                style = MaterialTheme.typography.labelSmall
                            )
                        } else {
                            Text(
                                "已暂停",
                                style = MaterialTheme.typography.labelSmall,
                                color = MaterialTheme.colorScheme.onSurfaceVariant
                            )
                        }
                    }
                    LinearProgressIndicator(
                        progress = { item.progress },
                        modifier = Modifier.fillMaxWidth()
                    )
                }
            }

            // 错误信息
            item.errorMessage?.let { error ->
                if (item.status == FileTransferStatus.FAILED) {
                    Text(
                        error,
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.error
                    )
                }
            }

            // 操作按钮
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.End,
                verticalAlignment = Alignment.CenterVertically
            ) {
                when {
                    // 失败：显示重试按钮
                    item.status.canRetry -> {
                        TextButton(onClick = onRetry) {
                            Icon(Icons.Default.Refresh, null, Modifier.size(18.dp))
                            Spacer(Modifier.width(4.dp))
                            Text("重试")
                        }
                    }
                    // 暂停：显示恢复按钮
                    item.status == FileTransferStatus.PAUSED -> {
                        TextButton(onClick = onResume) {
                            Icon(Icons.Default.PlayArrow, null, Modifier.size(18.dp))
                            Spacer(Modifier.width(4.dp))
                            Text("恢复")
                        }
                    }
                    // 传输中：显示暂停按钮
                    item.status.canPause -> {
                        TextButton(onClick = onPause) {
                            Icon(Icons.Default.Pause, null, Modifier.size(18.dp))
                            Spacer(Modifier.width(4.dp))
                            Text("暂停")
                        }
                    }
                }

                // 可以取消的状态：显示取消按钮
                if (item.status.canCancel) {
                    TextButton(onClick = onCancel) {
                        Icon(Icons.Default.Cancel, null, Modifier.size(18.dp))
                        Spacer(Modifier.width(4.dp))
                        Text("取消")
                    }
                }

                // 可以删除记录的状态：显示删除按钮
                if (item.status.canDelete) {
                    IconButton(onClick = onRemove) {
                        Icon(Icons.Default.Delete, "删除记录")
                    }
                }
            }
        }
    }
}

@Composable
fun TransferStatusChip(status: FileTransferStatus) {
    val (backgroundColor, contentColor, icon) = when (status) {
        FileTransferStatus.WAITING -> Triple(
            MaterialTheme.colorScheme.secondaryContainer,
            MaterialTheme.colorScheme.onSecondaryContainer,
            Icons.Default.Schedule
        )
        FileTransferStatus.TRANSFERRING -> Triple(
            MaterialTheme.colorScheme.primaryContainer,
            MaterialTheme.colorScheme.onPrimaryContainer,
            Icons.Default.CloudUpload
        )
        FileTransferStatus.PAUSED -> Triple(
            MaterialTheme.colorScheme.surfaceVariant,
            MaterialTheme.colorScheme.onSurfaceVariant,
            Icons.Default.Pause
        )
        FileTransferStatus.COMPLETED -> Triple(
            MaterialTheme.colorScheme.tertiaryContainer,
            MaterialTheme.colorScheme.onTertiaryContainer,
            Icons.Default.CheckCircle
        )
        FileTransferStatus.FAILED -> Triple(
            MaterialTheme.colorScheme.errorContainer,
            MaterialTheme.colorScheme.onErrorContainer,
            Icons.Default.Error
        )
        FileTransferStatus.CANCELLED -> Triple(
            MaterialTheme.colorScheme.surfaceVariant,
            MaterialTheme.colorScheme.onSurfaceVariant,
            Icons.Default.Cancel
        )
    }

    Surface(
        shape = RoundedCornerShape(12.dp),
        color = backgroundColor
    ) {
        Row(
            modifier = Modifier.padding(horizontal = 8.dp, vertical = 4.dp),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(4.dp)
        ) {
            Icon(
                icon,
                contentDescription = null,
                modifier = Modifier.size(16.dp),
                tint = contentColor
            )
            Text(
                status.displayName,
                style = MaterialTheme.typography.labelSmall,
                color = contentColor
            )
        }
    }
}

fun formatSpeed(bytesPerSecond: Long): String {
    return when {
        bytesPerSecond < 1024 -> "$bytesPerSecond B/s"
        bytesPerSecond < 1024 * 1024 -> "${bytesPerSecond / 1024} KB/s"
        else -> "${bytesPerSecond / (1024 * 1024)} MB/s"
    }
}