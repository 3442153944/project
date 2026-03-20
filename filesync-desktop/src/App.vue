<script setup lang="ts">
import {NMessageProvider, NDialogProvider, NNotificationProvider} from 'naive-ui'
</script>

<template>
  <n-message-provider>
    <n-notification-provider>
      <n-dialog-provider>

        <div class="main">
          <router-view/>
        </div>

      </n-dialog-provider>
    </n-notification-provider>
  </n-message-provider>
</template>

<style scoped>
.main {
  /* 你的外层容器样式，比如 100vh 满屏等 */
  height: 100vh;
  width: 100vw;
}
</style>

<style>
/* 全局样式 */
</style>