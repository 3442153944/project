package com.example.rootinstaller.core

import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext

object RootShell {

    data class ShellResult(
        val exitCode: Int,
        val stdout: String,
        val stderr: String
    ) {
        val success get() = exitCode == 0
        val output get() = (stdout + "\n" + stderr).trim()
    }

    /** 普通 shell 权限，不需要 su */
    suspend fun execShell(cmd: String): ShellResult = withContext(Dispatchers.IO) {
        exec(arrayOf("sh", "-c", cmd))
    }

    /** root 权限，su -c */
    suspend fun execRoot(cmd: String): ShellResult = withContext(Dispatchers.IO) {
        exec(arrayOf("su", "-c", cmd))
    }

    /** 自动降级：先用 shell，输出为空或 Permission denied 则升级到 root */
    suspend fun execAuto(cmd: String): ShellResult = withContext(Dispatchers.IO) {
        val shellResult = execShell(cmd)
        if (shellResult.success && !shellResult.stdout.contains("Permission denied")) {
            shellResult
        } else {
            execRoot(cmd)
        }
    }

    suspend fun isAvailable(): Boolean = withContext(Dispatchers.IO) {
        try {
            execRoot("id").stdout.contains("uid=0")
        } catch (e: Exception) {
            false
        }
    }

    private fun exec(cmd: Array<String>): ShellResult {
        return try {
            val process = Runtime.getRuntime().exec(cmd)
            val stdout = process.inputStream.bufferedReader().readText().trim()
            val stderr = process.errorStream.bufferedReader().readText().trim()
            val exitCode = process.waitFor()
            ShellResult(exitCode, stdout, stderr)
        } catch (e: Exception) {
            ShellResult(-1, "", e.message ?: "unknown error")
        }
    }
}