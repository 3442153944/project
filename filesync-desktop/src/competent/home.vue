<script setup lang="ts">
import {onMounted, ref} from "vue"
import {useRouter} from "vue-router"
import {useLogin} from "./login/login.ts"
import {useMenuConfig} from "./composeables/menu-config.ts"
import {ArrowDown, ArrowRight} from "@element-plus/icons-vue"

const router = useRouter()
const {verify} = useLogin()
const {menuConfig} = useMenuConfig()

const userInfo = ref<any>(null)

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

const handleMenuClick = (path: string) => {
  router.push(path)
}
</script>

<template>
  <div class="layout">
    <div class="header">
      <div class="header-left">
        <div class="logo">云梯</div>

        <div class="menu-bar">
          <template v-for="item in menuConfig" :key="item.path">
            <!-- 有子菜单 -->
            <el-dropdown
                v-if="item.children && item.children.length"
                trigger="hover"
                @command="handleMenuClick"
            >
              <span class="menu-item">
                {{ item.name }}
                <el-icon class="menu-arrow"><ArrowDown/></el-icon>
              </span>

              <template #dropdown>
                <el-dropdown-menu>
                  <template v-for="child in item.children" :key="child.path">
                    <!-- 有子项，继续嵌套 -->
                    <el-dropdown-item v-if="child.children && child.children.length" class="has-children">
                      <el-dropdown
                          trigger="hover"
                          placement="right-start"
                          @command="handleMenuClick"
                      >
                        <span class="submenu-item">
                          {{ child.name }}
                          <el-icon class="submenu-arrow"><ArrowRight/></el-icon>
                        </span>

                        <template #dropdown>
                          <el-dropdown-menu>
                            <el-dropdown-item
                                v-for="subChild in child.children"
                                :key="subChild.path"
                                :command="subChild.path"
                            >
                              {{ subChild.name }}
                            </el-dropdown-item>
                          </el-dropdown-menu>
                        </template>
                      </el-dropdown>
                    </el-dropdown-item>

                    <!-- 无子项，直接渲染 -->
                    <el-dropdown-item v-else :command="child.path">
                      {{ child.name }}
                    </el-dropdown-item>
                  </template>
                </el-dropdown-menu>
              </template>
            </el-dropdown>

            <!-- 无子菜单，直接点击 -->
            <span v-else class="menu-item" @click="handleMenuClick(item.path)">
              {{ item.name }}
            </span>
          </template>
        </div>
      </div>

      <div class="header-right">
        <span v-if="userInfo" class="username">{{ userInfo.username }}</span>
        <el-button type="primary" size="small" @click="handleLogout">
          退出登录
        </el-button>
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
  color: #1890ff;
}

.menu-bar {
  display: flex;
  align-items: center;
  gap: 8px;
}

.menu-item {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  height: 40px;
  padding: 0 16px;
  cursor: pointer;
  font-size: 14px;
  color: #333;
  border-radius: 4px;
  transition: all 0.2s;
  user-select: none;
}

.menu-item:hover {
  background: #f0f0f0;
  color: #1890ff;
}

.menu-arrow {
  font-size: 12px;
  transition: transform 0.2s;
}

.menu-item:hover .menu-arrow {
  transform: rotate(180deg);
}

.submenu-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  min-width: 140px;
}

.submenu-arrow {
  margin-left: 12px;
  font-size: 12px;
  color: #999;
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

/* 下拉菜单样式优化 */
:deep(.el-dropdown-menu) {
  padding: 4px 0;
}

:deep(.el-dropdown-menu__item) {
  padding: 8px 16px;
  font-size: 14px;
  line-height: 22px;
  color: #333;
}

:deep(.el-dropdown-menu__item.has-children) {
  padding: 0;
}

:deep(.el-dropdown-menu__item.has-children .submenu-item) {
  padding: 8px 16px;
}

:deep(.el-dropdown-menu__item:hover) {
  background: #f5f5f5;
  color: #1890ff;
}

:deep(.el-dropdown-menu__item:not(.is-disabled):focus) {
  background: #f5f5f5;
  color: #1890ff;
}
</style>