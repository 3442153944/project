package com.example.rootinstaller.ui

import android.content.Context
import android.content.pm.PackageManager
import android.util.Log
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.example.rootinstaller.core.FileManager
import com.example.rootinstaller.core.InstallManager
import com.example.rootinstaller.core.RootShell
import com.example.rootinstaller.model.ApkInfo
import com.example.rootinstaller.model.FileItem
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

data class UiState(
    val rootAvailable: Boolean? = null,   // null = 检测中
    val currentPath: String = "/",
    val files: List<FileItem> = emptyList(),
    val isLoading: Boolean = false,
    val error: String? = null,
    val selectedApk: ApkInfo? = null,
    val installResult: InstallResult? = null,
    val pathHistory: List<String> = listOf("/"),
    val allowDowngrade: Boolean = true,
    val allowReplace: Boolean = true,
)

data class InstallResult(
    val success: Boolean,
    val message: String
)

class MainViewModel : ViewModel() {

    private val _state = MutableStateFlow(UiState())
    val state = _state.asStateFlow()

    init {
        viewModelScope.launch {
            Log.d("RootInstaller", "checking root...")
            val ok = RootShell.isAvailable()
            Log.d("RootInstaller", "root available: $ok")
            _state.update { it.copy(rootAvailable = ok) }
            if (ok) navigateTo("/")
        }
    }

    fun navigateTo(path: String) {
        viewModelScope.launch {
            _state.update { it.copy(isLoading = true, error = null) }
            Log.d("RootInstaller", "listing: $path")
            FileManager.listDir(path).fold(
                onSuccess = { files ->
                    Log.d("RootInstaller", "got ${files.size} items in $path")
                    _state.update { s ->
                        val newHistory = if (path != s.currentPath) s.pathHistory + path else s.pathHistory
                        s.copy(
                            currentPath = path,
                            files = files,
                            isLoading = false,
                            pathHistory = newHistory
                        )
                    }
                },
                onFailure = { e ->
                    Log.e("RootInstaller", "listDir failed: ${e.message}")
                    _state.update { it.copy(isLoading = false, error = "无法读取目录: ${e.message}") }
                }
            )
        }
    }

    fun navigateUp() {
        val history = _state.value.pathHistory
        if (history.size <= 1) return
        val newHistory = history.dropLast(1)
        val prevPath = newHistory.last()
        viewModelScope.launch {
            _state.update { it.copy(isLoading = true, error = null) }
            FileManager.listDir(prevPath).fold(
                onSuccess = { files ->
                    _state.update { s ->
                        s.copy(
                            currentPath = prevPath,
                            files = files,
                            isLoading = false,
                            pathHistory = newHistory
                        )
                    }
                },
                onFailure = { e ->
                    _state.update { it.copy(isLoading = false, error = "无法读取目录: ${e.message}") }
                }
            )
        }
    }

    fun selectApk(context: Context, item: FileItem) {
        if (!item.isApk) return
        viewModelScope.launch {
            val apkInfo = parseApkInfo(context, item.path) ?: return@launch
            _state.update { it.copy(selectedApk = apkInfo) }
        }
    }

    fun dismissApkSheet() {
        _state.update { it.copy(selectedApk = null) }
    }

    fun install() {
        val apk = _state.value.selectedApk ?: return
        val s = _state.value
        viewModelScope.launch {
            _state.update { it.copy(isLoading = true) }
            val result = InstallManager.install(
                apkPath = apk.path,
                options = InstallManager.InstallOptions(
                    allowReplace = s.allowReplace,
                    allowDowngrade = s.allowDowngrade,
                )
            )
            _state.update {
                it.copy(
                    isLoading = false,
                    selectedApk = null,
                    installResult = InstallResult(result.success, result.message)
                )
            }
        }
    }

    fun clearInstallResult() {
        _state.update { it.copy(installResult = null) }
    }

    fun toggleDowngrade(v: Boolean) = _state.update { it.copy(allowDowngrade = v) }
    fun toggleReplace(v: Boolean) = _state.update { it.copy(allowReplace = v) }
    fun refresh() = navigateTo(_state.value.currentPath)

    private fun parseApkInfo(context: Context, apkPath: String): ApkInfo? {
        return try {
            val pm = context.packageManager
            @Suppress("DEPRECATION")
            val pkgInfo = pm.getPackageArchiveInfo(apkPath, PackageManager.GET_ACTIVITIES)
                ?: return null
            pkgInfo.applicationInfo?.sourceDir = apkPath
            pkgInfo.applicationInfo?.publicSourceDir = apkPath

            val appName = pkgInfo.applicationInfo?.loadLabel(pm)?.toString() ?: pkgInfo.packageName
            val icon = pkgInfo.applicationInfo?.loadIcon(pm)

            val installedInfo = try {
                pm.getPackageInfo(pkgInfo.packageName, 0)
            } catch (e: PackageManager.NameNotFoundException) {
                null
            }

            @Suppress("DEPRECATION")
            ApkInfo(
                path = apkPath,
                packageName = pkgInfo.packageName,
                appName = appName,
                versionName = pkgInfo.versionName ?: "未知",
                versionCode = pkgInfo.longVersionCode,
                icon = icon,
                installedVersionName = installedInfo?.versionName,
                installedVersionCode = installedInfo?.longVersionCode
            )
        } catch (e: Exception) {
            Log.e("RootInstaller", "parseApkInfo failed: ${e.message}")
            null
        }
    }
}