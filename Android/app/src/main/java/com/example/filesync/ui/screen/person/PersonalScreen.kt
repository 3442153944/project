package com.example.filesync.ui.screen.person

import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import com.example.filesync.network.Request
import kotlinx.coroutines.launch

@Composable
fun PersonalScreen(modifier: Modifier = Modifier) {
    var uiState by remember { mutableStateOf<PersonalUiState>(PersonalUiState.Loading) }
    val scope = rememberCoroutineScope()

    // 初始化时检查登录状态
    LaunchedEffect(Unit) {
        checkLoginStatus { state ->
            uiState = state
        }
    }

    when (val state = uiState) {
        is PersonalUiState.Loading -> {
            LoadingScreen()
        }
        is PersonalUiState.NotLoggedIn -> {
            LoginScreen(
                onLoginSuccess = { userInfo ->
                    uiState = PersonalUiState.LoggedIn(userInfo)
                },
                onLoginError = { error ->
                    uiState = PersonalUiState.Error(error)
                }
            )
        }
        is PersonalUiState.LoggedIn -> {
            UserInfoScreen(
                userInfo = state.userInfo,
                onLogout = {
                    scope.launch {
                        Request.clearToken()
                        uiState = PersonalUiState.NotLoggedIn
                    }
                }
            )
        }
        is PersonalUiState.Error -> {
            ErrorScreen(
                message = state.message,
                onRetry = {
                    scope.launch {
                        uiState = PersonalUiState.Loading
                        checkLoginStatus { newState ->
                            uiState = newState
                        }
                    }
                }
            )
        }
    }
}

// 检查登录状态
private suspend fun checkLoginStatus(onResult: (PersonalUiState) -> Unit) {
    val token = Request.getToken()

    if (token == null) {
        onResult(PersonalUiState.NotLoggedIn)
        return
    }

    Request.post<VerifyResponse>("/auth/verify") { result ->
        result.onSuccess { response ->
            if (response.code == 200) {
                onResult(PersonalUiState.LoggedIn(response.data.userInfo))
            } else {
                onResult(PersonalUiState.NotLoggedIn)
            }
        }.onFailure { error ->
            onResult(PersonalUiState.Error(error.message ?: "验证失败"))
        }
    }
}