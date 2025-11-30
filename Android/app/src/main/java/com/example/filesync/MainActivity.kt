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
import androidx.compose.ui.platform.LocalLifecycleOwner
import androidx.compose.ui.tooling.preview.PreviewScreenSizes
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.LifecycleEventObserver
import androidx.lifecycle.compose.LocalLifecycleOwner
import com.example.filesync.data.sync.WebSocketManager
import com.example.filesync.network.Request
import com.example.filesync.ui.screen.HomeScreen
import com.example.filesync.ui.screen.person.PersonalScreen
import com.example.filesync.ui.theme.FileSyncTheme
import com.example.filesync.util.PermissionHelper
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

        // 可选：设置自定义服务器地址
        // Request.baseUrl = "http://192.168.31.100:9999/api"
        // WebSocketManager.setServerUrl("ws://192.168.31.100:9999/api/ws/connect")

        enableEdgeToEdge()
        setContent {
            FileSyncTheme {
                PermissionCheckWrapper()
            }
        }
    }
}

/**
 * 权限检查包装器
 * 自动检查权限状态，只在需要时显示申请界面
 */
@Composable
fun PermissionCheckWrapper() {
    val context = androidx.compose.ui.platform.LocalContext.current
    val scope = rememberCoroutineScope()

    var isChecking by remember { mutableStateOf(true) }
    var needsPermissionRequest by remember { mutableStateOf(false) }
    var initialPermissionCheck by remember { mutableStateOf<PermissionCheckResult?>(null) }

    // 启动时自动检查所有权限
    LaunchedEffect(Unit) {
        scope.launch {
            // 检查基础权限
            val hasBasicPermissions = PermissionHelper.hasAllPermissions(context)
            // 检查文件管理权限
            val hasManageStorage = PermissionHelper.hasManageExternalStoragePermission()
            // 检查设备是否已Root
            val isRooted = RootHelper.isDeviceRooted()
            // 如果设备已Root，检查是否已授予Root权限
            val hasRootAccess = if (isRooted) {
                RootHelper.checkRootAccess()
            } else {
                true // 设备未Root，视为不需要Root权限
            }

            initialPermissionCheck = PermissionCheckResult(
                hasBasicPermissions = hasBasicPermissions,
                hasManageStorage = hasManageStorage,
                isRooted = isRooted,
                hasRootAccess = hasRootAccess
            )

            // 如果所有权限都已授予，直接进入应用
            needsPermissionRequest = !(hasBasicPermissions && hasManageStorage && hasRootAccess)
            isChecking = false
        }
    }

    when {
        isChecking -> {
            // 显示加载界面
            Box(
                modifier = Modifier.fillMaxSize(),
                contentAlignment = Alignment.Center
            ) {
                Column(
                    horizontalAlignment = Alignment.CenterHorizontally,
                    verticalArrangement = Arrangement.spacedBy(16.dp)
                ) {
                    CircularProgressIndicator()
                    Text("正在检查权限...", fontSize = 16.sp)
                }
            }
        }
        needsPermissionRequest -> {
            // 显示权限申请界面
            PermissionRequestScreen(
                initialCheck = initialPermissionCheck!!,
                onPermissionsGranted = { needsPermissionRequest = false }
            )
        }
        else -> {
            // 所有权限已授予，进入主应用
            FileSyncApp()
        }
    }
}

/**
 * 权限检查结果
 */
data class PermissionCheckResult(
    val hasBasicPermissions: Boolean,
    val hasManageStorage: Boolean,
    val isRooted: Boolean,
    val hasRootAccess: Boolean
)

/**
 * 权限申请界面
 * 只显示未授予的权限，已授予的权限自动跳过
 *
 * @param initialCheck 初始权限检查结果
 * @param onPermissionsGranted 所有权限授予后的回调
 */
@Composable
fun PermissionRequestScreen(
    initialCheck: PermissionCheckResult,
    onPermissionsGranted: () -> Unit
) {
    val scope = rememberCoroutineScope()
    var rootGranted by remember { mutableStateOf(initialCheck.hasRootAccess) }

    val permissionState = rememberPermissionState(
        initialBasicPermissions = initialCheck.hasBasicPermissions,
        initialManageStorage = initialCheck.hasManageStorage,
        onAllPermissionsGranted = {
            // 检查是否还需要Root权限
            if (!initialCheck.isRooted || rootGranted) {
                onPermissionsGranted()
            }
        }
    )

    // 自动检查是否所有权限都已授予
    LaunchedEffect(permissionState.permissionsGranted, permissionState.manageStorageGranted, rootGranted) {
        if (permissionState.permissionsGranted &&
            permissionState.manageStorageGranted &&
            (!initialCheck.isRooted || rootGranted)) {
            onPermissionsGranted()
        }
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
            Icon(
                imageVector = Icons.Default.Security,
                contentDescription = null,
                modifier = Modifier.size(64.dp),
                tint = MaterialTheme.colorScheme.primary
            )

            Text(
                "权限申请",
                fontSize = 24.sp,
                style = MaterialTheme.typography.headlineMedium
            )

            Text(
                "应用需要以下权限才能正常运行",
                fontSize = 14.sp,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )

            Spacer(modifier = Modifier.height(8.dp))

            // 基础权限（只在未授予时显示）
            if (!permissionState.permissionsGranted) {
                PermissionCard(
                    title = "基础权限",
                    description = "存储、通讯录、短信、摄像头、麦克风",
                    isGranted = false,
                    onRequest = { permissionState.requestPermissions() }
                )
            }

            // 文件管理权限（只在未授予时显示）
            if (!permissionState.manageStorageGranted) {
                PermissionCard(
                    title = "文件管理权限",
                    description = "访问所有文件和文件夹",
                    isGranted = false,
                    onRequest = { permissionState.requestManageStorage() }
                )
            }

            // Root权限（只在设备已Root且未授予时显示）
            if (initialCheck.isRooted && !rootGranted) {
                PermissionCard(
                    title = "Root权限",
                    description = "访问系统级文件和功能",
                    isGranted = false,
                    onRequest = {
                        scope.launch {
                            rootGranted = RootHelper.requestRootAccess()
                        }
                    }
                )
            }

            // 如果所有必需权限都已授予，显示提示
            if (permissionState.permissionsGranted &&
                permissionState.manageStorageGranted &&
                (!initialCheck.isRooted || rootGranted)) {
                Spacer(modifier = Modifier.height(8.dp))
                Text(
                    "✓ 所有权限已授予，正在进入应用...",
                    fontSize = 16.sp,
                    color = MaterialTheme.colorScheme.primary
                )
            }
        }
    }
}

/**
 * 权限卡片
 */
@Composable
fun PermissionCard(
    title: String,
    description: String,
    isGranted: Boolean,
    onRequest: () -> Unit
) {
    ElevatedCard(
        modifier = Modifier.fillMaxWidth()
    ) {
        Column(
            modifier = Modifier.padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(8.dp)
        ) {
            Row(
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.spacedBy(8.dp)
            ) {
                Icon(
                    imageVector = if (isGranted) Icons.Default.CheckCircle else Icons.Default.Info,
                    contentDescription = null,
                    tint = if (isGranted)
                        MaterialTheme.colorScheme.primary
                    else
                        MaterialTheme.colorScheme.onSurfaceVariant
                )
                Text(
                    text = title,
                    style = MaterialTheme.typography.titleMedium
                )
            }

            Text(
                text = description,
                style = MaterialTheme.typography.bodyMedium,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )

            if (!isGranted) {
                Button(
                    onClick = onRequest,
                    modifier = Modifier.fillMaxWidth()
                ) {
                    Text("授予权限")
                }
            }
        }
    }
}

/**
 * 主应用界面
 * 包含底部导航栏和四个主要页面
 * 管理 WebSocket 连接的生命周期
 */
@PreviewScreenSizes
@Composable
fun FileSyncApp() {
    var currentDestination by rememberSaveable { mutableStateOf(AppDestinations.HOME) }
    val lifecycleOwner = androidx.lifecycle.compose.LocalLifecycleOwner.current
    val scope = rememberCoroutineScope()

    // 监听应用前后台状态，管理 WebSocket 连接
    DisposableEffect(lifecycleOwner) {
        val observer = LifecycleEventObserver { _, event ->
            when (event) {
                Lifecycle.Event.ON_START -> {
                    // 应用进入前台，连接 WebSocket
                    scope.launch {
                        if (Request.hasToken()) {
                            WebSocketManager.connect()
                        }
                    }
                }
                Lifecycle.Event.ON_STOP -> {
                    // 应用进入后台，断开 WebSocket
                    WebSocketManager.disconnect()
                }
                else -> {}
            }
        }

        lifecycleOwner.lifecycle.addObserver(observer)

        onDispose {
            lifecycleOwner.lifecycle.removeObserver(observer)
        }
    }

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