package com.example.filesync.ui.screen.person

import kotlinx.serialization.Serializable

@Serializable
data class VerifyResponse(
    val code: Int,
    val message: String,
    val data: VerifyData
)

@Serializable
data class VerifyData(
    val msg: String,
    val userInfo: UserInfo
)

@Serializable
data class UserInfo(
    val user_id: Int,
    val username: String,
    val email: String,
    val issued_at: Long,
    val expires_at: Long
)

@Serializable
data class LoginRequest(
    val username: String,
    val password: String
)

@Serializable
data class LoginResponse(
    val code: Int,
    val message: String,
    val data: LoginData
)

@Serializable
data class LoginData(
    val token: String,
    val user: User
)

@Serializable
data class User(
    val avatar: String,
    val email: String,
    val id: Int,
    val last_login: String,
    val phone: String,
    val role: String,
    val status: Int,
    val username: String
)