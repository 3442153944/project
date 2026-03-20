import {defineStore} from "pinia"
import {ref} from "vue"
import {request} from "@syl/base-request"
import {FileItem} from "@syl/models";


export const useCatalogStore = defineStore("catalog", () => {
    const currentPath = ref<string>("")
    const parentPath = ref<string>("")
    const items = ref<FileItem[]>([])
    const loading = ref<boolean>(false)

    // 核心拉取方法
    const fetchDirectory = async (targetPath: string) => {
        loading.value = true
        try {
            // 注意：你的 JSON 示例中键名是大写的 "Path"
            const res = await request.post("/files/traverse-directory", {Path: targetPath})
            if (res) {
                currentPath.value = res.current_path
                parentPath.value = res.parent_path
                items.value = res.items || []
            }
        } catch (error) {
            console.error("读取目录失败", error)
        } finally {
            loading.value = false
        }
    }

    // 返回上一级
    const goParent = async () => {
        if (parentPath.value) {
            await fetchDirectory(parentPath.value)
        }
    }

    const parseMode = (modeStr: string) => {
        if (!modeStr || modeStr.length < 4) return []

        const tags = []
        // 提取前三个代表权限的字符 (索引 1, 2, 3)
        const ownerPerms = modeStr.substring(1, 4)

        if (ownerPerms.includes('r')) {
            tags.push({label: '可读', type: 'info'})
        }
        if (ownerPerms.includes('w')) {
            tags.push({label: '可写', type: 'success'})
        }
        if (ownerPerms.includes('x')) {
            // 对于文件夹来说，x 通常意味着“可进入/遍历”
            // 对于文件来说，x 是“可执行”
            const isDir = modeStr.charAt(0) === 'd'
            tags.push({
                label: isDir ? '可进入' : '可执行',
                type: 'warning'
            })
        }

        // 如果什么都没有，或者后端传来的格式不对
        if (tags.length === 0) {
            tags.push({label: '无权限', type: 'error'})
        }

        return tags
    }

    return {
        currentPath,
        parentPath,
        items,
        loading,
        fetchDirectory,
        goParent,
        parseMode
    }
})