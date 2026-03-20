export interface DiskInfo {
    path: string
    mountpoint: string
    device: string
    fstype: string
    total: number
    free: number
    used: number
    used_percent: number
    total_gb: string
    free_gb: string
    is_allowed: boolean
    is_accessible: boolean
    is_ssd: boolean
}

// 根据你的 JSON 定义类型
export interface FileItem {
    name: string
    path: string
    is_dir: boolean
    mod_time: string
    mode: string
    children_count?: number
    size?: number // 如果后端有返回大小的话
}