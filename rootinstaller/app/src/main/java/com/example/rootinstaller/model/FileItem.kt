package com.example.rootinstaller.model

data class FileItem(
    val name: String,
    val path: String,
    val isDirectory: Boolean,
    val size: Long = 0L,
    val isApk: Boolean = false
) {
    val displaySize: String get() = when {
        isDirectory -> ""
        size < 1024 -> "${size}B"
        size < 1024 * 1024 -> "${size / 1024}KB"
        else -> "${"%.1f".format(size / 1024f / 1024f)}MB"
    }
}