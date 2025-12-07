// ui/screen/files/FileListScreen.kt
package com.example.filesync.ui.screen.files

import androidx.activity.compose.BackHandler
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.lifecycle.viewmodel.compose.viewModel
import com.example.filesync.ui.components.files.*
import com.example.filesync.ui.viewModel.files.FileListViewModel

@Composable
fun FileListScreen(
    initialPath: String,
    onBack: () -> Unit,
    modifier: Modifier = Modifier
) {
    val viewModel: FileListViewModel = viewModel()

    val fileData by viewModel.fileData.collectAsState()
    val loading by viewModel.loading.collectAsState()
    val error by viewModel.error.collectAsState()
    val pathStack by viewModel.pathStack.collectAsState()

    // 初始加载
    LaunchedEffect(initialPath) {
        viewModel.loadDirectory(initialPath)
    }

    // 系统返回键处理
    BackHandler(enabled = pathStack.isNotEmpty()) {
        viewModel.navigateBack()
    }

    LazyColumn(
        modifier = modifier
            .fillMaxSize()
            .padding(16.dp),
        verticalArrangement = Arrangement.spacedBy(8.dp)
    ) {
        // 头部
        item {
            FileListHeader(
                currentPath = fileData?.currentPath ?: initialPath,
                canGoBack = pathStack.isNotEmpty() || fileData?.parentPath?.isNotEmpty() == true,
                loading = loading,
                onBack = {
                    if (!viewModel.navigateBack()) {
                        // 栈空了，尝试返回上级目录或退出
                        if (fileData?.parentPath?.isNotEmpty() == true) {
                            viewModel.navigateToParent()
                        } else {
                            onBack()
                        }
                    }
                },
                onRefresh = { viewModel.refresh() }
            )
        }

        // 统计
        fileData?.let { data ->
            item {
                FileListStats(
                    totalCount = data.totalCount,
                    dirCount = data.dirCount,
                    fileCount = data.fileCount
                )
            }
        }

        // 加载中
        if (loading) {
            item { LoadingIndicator() }
        }

        // 错误
        if (error != null) {
            item { ErrorCard(message = error!!) }
        }

        // 文件列表：先文件夹后文件
        val items = fileData?.items ?: emptyList()
        val sortedItems = items.sortedWith(compareBy({ !it.isDir }, { it.name.lowercase() }))

        items(sortedItems, key = { it.path }) { item ->
            FileItemCard(
                item = item,
                onClick = {
                    if (item.isDir) {
                        viewModel.navigateTo(item.path)
                    } else {
                        // TODO: 打开文件
                    }
                }
            )
        }
    }
}