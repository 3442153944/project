<script setup lang="ts">
import {onMounted, h} from "vue"
import {useRoute, useRouter} from "vue-router"
import {storeToRefs} from "pinia"
import {useCatalogStore} from "./composeables/useCatalogStore"
import {NDataTable, NButton, NSpace, NSpin, NEmpty, NTag} from "naive-ui"
import type {DataTableColumns} from "naive-ui"
import type {FileItem} from "@syl/models"

const route = useRoute()
const router = useRouter()
const catalogStore = useCatalogStore()

// 使用 storeToRefs 保持响应式解构
const {currentPath, parentPath, items, loading} = storeToRefs(catalogStore)

// 首次进入时，从路由参数获取目标路径 (例如从磁盘列表跳过来)
onMounted(() => {
  const initPath = route.query.path as string
  if (initPath) {
    catalogStore.fetchDirectory(initPath)
  }
})

// 处理表格行的双击/点击事件
const handleRowClick = (row: FileItem) => {
  if (row.is_dir) {
    // 是目录：请求下一级，并同步更新路由 URL (方便浏览器前进后退)
    catalogStore.fetchDirectory(row.path)
    router.push({query: {path: row.path}})
  } else {
    // 是文件：后续在这里对接预览或下载逻辑
    console.log("准备预览/下载文件:", row.path)
  }
}

// 返回上一级
const handleGoUp = () => {
  catalogStore.goParent()
  router.push({query: {path: parentPath.value}})
}

// 格式化时间的小工具
const formatTime = (isoString: string) => {
  if (!isoString) return "-"
  const date = new Date(isoString)
  return date.toLocaleString("zh-CN", {
    year: "numeric", month: "2-digit", day: "2-digit",
    hour: "2-digit", minute: "2-digit", second: "2-digit"
  })
}

// 定义 Naive UI 表格列
const columns: DataTableColumns<FileItem> = [
  {
    title: "名称",
    key: "name",
    sorter: "default", // 开启默认排序
    render(row) {
      // 简单用 Emoji 区分文件和文件夹，后期可以换成 NIcon + 图标库
      const icon = row.is_dir ? "📁" : "📄"
      return h(
          "div",
          {
            style: {display: "flex", alignItems: "center", gap: "8px", cursor: "pointer"},
            onClick: () => handleRowClick(row) // 点击名称进入
          },
          [
            h("span", {style: {fontSize: "18px"}}, icon),
            h("span", {style: {fontWeight: row.is_dir ? "bold" : "normal"}}, row.name)
          ]
      )
    }
  },
  {
    title: "修改时间",
    key: "mod_time",
    render(row) {
      return formatTime(row.mod_time)
    }
  },
  {
    title: "属性",
    key: "mode",
    render(row) {
      const perms = catalogStore.parseMode(row.mode)

      // 使用 NSpace 组合多个 NTag，让它们横向排列并且有间距
      return h(
          NSpace,
          {size: 'small'},
          {
            default: () => perms.map(p =>
                h(
                    NTag,
                    {
                      size: "small",
                      type: p.type as any, // 'info' | 'success' | 'warning' | 'error'
                      bordered: false
                    },
                    {default: () => p.label}
                )
            )
          }
      )
    }
  },
  {
    title: "包含项",
    key: "children_count",
    render(row) {
      return row.is_dir ? `${row.children_count} 项` : "-"
    }
  }
]
</script>

<template>
  <div class="catalog-container">
    <div class="toolbar">
      <n-space align="center">
        <n-button
            v-if="parentPath"
            type="primary"
            ghost
            size="small"
            @click="handleGoUp"
        >
          ⬆ 返回上一级
        </n-button>

        <div class="current-path">
          <span class="path-label">当前路径：</span>
          <span class="path-value">{{ currentPath || "加载中..." }}</span>
        </div>
      </n-space>
    </div>

    <div class="table-wrapper">
      <n-data-table
          :columns="columns"
          :data="items"
          :loading="loading"
          :row-key="(row) => row.path"
          :bordered="false"
          :striped="true"
          size="small"
          :max-height="'calc(100vh - 200px)'"
      >
        <template #empty>
          <n-empty description="此文件夹为空"/>
        </template>
      </n-data-table>
    </div>
  </div>
</template>

<style scoped>
.catalog-container {
  padding: 16px;
  background-color: #fff;
  border-radius: 8px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.05);
  height: 100%;
  display: flex;
  flex-direction: column;
}

.toolbar {
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid #f0f0f0;
}

.current-path {
  font-size: 14px;
  padding: 4px 8px;
  background-color: #f5f7fa;
  border-radius: 4px;
}

.path-label {
  color: #909399;
}

.path-value {
  color: #303133;
  font-family: monospace;
  font-weight: bold;
}

.table-wrapper {
  flex: 1;
  overflow: hidden;
}

/* 隐藏表格悬浮时的默认背景色，改用更清爽的样式 */
:deep(.n-data-table-tr:hover) {
  background-color: #f0f7ff !important;
}
</style>