<script setup lang="ts">
import { ref } from "vue"
import { useRouter } from "vue-router"
import { useLogin } from "./login.ts"
import { ElMessage } from "element-plus"

const router = useRouter()
const { login } = useLogin()

const loginForm = ref({
  username: "",
  password: ""
})

const loading = ref(false)
const rememberMe = ref(false)

// 页面加载时检查是否有记住的账号
const loadRememberedAccount = () => {
  const remembered = localStorage.getItem("rememberedAccount")
  if (remembered) {
    const account = JSON.parse(remembered)
    loginForm.value.username = account.username
    rememberMe.value = true
  }
}

loadRememberedAccount()

const handleLogin = async () => {
  if (!loginForm.value.username || !loginForm.value.password) {
    ElMessage.warning("请输入用户名和密码")
    return
  }

  loading.value = true

  try {
    const res = await login(loginForm.value)

    // 保存 token
    localStorage.setItem("token", res.token)

    // 保存用户信息
    localStorage.setItem("userInfo", JSON.stringify(res.user))

    // 记住账号
    if (rememberMe.value) {
      localStorage.setItem("rememberedAccount", JSON.stringify({
        username: loginForm.value.username
      }))
    } else {
      localStorage.removeItem("rememberedAccount")
    }

    ElMessage.success("登录成功")

    // 跳转到首页
    await router.push({ name: "Home" })
  } catch (error) {
    console.error("登录失败", error)
  } finally {
    loading.value = false
  }
}

// 回车登录
const handleKeydown = (e: KeyboardEvent) => {
  if (e.key === "Enter") {
    handleLogin()
  }
}
</script>

<template>
  <div class="login" @keydown="handleKeydown">
    <div class="login-container">
      <div class="login-header">
        <h1>私有云系统</h1>
        <p>File Sync Platform</p>
      </div>

      <div class="login-form">
        <el-form :model="loginForm" label-position="top">
          <el-form-item label="用户名">
            <el-input
                v-model="loginForm.username"
                placeholder="请输入用户名"
                size="large"
                clearable
                :prefix-icon="User"
            />
          </el-form-item>

          <el-form-item label="密码">
            <el-input
                v-model="loginForm.password"
                type="password"
                placeholder="请输入密码"
                size="large"
                show-password
                :prefix-icon="Lock"
            />
          </el-form-item>

          <el-form-item>
            <div class="login-options">
              <el-checkbox v-model="rememberMe">
                记住账号
              </el-checkbox>
              <el-link type="primary" :underline="false">
                忘记密码?
              </el-link>
            </div>
          </el-form-item>

          <el-form-item>
            <el-button
                type="primary"
                size="large"
                :loading="loading"
                @click="handleLogin"
                class="login-button"
            >
              {{ loading ? "登录中..." : "登录" }}
            </el-button>
          </el-form-item>
        </el-form>
      </div>

      <div class="login-footer">
        <p>© 2025 私有云系统 - 多端文件同步平台</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.login {
  width: 100%;
  height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  position: relative;
  overflow: hidden;
}

.login::before {
  content: "";
  position: absolute;
  width: 200%;
  height: 200%;
  background-image:
      radial-gradient(circle, rgba(255, 255, 255, 0.1) 1px, transparent 1px);
  background-size: 50px 50px;
  animation: moveBackground 20s linear infinite;
}

@keyframes moveBackground {
  0% {
    transform: translate(0, 0);
  }
  100% {
    transform: translate(50px, 50px);
  }
}

.login-container {
  width: 420px;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  border-radius: 16px;
  padding: 40px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  position: relative;
  z-index: 1;
}

.login-header {
  text-align: center;
  margin-bottom: 40px;
}

.login-header h1 {
  font-size: 28px;
  font-weight: 600;
  color: #333;
  margin: 0 0 8px 0;
}

.login-header p {
  font-size: 14px;
  color: #666;
  margin: 0;
}

.login-form {
  margin-bottom: 20px;
}

.login-options {
  width: 100%;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.login-button {
  width: 100%;
  height: 48px;
  font-size: 16px;
  font-weight: 500;
  border-radius: 8px;
}

.login-footer {
  text-align: center;
  padding-top: 20px;
  border-top: 1px solid #eee;
}

.login-footer p {
  font-size: 12px;
  color: #999;
  margin: 0;
}

:deep(.el-input__wrapper) {
  border-radius: 8px;
}

:deep(.el-button--primary) {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: none;
}

:deep(.el-button--primary:hover) {
  background: linear-gradient(135deg, #5568d3 0%, #6a4292 100%);
}
</style>