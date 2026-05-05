<!-- Register.vue -->
<script setup lang="ts">
import {ref} from 'vue'
import {useMessage, NForm, NFormItem, NInput, NButton} from 'naive-ui'
import type {FormInst, FormRules} from 'naive-ui'
import {useRegister} from './register'
import {useRouter} from 'vue-router'

const message = useMessage()
const router = useRouter()
const formRef = ref<FormInst | null>(null)
const {loading, register} = useRegister()

const form = ref({
  username: '',
  password: '',
  confirmPassword: '',
  email: '',
  phone: '',
})

const rules: FormRules = {
  username: [
    {min: 3, message: '用户名至少3位', trigger: 'blur'},
  ],
  password: [
    {required: true, message: '请输入密码', trigger: 'blur'},
    {min: 6, message: '密码至少6位', trigger: 'blur'},
  ],
  confirmPassword: [
    {required: true, message: '请确认密码', trigger: 'blur'},
    {
      validator: (_rule, value) => {
        if (value !== form.value.password) return new Error('两次密码不一致')
        return true
      },
      trigger: 'blur',
    },
  ],
  email: [
    {type: 'email', message: '邮箱格式不正确', trigger: 'blur'},
  ],
}

const handleRegister = () => {
  formRef.value?.validate(async (errors) => {
    if (errors) return

    const {username, email, phone} = form.value
    if (!username && !email && !phone) {
      message.warning('请至少填写用户名、邮箱或手机号其中一项')
      return
    }

    await register({
      username: username || undefined,
      password: form.value.password,
      email: email || undefined,
      phone: phone || undefined,
    })

    message.success('注册成功')
    await router.push('/login')
  })
}
</script>

<template>
  <div class="register">
    <n-card title="注册账号" class="register-card">
      <n-form ref="formRef" :model="form" :rules="rules" label-placement="top">
        <n-form-item label="用户名" path="username">
          <n-input v-model:value="form.username" placeholder="至少3位，可不填"/>
        </n-form-item>

        <n-form-item label="邮箱" path="email">
          <n-input v-model:value="form.email" placeholder="可不填"/>
        </n-form-item>

        <n-form-item label="手机号" path="phone">
          <n-input v-model:value="form.phone" placeholder="可不填"/>
        </n-form-item>

        <n-form-item label="密码" path="password">
          <n-input
              v-model:value="form.password"
              type="password"
              show-password-on="click"
              placeholder="至少6位"
          />
        </n-form-item>

        <n-form-item label="确认密码" path="confirmPassword">
          <n-input
              v-model:value="form.confirmPassword"
              type="password"
              show-password-on="click"
              placeholder="再次输入密码"
          />
        </n-form-item>

        <n-button type="primary" block :loading="loading" @click="handleRegister">
          注册
        </n-button>
      </n-form>

      <div class="register-footer">
        已有账号？
        <n-button text type="primary" @click="router.push('/login')">去登录</n-button>
      </div>
    </n-card>
  </div>
</template>

<style scoped>
.register {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
}

.register-card {
  width: 400px;
}

.register-footer {
  margin-top: 16px;
  text-align: center;
  font-size: 14px;
}
</style>