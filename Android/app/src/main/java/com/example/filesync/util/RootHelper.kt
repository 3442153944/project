package com.example.filesync.util

import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import java.io.DataOutputStream

object RootHelper {

    // 检查设备是否已Root
    suspend fun isDeviceRooted(): Boolean = withContext(Dispatchers.IO) {
        val paths = arrayOf(
            "/system/app/Superuser.apk",
            "/sbin/su",
            "/system/bin/su",
            "/system/xbin/su",
            "/data/local/xbin/su",
            "/data/local/bin/su",
            "/system/sd/xbin/su",
            "/system/bin/failsafe/su",
            "/data/local/su",
            "/su/bin/su"
        )

        paths.any { java.io.File(it).exists() }
    }

    /**
     * 检查是否已拥有Root权限
     * 与 requestRootAccess() 的区别：
     * - checkRootAccess(): 静默检查，不会触发授权弹窗
     * - requestRootAccess(): 主动请求授权，会触发弹窗
     */
    suspend fun checkRootAccess(): Boolean = withContext(Dispatchers.IO) {
        try {
            // 使用 -c 参数执行简单命令，避免交互式shell
            val process = Runtime.getRuntime().exec(arrayOf("su", "-c", "id"))
            val output = process.inputStream.bufferedReader().readText()

            // 等待进程结束，设置超时时间
            val exitCode = withContext(Dispatchers.IO) {
                process.waitFor()
                process.exitValue()
            }

            // 检查命令是否成功执行
            exitCode == 0 && output.contains("uid=0")
        } catch (e: Exception) {
            false
        }
    }

    /**
     * 请求Root权限
     * 会触发授权弹窗（如果之前未授权）
     */
    suspend fun requestRootAccess(): Boolean = withContext(Dispatchers.IO) {
        try {
            val process = Runtime.getRuntime().exec("su")
            val os = DataOutputStream(process.outputStream)
            os.writeBytes("id\n")
            os.writeBytes("exit\n")
            os.flush()
            os.close()

            val output = process.inputStream.bufferedReader().readText()
            process.waitFor()

            process.exitValue() == 0 && output.contains("uid=0")
        } catch (e: Exception) {
            false
        }
    }

    /**
     * 执行Root命令
     */
    suspend fun executeRootCommand(command: String): Result<String> = withContext(Dispatchers.IO) {
        try {
            val process = Runtime.getRuntime().exec("su")
            val os = DataOutputStream(process.outputStream)

            os.writeBytes("$command\n")
            os.writeBytes("exit\n")
            os.flush()
            os.close()

            val output = process.inputStream.bufferedReader().readText()
            val error = process.errorStream.bufferedReader().readText()

            process.waitFor()

            if (process.exitValue() == 0) {
                Result.success(output)
            } else {
                Result.failure(Exception(error.ifEmpty { "Command failed with exit code ${process.exitValue()}" }))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
}