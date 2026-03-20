<script setup lang="ts">
import {onMounted, ref, computed} from "vue"
import {useRouter, useRoute} from "vue-router"
import {useLogin} from "./login/login.ts"
import {useMenuConfig} from "./composeables/menu-config.ts"

// 1. 按需引入 Naive UI 的组件和类型 (不需要全局注册)
import {NMenu, NButton} from "naive-ui"
import type {MenuOption} from "naive-ui"

const router = useRouter()
const route = useRoute() // 引入 route 用于菜单高亮
const {verify} = useLogin()
const {menuConfig} = useMenuConfig()

const userInfo = ref<any>(null)

// 2. 数据转换：将你的 menuConfig 转为 Naive UI 要求的 MenuOption 格式
const mapMenus = (menus: any[]): MenuOption[] => {
  return menus.map(item => {
    const option: MenuOption = {
      label: item.name,
      key: item.path, // 使用 path 作为唯一 key
    }
    // 如果有子节点，递归处理
    if (item.children && item.children.length > 0) {
      option.children = mapMenus(item.children)
    }
    return option
  })
}

// 使用 computed 确保数据响应式
const menuOptions = computed(() => mapMenus(menuConfig || []))

// 自动匹配当前路由高亮菜单
const activeKey = computed(() => route.path)

onMounted(async () => {
  try {
    await verify()
    const saved = localStorage.getItem("userInfo")
    if (saved) userInfo.value = JSON.parse(saved)
  } catch {
    localStorage.removeItem("token")
    await router.push("/login")
  }
})

const handleLogout = () => {
  localStorage.removeItem("token")
  localStorage.removeItem("userInfo")
  router.push("/login")
}

// 3. 菜单点击事件：Naive UI 会直接传入对应的 key (即 item.path)
const handleMenuClick = (key: string) => {
  router.push(key)
}
</script>

<template>
  <div class="layout">
    <div class="header">
      <div class="header-left">
        <div class="logo">云梯</div>

        <n-menu
            mode="horizontal"
            :value="activeKey"
            :options="menuOptions"
            @update:value="handleMenuClick"
        />
      </div>

      <div class="header-right">
        <span v-if="userInfo" class="username">{{ userInfo.username }}</span>
        <n-button type="primary" size="small" @click="handleLogout">
          退出登录
        </n-button>
      </div>
    </div>

    <div class="content">
      <router-view/>
    </div>
  </div>
</template>

<style scoped>
.layout {
  width: 100%;
  height: 100vh;
  display: flex;
  flex-direction: column;
}

.header {
  height: 60px;
  background: white;
  border-bottom: 1px solid #e8e8e8;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 32px;
}

.logo {
  font-size: 20px;
  font-weight: bold;
  color: #1890ff; /* 后续你可以用 Naive UI 的主题变量替换这里的硬编码颜色 */
}

.header-right {
  display: flex;
  align-items: center;
  gap: 16px;
}

.username {
  color: #666;
  font-size: 14px;
}

.content {
  flex: 1;
  background: #f0f2f5;
  padding: 20px;
  overflow: auto;
}

/* 所有关于 dropdown、hover、arrow 动画的恶心 CSS 都可以删掉了！
  Naive UI 内部已经处理好了绝佳的过渡动画和阴影。
*/
</style>