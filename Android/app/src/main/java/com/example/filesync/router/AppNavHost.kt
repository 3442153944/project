// router/AppNavHost.kt
package com.example.filesync.router

import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.navigation.NavHostController
import androidx.navigation.NavType
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.navArgument
import com.example.filesync.ui.screen.*
import com.example.filesync.ui.screen.files.*
import com.example.filesync.ui.screen.permission.PermissionScreen
import com.example.filesync.ui.screen.person.PersonalScreen
//import com.example.filesync.ui.screen.settings.*
import com.example.filesync.ui.screen.transmission.TransferScreen

/**
 * 应用导航主机
 * 统一管理所有页面路由
 */
@Composable
fun AppNavHost(
    navController: NavHostController,
    modifier: Modifier = Modifier,
    startDestination: String = AppRoute.Home.route
) {
    NavHost(
        navController = navController,
        startDestination = startDestination,
        modifier = modifier
    ) {
        // ==================== 主 Tab 页面 ====================
        composable(AppRoute.Home.route) {
            HomeScreen(navController = navController)
        }

        composable(AppRoute.Files.route) {
            FileScreen()
        }

        composable(AppRoute.Monitor.route) {
            MonitorScreen(navController = navController)
        }

        composable(AppRoute.Personal.route) {
            PersonalScreen()
        }

        // ==================== 详情页面 ====================

        // 传输页面
        composable(AppRoute.Transfer.route) {
            TransferScreen(
                onBackClick = { navController.navigateUp() }
            )
        }

        // 文件详情
        composable(
            route = AppRoute.FileDetail.route,
            arguments = listOf(
                navArgument(AppRoute.FileDetail.ARG_FILE_ID) {
                    type = NavType.StringType
                }
            )
        ) { backStackEntry ->
            val fileId = backStackEntry.arguments?.getString(AppRoute.FileDetail.ARG_FILE_ID) ?: ""
            FileDetailScreen(
                fileId = fileId,
                onBackClick = { navController.navigateUp() }
            )
        }

        // 文件上传
        composable(AppRoute.FileUpload.route) {
            FileUploadScreen(
                onBackClick = { navController.navigateUp() },
                onUploadComplete = {
                    navController.navigateUp()
                    // 可以刷新文件列表
                }
            )
        }

        // 文件搜索
        composable(AppRoute.FileSearch.route) {
            FileSearchScreen(
                onBackClick = { navController.navigateUp() },
                onFileClick = { fileId ->
                    navController.navigateToDetail(
                        AppRoute.FileDetail.createRoute(fileId)
                    )
                }
            )
        }

        // ==================== 设置相关 ====================

        // 设置主页
        composable(AppRoute.Settings.route) {
            SettingsScreen(
                onBackClick = { navController.navigateUp() },
                onServerSettingsClick = {
                    navController.navigateToDetail(AppRoute.ServerSettings)
                },
                onSyncSettingsClick = {
                    navController.navigateToDetail(AppRoute.SyncSettings)
                },
                onAboutClick = {
                    navController.navigateToDetail(AppRoute.About)
                }
            )
        }

        // 服务器设置
        composable(AppRoute.ServerSettings.route) {
            ServerSettingsScreen(
                onBackClick = { navController.navigateUp() }
            )
        }

        // 同步设置
        composable(AppRoute.SyncSettings.route) {
            SyncSettingsScreen(
                onBackClick = { navController.navigateUp() }
            )
        }

        // 关于页面
        composable(AppRoute.About.route) {
            AboutScreen(
                onBackClick = { navController.navigateUp() }
            )
        }

        // ==================== 特殊页面 ====================

        // 权限申请
        composable(AppRoute.Permission.route) {
            PermissionScreen(
                onPermissionsGranted = {
                    navController.navigateAndClearBackStack(AppRoute.Home)
                }
            )
        }

        // 登录页面
        composable(AppRoute.Login.route) {
            LoginScreen(
                onLoginSuccess = {
                    navController.navigateAndClearBackStack(AppRoute.Home)
                },
                onSkipLogin = {
                    navController.navigateAndClearBackStack(AppRoute.Home)
                }
            )
        }
    }
}