// SyncStatusCard.kt
package com.example.filesync.ui.components.home

import androidx.compose.foundation.layout.*
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.lifecycle.viewmodel.compose.viewModel
import com.example.filesync.ui.viewmodel.home.*

@Composable
fun SyncStatusCard(
    viewModel: SyncStatusViewModel = viewModel()
) {
    val status by viewModel.status.collectAsState()

    ElevatedCard(
        modifier = Modifier.fillMaxWidth(),
        colors = CardDefaults.elevatedCardColors(
            containerColor = MaterialTheme.colorScheme.primaryContainer
        )
    ) {
        Column(
            modifier = Modifier.padding(20.dp),
            verticalArrangement = Arrangement.spacedBy(12.dp)
        ) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text(
                    text = "同步状态",
                    fontSize = 18.sp,
                    fontWeight = FontWeight.SemiBold
                )
                Icon(
                    imageVector = Icons.Default.Sync,
                    contentDescription = null,
                    tint = MaterialTheme.colorScheme.primary
                )
            }

            when (status) {
                is SyncStatus.Idle -> {
                    Text("空闲中")
                    Button(
                        onClick = { viewModel.startSync() },
                        modifier = Modifier.fillMaxWidth()
                    ) {
                        Text("开始同步")
                    }
                }
                is SyncStatus.Syncing -> {
                    SyncingContent(status as SyncStatus.Syncing, viewModel)
                }
                is SyncStatus.Success -> {
                    Row(
                        verticalAlignment = Alignment.CenterVertically,
                        horizontalArrangement = Arrangement.spacedBy(8.dp)
                    ) {
                        Icon(
                            imageVector = Icons.Default.CheckCircle,
                            contentDescription = null,
                            tint = MaterialTheme.colorScheme.primary
                        )
                        Text("同步完成")
                    }
                }
                is SyncStatus.Failed -> {
                    Row(
                        verticalAlignment = Alignment.CenterVertically,
                        horizontalArrangement = Arrangement.spacedBy(8.dp)
                    ) {
                        Icon(
                            imageVector = Icons.Default.Error,
                            contentDescription = null,
                            tint = MaterialTheme.colorScheme.error
                        )
                        Text("同步失败: ${(status as SyncStatus.Failed).error}")
                    }
                }
            }
        }
    }
}

@Composable
private fun SyncingContent(status: SyncStatus.Syncing, viewModel: SyncStatusViewModel) {
    Row(
        modifier = Modifier.fillMaxWidth(),
        horizontalArrangement = Arrangement.SpaceBetween
    ) {
        Text("进行中的任务")
        Text(
            text = "${status.activeTaskCount}个",
            fontWeight = FontWeight.Bold
        )
    }

    Row(
        modifier = Modifier.fillMaxWidth(),
        horizontalArrangement = Arrangement.SpaceBetween
    ) {
        Row(horizontalArrangement = Arrangement.spacedBy(4.dp)) {
            Icon(
                imageVector = Icons.Default.ArrowUpward,
                contentDescription = null,
                modifier = Modifier.size(16.dp)
            )
            Text("上传: %.1f MB/s".format(status.uploadSpeed))
        }
        Row(horizontalArrangement = Arrangement.spacedBy(4.dp)) {
            Icon(
                imageVector = Icons.Default.ArrowDownward,
                contentDescription = null,
                modifier = Modifier.size(16.dp)
            )
            Text("下载: %.1f MB/s".format(status.downloadSpeed))
        }
    }

    Column(verticalArrangement = Arrangement.spacedBy(4.dp)) {
        LinearProgressIndicator(
            progress = { status.progress / 100f },
            modifier = Modifier.fillMaxWidth()
        )
        Text(
            text = "总体进度: ${status.progress}%",
            fontSize = 12.sp,
            color = MaterialTheme.colorScheme.onSurfaceVariant
        )
    }

    Button(
        onClick = { viewModel.stopSync() },
        modifier = Modifier.fillMaxWidth(),
        colors = ButtonDefaults.buttonColors(
            containerColor = MaterialTheme.colorScheme.error
        )
    ) {
        Text("停止同步")
    }
}