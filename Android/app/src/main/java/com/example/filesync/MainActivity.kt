package com.example.filesync

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.compose.foundation.layout.*
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.material3.adaptive.navigationsuite.NavigationSuiteScaffold
import androidx.compose.runtime.*
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.tooling.preview.PreviewScreenSizes
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.example.filesync.network.Request
import com.example.filesync.network.WebSocket
import com.example.filesync.ui.screen.HomeScreen
import com.example.filesync.ui.screen.person.PersonalScreen
import com.example.filesync.ui.theme.FileSyncTheme
import com.example.filesync.util.RootHelper
import com.example.filesync.util.rememberPermissionState
import kotlinx.coroutines.launch

/**
 * 主 Activity
 * 应用启动入口，负责初始化全局配置和权限检查
 */
class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)

        // 初始化网络请求模块（必须在使用 Request 前调用）
        Request.init(this)
        // 初始化 WebSocket 模块
        WebSocket.init()

        // 可选：设置自定义 baseUrl
        // Request.setBaseUrl("https://www.sunyuanling.cn")

        enableEdgeToEdge()
        setContent {
            FileSyncTheme {
                var permissionChecked by remember { mutableStateOf(false) }

                if (!permissionChecked) {
                    PermissionRequestScreen(
                        onPermissionsGranted = { permissionChecked = true }
                    )
                } else {
                    FileSyncApp()
                }
            }
        }
    }
}

/**
 * 权限申请界面
 *
 * 依次申请：
 * 1. 基础权限（存储、网络等）
 * 2. 文件管理权限（All Files Access）
 * 3. Root 权限（如果设备已 Root）
 *
 * @param onPermissionsGranted 所有权限授予后的回调
 */
@Composable
fun PermissionRequestScreen(onPermissionsGranted: () -> Unit) {
    val scope = rememberCoroutineScope()
    var rootAvailable by remember { mutableStateOf(false) }
    var rootGranted by remember { mutableStateOf(false) }
    var checkingRoot by remember { mutableStateOf(true) }

    val permissionState = rememberPermissionState(
        onAllPermissionsGranted = {
            if (rootAvailable && rootGranted || !rootAvailable) {
                onPermissionsGranted()
            }
        }
    )

    // 检查设备是否已 Root
    LaunchedEffect(Unit) {
        rootAvailable = RootHelper.isDeviceRooted()
        checkingRoot = false
    }

    Box(
        modifier = Modifier.fillMaxSize(),
        contentAlignment = Alignment.Center
    ) {
        Column(
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.spacedBy(16.dp),
            modifier = Modifier.padding(32.dp)
        ) {
            Text("权限申请", fontSize = 24.sp)

            Spacer(modifier = Modifier.height(16.dp))

            // 基础权限（存储、网络等）
            Button(
                onClick = { permissionState.requestPermissions() },
                enabled = !permissionState.permissionsGranted
            ) {
                Text(if (permissionState.permissionsGranted) "✓ 基础权限已授予" else "申请基础权限")
            }

            // 文件管理权限（All Files Access）
            Button(
                onClick = { permissionState.requestManageStorage() },
                enabled = !permissionState.manageStorageGranted
            ) {
                Text(if (permissionState.manageStorageGranted) "✓ 文件管理权限已授予" else "申请文件管理权限")
            }

            // Root 权限（仅在设备已 Root 时显示）
            if (!checkingRoot) {
                if (rootAvailable) {
                    Button(
                        onClick = {
                            scope.launch {
                                rootGranted = RootHelper.requestRootAccess()
                            }
                        },
                        enabled = !rootGranted
                    ) {
                        Text(if (rootGranted) "✓ Root权限已授予" else "申请Root权限")
                    }
                } else {
                    Text(
                        "设备未Root，将使用普通模式",
                        fontSize = 14.sp,
                        color = MaterialTheme.colorScheme.secondary
                    )
                }
            }

            Spacer(modifier = Modifier.height(16.dp))

            // 继续按钮（所有必需权限授予后才可点击）
            Button(
                onClick = onPermissionsGranted,
                enabled = permissionState.permissionsGranted &&
                        permissionState.manageStorageGranted &&
                        (rootGranted || !rootAvailable)
            ) {
                Text("继续")
            }
        }
    }
}

/**
 * 主应用界面
 * 包含底部导航栏和四个主要页面
 */
@PreviewScreenSizes
@Composable
fun FileSyncApp() {
    var currentDestination by rememberSaveable { mutableStateOf(AppDestinations.HOME) }

    NavigationSuiteScaffold(
        navigationSuiteItems = {
            AppDestinations.entries.forEach {
                item(
                    icon = { Icon(it.icon, contentDescription = it.label) },
                    label = { Text(it.label) },
                    selected = it == currentDestination,
                    onClick = { currentDestination = it }
                )
            }
        }
    ) {
        Scaffold(modifier = Modifier.fillMaxSize()) { innerPadding ->
            when (currentDestination) {
                AppDestinations.HOME -> HomeScreen(modifier = Modifier.padding(innerPadding))
                AppDestinations.FILES -> PlaceholderScreen("文件", Modifier.padding(innerPadding))
                AppDestinations.MONITOR -> PlaceholderScreen("监控", Modifier.padding(innerPadding))
                AppDestinations.PERSONAL -> PersonalScreen(modifier = Modifier.padding(innerPadding))
            }
        }
    }
}

/**
 * 应用目的地枚举
 * 定义应用的四个主要页面
 */
enum class AppDestinations(
    val label: String,
    val icon: ImageVector,
) {
    HOME("主页", Icons.Default.Home),
    FILES("文件", Icons.Default.Folder),
    MONITOR("监控", Icons.Default.Monitor),
    PERSONAL("个人中心", Icons.Default.Person)
}

/**
 * 占位界面
 * 用于未实现的页面
 */
@Composable
fun PlaceholderScreen(name: String, modifier: Modifier = Modifier) {
    Box(modifier = modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
        Text(text = "当前页面: $name", fontSize = 20.sp)
    }
}