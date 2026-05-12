package com.example.rootinstaller.model

import android.graphics.drawable.Drawable

data class ApkInfo(
    val path: String,
    val packageName: String,
    val appName: String,
    val versionName: String,
    val versionCode: Long,
    val icon: Drawable?,
    // 已安装版本信息（null 表示未安装）
    val installedVersionName: String? = null,
    val installedVersionCode: Long? = null
) {
    val installState: InstallState get() = when {
        installedVersionCode == null -> InstallState.NEW
        versionCode > installedVersionCode -> InstallState.UPGRADE
        versionCode == installedVersionCode -> InstallState.SAME
        else -> InstallState.DOWNGRADE
    }
}

enum class InstallState(val label: String) {
    NEW("新安装"),
    UPGRADE("升级"),
    SAME("重复安装"),
    DOWNGRADE("降级")
}