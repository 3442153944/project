package com.example.rootinstaller.core

import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext

object InstallManager {

    data class InstallResult(
        val success: Boolean,
        val message: String,
        val errorCode: InstallError? = null
    )

    enum class InstallError {
        SIGNATURE_MISMATCH,
        VERSION_DOWNGRADE,
        PARSE_FAILED,
        NO_STORAGE,
        COPY_FAILED,
        UNKNOWN
    }

    data class InstallOptions(
        val allowReplace: Boolean = true,     // -r 覆盖安装
        val allowDowngrade: Boolean = true,   // -d 降级
        val grantPermissions: Boolean = true, // -g 自动授权
    )

    suspend fun install(
        apkPath: String,
        options: InstallOptions = InstallOptions()
    ): InstallResult = withContext(Dispatchers.IO) {

        // 1. 复制到 /data/local/tmp（处理 SELinux 隔离）
        val tempPath = FileManager.copyToTemp(apkPath).getOrElse {
            return@withContext InstallResult(false, "APK 复制失败: ${it.message}", InstallError.COPY_FAILED)
        }

        try {
            // 2. 构建 pm install 命令
            val cmd = buildString {
                append("pm install")
                if (options.allowReplace) append(" -r")
                if (options.allowDowngrade) append(" -d")
                if (options.grantPermissions) append(" -g")
                append(" -t")                              // 允许 test APK
                append(" --bypass-low-target-sdk-block")  // 绕过低 targetSdk 限制
                append(""" "$tempPath" """)
            }

            // 3. 执行
            val result = RootShell.execRoot(cmd)
            parseResult(result.output, result.exitCode)
        } finally {
            // 4. 清理临时文件
            FileManager.deleteFile(tempPath)
        }
    }

    private fun parseResult(output: String, exitCode: Int): InstallResult {
        if (exitCode == 0 && output.contains("Success", ignoreCase = true)) {
            return InstallResult(true, "安装成功")
        }
        return when {
            output.contains("INSTALL_FAILED_UPDATE_INCOMPATIBLE") ->
                InstallResult(false, "签名不匹配，请先卸载原版本再安装", InstallError.SIGNATURE_MISMATCH)

            output.contains("INSTALL_FAILED_VERSION_DOWNGRADE") ->
                InstallResult(false, "版本降级被拒绝", InstallError.VERSION_DOWNGRADE)

            output.contains("INSTALL_PARSE_FAILED") ->
                InstallResult(false, "APK 解析失败，文件可能已损坏", InstallError.PARSE_FAILED)

            output.contains("INSTALL_FAILED_INSUFFICIENT_STORAGE") ->
                InstallResult(false, "存储空间不足", InstallError.NO_STORAGE)

            output.contains("INSTALL_FAILED_ALREADY_EXISTS") ->
                InstallResult(false, "应用已存在（相同版本）", InstallError.UNKNOWN)

            output.isEmpty() && exitCode != 0 ->
                InstallResult(false, "安装失败（exit=$exitCode），请检查 root 权限", InstallError.UNKNOWN)

            else ->
                InstallResult(false, output.take(200).ifEmpty { "未知错误" }, InstallError.UNKNOWN)
        }
    }
}