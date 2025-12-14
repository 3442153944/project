// ui/screen/HomeScreen.kt
package com.example.filesync.ui.screen

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.navigation.NavController
import com.example.filesync.router.AppRoute
import com.example.filesync.router.navigateToDetail
import com.example.filesync.ui.components.home.*
import com.example.filesync.ui.screen.transmission.MicroTransmissionCard

@Composable
fun HomeScreen(
    navController: NavController,
    modifier: Modifier = Modifier
) {
    val downloadingCount by remember { mutableIntStateOf(2) }
    val uploadingCount by remember { mutableIntStateOf(1) }

    LazyColumn(
        modifier = modifier
            .fillMaxSize()
            .padding(16.dp),
        verticalArrangement = Arrangement.spacedBy(16.dp)
    ) {
        // 标题 + 运行模式标签
        item {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Column(modifier = Modifier.weight(1f)) {
                    Text(
                        text = "文件同步",
                        fontSize = 28.sp,
                        fontWeight = FontWeight.Bold
                    )
                    Text(
                        text = "一切正常运行中",
                        fontSize = 14.sp,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }
                RunningModeBadge()
            }
        }

        // 传输列表卡片 - 点击打开传输页面
        item {
            MicroTransmissionCard(
                onClick = {
                    navController.navigateToDetail(AppRoute.Transfer)
                },
                downloadingCount = downloadingCount,
                uploadingCount = uploadingCount
            )
        }

        // 同步状态
        item { SyncStatusCard() }

        // 存储空间
        item { StorageCard() }

        // 快速操作
        item {
            QuickActionsSection(
                onUploadClick = {
                    navController.navigateToDetail(AppRoute.FileUpload)
                },
                onSearchClick = {
                    navController.navigateToDetail(AppRoute.FileSearch)
                }
            )
        }

        // 最近文件
        item {
            RecentFilesList(
                onFileClick = { fileId ->
                    navController.navigateToDetail(
                        AppRoute.FileDetail.createRoute(fileId)
                    )
                }
            )
        }

        // 已连接设备
        item { DevicesList() }

        item { Spacer(modifier = Modifier.height(16.dp)) }
    }
}

@Composable
fun QuickActionsSection(
    onUploadClick: () -> Unit = {},
    onSearchClick: () -> Unit = {}
) {
    Column(verticalArrangement = Arrangement.spacedBy(12.dp)) {
        Text(
            text = "快速操作",
            fontSize = 18.sp,
            fontWeight = FontWeight.SemiBold
        )
        Row(
            horizontalArrangement = Arrangement.spacedBy(8.dp)
        ) {
            FilledTonalButton(onClick = onUploadClick) {
                Text("上传文件")
            }
            FilledTonalButton(onClick = onSearchClick) {
                Text("搜索文件")
            }
        }
    }
}

@Composable
fun RecentFilesList(
    onFileClick: (String) -> Unit = {}
) {
    Column(verticalArrangement = Arrangement.spacedBy(12.dp)) {
        Text(
            text = "最近文件",
            fontSize = 18.sp,
            fontWeight = FontWeight.SemiBold
        )
        // 这里可以显示文件列表
        Text(
            text = "暂无文件",
            fontSize = 14.sp,
            color = MaterialTheme.colorScheme.onSurfaceVariant
        )
    }
}

@Composable
fun DevicesList() {
    Column(verticalArrangement = Arrangement.spacedBy(12.dp)) {
        Text(
            text = "已连接设备",
            fontSize = 18.sp,
            fontWeight = FontWeight.SemiBold
        )
        Text(
            text = "暂无设备",
            fontSize = 14.sp,
            color = MaterialTheme.colorScheme.onSurfaceVariant
        )
    }
}