// ui/screen/PlaceholderScreen.kt
package com.example.filesync.ui.screen

import androidx.compose.foundation.layout.*
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.ArrowBack
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.sp
import androidx.navigation.NavController

@Composable
fun MonitorScreen(
    navController: NavController,
    modifier: Modifier = Modifier
) {
    PlaceholderScreen("监控", modifier)
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun FileDetailScreen(
    fileId: String,
    onBackClick: () -> Unit,
    modifier: Modifier = Modifier
) {
    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("文件详情") },
                navigationIcon = {
                    IconButton(onClick = onBackClick) {
                        Icon(Icons.Default.ArrowBack, contentDescription = "返回")
                    }
                }
            )
        }
    ) { paddingValues ->
        Box(
            modifier = modifier
                .fillMaxSize()
                .padding(paddingValues),
            contentAlignment = Alignment.Center
        ) {
            Text(text = "文件ID: $fileId", fontSize = 20.sp)
        }
    }
}

@Composable
fun FileUploadScreen(
    onBackClick: () -> Unit,
    onUploadComplete: () -> Unit,
    modifier: Modifier = Modifier
) {
    PlaceholderScreenWithBack("文件上传", onBackClick, modifier)
}

@Composable
fun FileSearchScreen(
    onBackClick: () -> Unit,
    onFileClick: (String) -> Unit,
    modifier: Modifier = Modifier
) {
    PlaceholderScreenWithBack("文件搜索", onBackClick, modifier)
}

@Composable
fun SettingsScreen(
    onBackClick: () -> Unit,
    onServerSettingsClick: () -> Unit,
    onSyncSettingsClick: () -> Unit,
    onAboutClick: () -> Unit,
    modifier: Modifier = Modifier
) {
    PlaceholderScreenWithBack("设置", onBackClick, modifier)
}

@Composable
fun ServerSettingsScreen(
    onBackClick: () -> Unit,
    modifier: Modifier = Modifier
) {
    PlaceholderScreenWithBack("服务器设置", onBackClick, modifier)
}

@Composable
fun SyncSettingsScreen(
    onBackClick: () -> Unit,
    modifier: Modifier = Modifier
) {
    PlaceholderScreenWithBack("同步设置", onBackClick, modifier)
}

@Composable
fun AboutScreen(
    onBackClick: () -> Unit,
    modifier: Modifier = Modifier
) {
    PlaceholderScreenWithBack("关于", onBackClick, modifier)
}

@Composable
fun LoginScreen(
    onLoginSuccess: () -> Unit,
    onSkipLogin: () -> Unit,
    modifier: Modifier = Modifier
) {
    PlaceholderScreen("登录", modifier)
}

@Composable
fun PlaceholderScreen(
    name: String,
    modifier: Modifier = Modifier
) {
    Box(
        modifier = modifier.fillMaxSize(),
        contentAlignment = Alignment.Center
    ) {
        Text(text = "当前页面: $name", fontSize = 20.sp)
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun PlaceholderScreenWithBack(
    name: String,
    onBackClick: () -> Unit,
    modifier: Modifier = Modifier
) {
    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text(name) },
                navigationIcon = {
                    IconButton(onClick = onBackClick) {
                        Icon(Icons.Default.ArrowBack, contentDescription = "返回")
                    }
                }
            )
        }
    ) { paddingValues ->
        Box(
            modifier = modifier
                .fillMaxSize()
                .padding(paddingValues),
            contentAlignment = Alignment.Center
        ) {
            Text(text = "当前页面: $name", fontSize = 20.sp)
        }
    }
}