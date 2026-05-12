package com.example.rootinstaller.core

import android.util.Log
import com.example.rootinstaller.model.FileItem
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext

object FileManager {

    private const val TAG = "RootInstaller"

    // 匹配 ls -la 每一行：
    // 第1段：权限（必须以 d/l/- 等开头）
    // 中间：任意字段（links owner group size 等，数量不定）
    // 锚点：YYYY-MM-DD HH:MM（日期时间是唯一格式固定的字段）
    // 最后：文件名（可含空格、中文）
    private val LINE_REGEX = Regex(
        """^([dlcbps-]\S+)\s+.*?\s+(\d+)\s+\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2}\s+(.+)$"""
    )

    suspend fun listDir(path: String): Result<List<FileItem>> = withContext(Dispatchers.IO) {
        runCatching {
            val result = RootShell.execAuto("""ls -la "$path" 2>&1""")
            Log.d(TAG, "ls [$path] exit=${result.exitCode}")

            val items = mutableListOf<FileItem>()

            for (line in result.stdout.lines()) {
                val trimmed = line.trim()
                if (trimmed.isEmpty() || trimmed.startsWith("total")) continue

                val match = LINE_REGEX.find(trimmed) ?: continue

                val perms   = match.groupValues[1]
                val size    = match.groupValues[2].toLongOrNull() ?: 0L
                val rawName = match.groupValues[3]

                val isDir  = perms.startsWith("d")
                val isLink = perms.startsWith("l")

                // 软链接 "name -> target"，只取箭头前
                val name = if (isLink) rawName.substringBefore(" -> ").trim()
                else rawName.trim()

                if (name == "." || name == "..") continue

                val fullPath = if (path.endsWith("/")) "$path$name" else "$path/$name"

                items.add(FileItem(
                    name        = name,
                    path        = fullPath,
                    isDirectory = isDir,
                    size        = size,
                    isApk       = !isDir && name.endsWith(".apk", ignoreCase = true)
                ))
            }

            Log.d(TAG, "parsed ${items.size} items")
            items.sortedWith(
                compareByDescending<FileItem> { it.isDirectory }.thenBy { it.name.lowercase() }
            )
        }
    }

    suspend fun copyToTemp(sourcePath: String): Result<String> = withContext(Dispatchers.IO) {
        runCatching {
            val destPath = "/data/local/tmp/ri_${System.currentTimeMillis()}.apk"
            val r = RootShell.execRoot("""cp -f "$sourcePath" "$destPath" && chmod 644 "$destPath"""")
            if (!r.success) error("复制失败: ${r.output}")
            destPath
        }
    }

    suspend fun deleteFile(path: String) {
        RootShell.execRoot("""rm -f "$path"""")
    }
}