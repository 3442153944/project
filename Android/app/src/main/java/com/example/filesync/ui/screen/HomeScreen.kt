package com.example.filesync.ui.screen

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp

@Composable
fun HomeScreen(modifier: Modifier = Modifier) {
    LazyColumn(
        modifier = modifier
            .fillMaxSize()
            .padding(16.dp),
        verticalArrangement = Arrangement.spacedBy(16.dp)
    ) {
        // 欢迎标题
        item {
            Column {
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
        }

        // 同步状态卡片
        item {
            SyncStatusCard()
        }

        // 存储空间卡片
        item {
            StorageCard()
        }

        // 快速操作
        item {
            QuickActionsSection()
        }

        // 最近文件
        item {
            SectionHeader(
                title = "最近文件",
                onViewAll = { /* 跳转到文件页面 */ }
            )
        }

        items(getRecentFiles()) { file ->
            RecentFileItem(file)
        }

        // 已连接设备
        item {
            SectionHeader(
                title = "已连接设备",
                onViewAll = { /* 跳转到设备管理 */ }
            )
        }

        items(getConnectedDevices()) { device ->
            DeviceItem(device)
        }

        // 底部间距
        item {
            Spacer(modifier = Modifier.height(16.dp))
        }
    }
}

/**
 * 同步状态卡片
 */
@Composable
fun SyncStatusCard() {
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

            // 任务数量
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween
            ) {
                Text("进行中的任务")
                Text(
                    text = "3个",
                    fontWeight = FontWeight.Bold
                )
            }

            // 同步速度
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
                    Text("上传: 1.2 MB/s")
                }
                Row(horizontalArrangement = Arrangement.spacedBy(4.dp)) {
                    Icon(
                        imageVector = Icons.Default.ArrowDownward,
                        contentDescription = null,
                        modifier = Modifier.size(16.dp)
                    )
                    Text("下载: 0.8 MB/s")
                }
            }

            // 进度条
            Column(verticalArrangement = Arrangement.spacedBy(4.dp)) {
                LinearProgressIndicator(
                    progress = { 0.65f },
                    modifier = Modifier.fillMaxWidth()
                )
                Text(
                    text = "总体进度: 65%",
                    fontSize = 12.sp,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
            }
        }
    }
}

/**
 * 存储空间卡片
 */
@Composable
fun StorageCard() {
    OutlinedCard(
        modifier = Modifier.fillMaxWidth()
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
                    text = "存储空间",
                    fontSize = 18.sp,
                    fontWeight = FontWeight.SemiBold
                )
                Icon(
                    imageVector = Icons.Default.Storage,
                    contentDescription = null
                )
            }

            // 进度条
            LinearProgressIndicator(
                progress = { 0.8f },
                modifier = Modifier.fillMaxWidth(),
                color = if (0.8f > 0.9f)
                    MaterialTheme.colorScheme.error
                else
                    MaterialTheme.colorScheme.primary
            )

            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween
            ) {
                Text(
                    text = "已使用 9.6 GB",
                    fontSize = 14.sp
                )
                Text(
                    text = "共 12 GB",
                    fontSize = 14.sp,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
            }
        }
    }
}

/**
 * 快速操作区域
 */
@Composable
fun QuickActionsSection() {
    Column(verticalArrangement = Arrangement.spacedBy(12.dp)) {
        Text(
            text = "快速操作",
            fontSize = 18.sp,
            fontWeight = FontWeight.SemiBold
        )

        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.spacedBy(12.dp)
        ) {
            QuickActionButton(
                icon = Icons.Default.Refresh,
                label = "扫描文件",
                modifier = Modifier.weight(1f),
                onClick = { /* 执行扫描 */ }
            )
            QuickActionButton(
                icon = Icons.Default.Sync,
                label = "立即同步",
                modifier = Modifier.weight(1f),
                onClick = { /* 执行同步 */ }
            )
            QuickActionButton(
                icon = Icons.Default.MonitorHeart,
                label = "监控中心",
                modifier = Modifier.weight(1f),
                onClick = { /* 打开监控 */ }
            )
        }
    }
}

/**
 * 快速操作按钮
 */
@Composable
fun QuickActionButton(
    icon: androidx.compose.ui.graphics.vector.ImageVector,
    label: String,
    modifier: Modifier = Modifier,
    onClick: () -> Unit
) {
    ElevatedCard(
        modifier = modifier.clickable(onClick = onClick)
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(16.dp),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.spacedBy(8.dp)
        ) {
            Icon(
                imageVector = icon,
                contentDescription = null,
                modifier = Modifier.size(32.dp),
                tint = MaterialTheme.colorScheme.primary
            )
            Text(
                text = label,
                fontSize = 12.sp,
                maxLines = 1
            )
        }
    }
}

/**
 * 区域标题
 */
@Composable
fun SectionHeader(
    title: String,
    onViewAll: () -> Unit
) {
    Row(
        modifier = Modifier.fillMaxWidth(),
        horizontalArrangement = Arrangement.SpaceBetween,
        verticalAlignment = Alignment.CenterVertically
    ) {
        Text(
            text = title,
            fontSize = 18.sp,
            fontWeight = FontWeight.SemiBold
        )
        TextButton(onClick = onViewAll) {
            Text("查看全部")
            Icon(
                imageVector = Icons.Default.ChevronRight,
                contentDescription = null,
                modifier = Modifier.size(20.dp)
            )
        }
    }
}

/**
 * 最近文件项
 */
@Composable
fun RecentFileItem(file: RecentFile) {
    OutlinedCard(
        modifier = Modifier
            .fillMaxWidth()
            .clickable { /* 打开文件 */ }
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(16.dp),
            horizontalArrangement = Arrangement.spacedBy(12.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            Icon(
                imageVector = file.icon,
                contentDescription = null,
                modifier = Modifier.size(40.dp),
                tint = MaterialTheme.colorScheme.primary
            )

            Column(
                modifier = Modifier.weight(1f),
                verticalArrangement = Arrangement.spacedBy(4.dp)
            ) {
                Text(
                    text = file.name,
                    fontWeight = FontWeight.Medium
                )
                Text(
                    text = file.time,
                    fontSize = 12.sp,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
            }

            Text(
                text = file.size,
                fontSize = 12.sp,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )
        }
    }
}

/**
 * 设备项
 */
@Composable
fun DeviceItem(device: ConnectedDevice) {
    OutlinedCard(
        modifier = Modifier
            .fillMaxWidth()
            .clickable { /* 查看设备详情 */ }
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(16.dp),
            horizontalArrangement = Arrangement.spacedBy(12.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            Icon(
                imageVector = device.icon,
                contentDescription = null,
                modifier = Modifier.size(40.dp),
                tint = MaterialTheme.colorScheme.primary
            )

            Column(
                modifier = Modifier.weight(1f),
                verticalArrangement = Arrangement.spacedBy(4.dp)
            ) {
                Text(
                    text = device.name,
                    fontWeight = FontWeight.Medium
                )
                Text(
                    text = device.ip,
                    fontSize = 12.sp,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
            }

            // 在线状态指示器
            Row(
                horizontalArrangement = Arrangement.spacedBy(4.dp),
                verticalAlignment = Alignment.CenterVertically
            ) {
                Box(
                    modifier = Modifier
                        .size(8.dp)
                        .background(
                            color = if (device.isOnline)
                                MaterialTheme.colorScheme.primary
                            else
                                MaterialTheme.colorScheme.error,
                            shape = androidx.compose.foundation.shape.CircleShape
                        )
                )
                Text(
                    text = if (device.isOnline) "在线" else "离线",
                    fontSize = 12.sp,
                    color = if (device.isOnline)
                        MaterialTheme.colorScheme.primary
                    else
                        MaterialTheme.colorScheme.error
                )
            }
        }
    }
}

// ============ 数据模型 ============

data class RecentFile(
    val name: String,
    val time: String,
    val size: String,
    val icon: androidx.compose.ui.graphics.vector.ImageVector
)

data class ConnectedDevice(
    val name: String,
    val ip: String,
    val isOnline: Boolean,
    val icon: androidx.compose.ui.graphics.vector.ImageVector
)

// ============ 模拟数据 ============

fun getRecentFiles() = listOf(
    RecentFile("工作报告.docx", "刚刚", "2.3 MB", Icons.Default.Description),
    RecentFile("度假照片.jpg", "5分钟前", "4.8 MB", Icons.Default.Image),
    RecentFile("项目代码", "1小时前", "156 MB", Icons.Default.Folder),
    RecentFile("会议录音.mp3", "今天 14:30", "12.5 MB", Icons.Default.AudioFile)
)

fun getConnectedDevices() = listOf(
    ConnectedDevice("Windows PC", "192.168.1.100", true, Icons.Default.Computer),
    ConnectedDevice("小米12 Pro", "192.168.1.101", true, Icons.Default.Smartphone),
    ConnectedDevice("MacBook Pro", "192.168.1.102", false, Icons.Default.Laptop)
)