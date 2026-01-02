<script setup lang="ts">
import {useRouter} from "vue-router"
const router = useRouter()

//跳转登录页面
function goLogin() {
  router.push("/login")
}

</script>

<template>
  <span>主页面</span>
  <button @click="goLogin">跳转登录页面</button>
</template>

<style scoped>

</style>