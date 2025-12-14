// router/NavControllerExt.kt
package com.example.filesync.router

import androidx.navigation.NavController
import androidx.navigation.NavGraph.Companion.findStartDestination
import androidx.navigation.NavOptionsBuilder

/**
 * NavController 扩展函数
 * 提供统一的导航方法，简化使用
 */

/**
 * 导航到主 Tab 页面
 * 会清除所有详情页面的返回栈
 *
 * 使用场景：点击底部导航栏
 */
fun NavController.navigateToMainTab(route: String) {
    navigate(route) {
        // 弹出到起始目的地，保存状态
        popUpTo(graph.findStartDestination().id) {
            saveState = true
        }
        // 避免同一目的地的多个副本
        launchSingleTop = true
        // 恢复之前保存的状态
        restoreState = true
    }
}

/**
 * 导航到详情页面
 * 保留在当前导航栈之上
 *
 * 使用场景：打开文件详情、传输列表等
 */
fun NavController.navigateToDetail(route: String) {
    navigate(route) {
        launchSingleTop = true
    }
}

/**
 * 导航到详情页面（带构建器）
 *
 * 使用场景：需要自定义导航行为
 */
fun NavController.navigateToDetail(
    route: String,
    builder: NavOptionsBuilder.() -> Unit
) {
    navigate(route) {
        launchSingleTop = true
        builder()
    }
}

/**
 * 替换当前页面
 *
 * 使用场景：登录成功后跳转到主页
 */
fun NavController.navigateAndReplace(route: String) {
    navigate(route) {
        popUpTo(currentBackStackEntry?.destination?.route ?: return@navigate) {
            inclusive = true
        }
    }
}

/**
 * 清空返回栈并导航
 *
 * 使用场景：退出登录、权限授予后进入应用
 */
fun NavController.navigateAndClearBackStack(route: String) {
    navigate(route) {
        popUpTo(0) { inclusive = true }
    }
}

/**
 * 安全返回（如果无法返回则导航到主页）
 */
fun NavController.safeNavigateUp(): Boolean {
    return if (previousBackStackEntry != null) {
        navigateUp()
    } else {
        navigateToMainTab(AppRoute.Home.route)
        true
    }
}

/**
 * 导航到 AppRoute（类型安全）
 */
fun NavController.navigateToMainTab(route: AppRoute) {
    navigateToMainTab(route.route)
}

fun NavController.navigateToDetail(route: AppRoute) {
    navigateToDetail(route.route)
}

fun NavController.navigateAndReplace(route: AppRoute) {
    navigateAndReplace(route.route)
}

fun NavController.navigateAndClearBackStack(route: AppRoute) {
    navigateAndClearBackStack(route.route)
}