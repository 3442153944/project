// router/NavConfig.kt 新建文件
package com.example.filesync.router

import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*

object NavConfig {
    val bottomNavItems: List<BottomNavItem>
        get() = listOf(
            BottomNavItem(AppRoute.Home, "主页", Icons.Default.Home),
            BottomNavItem(AppRoute.Files, "文件", Icons.Default.Folder),
            BottomNavItem(AppRoute.Monitor, "监控", Icons.Default.Monitor),
            BottomNavItem(AppRoute.Personal, "个人中心", Icons.Default.Person)
        )

    // 纯字符串字面量，完全不引用 AppRoute 对象
    private val hideBottomNavRoutes = setOf("permission", "login")

    fun isMainTab(route: String?): Boolean {
        return bottomNavItems.any { it.route.route == route }
    }

    fun shouldShowBottomNav(route: String?): Boolean {
        return route !in hideBottomNavRoutes
    }
}