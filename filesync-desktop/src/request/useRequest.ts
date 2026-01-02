// useRequest.ts
import { ElNotification } from "element-plus"
import { config } from "./config.ts"
import { reactive, computed } from "vue"
import {router} from "../router/useRouter.ts"

export interface ApiResponse<T = any> {
    code: number
    data: T
    message: string
}

type Method = "GET" | "POST" | "PUT" | "DELETE"

export interface RequestOptions {
    showSuccess?: boolean
    showError?: boolean
    successMsg?: string
    errorMsg?: string
    headers?: Record<string, string>
    onSuccess?: (data: any) => void
    onFail?: (error: any) => void
    rethrow?: boolean // 是否在处理后重新抛出
    errorStrategy?: ErrorStrategy
}

// 错误处理策略
export enum ErrorStrategy {
    notify = "notify",   // 弹窗提示
    console = "console", // 控制台输出
    throw = "throw",     // 直接抛出
    silent = "silent"    // 静默处理
}

export interface GlobalConfig extends RequestOptions {
    errorStrategy?: ErrorStrategy
    rethrow?: boolean // 是否在处理后重新抛出
}

class RequestManager {
    private globalConfig: GlobalConfig = {
        showSuccess: false,
        showError: true,
        errorStrategy: ErrorStrategy.notify,
        rethrow: false
    }

    // 全局配置
    config(cfg: Partial<GlobalConfig>) {
        this.globalConfig = { ...this.globalConfig, ...cfg }
        return this
    }

    // 核心请求方法
    private async fetch<T = any>(
        method: Method,
        url: string,
        body?: any,
        options: RequestOptions = {}
    ): Promise<T> {
        const finalConfig = { ...this.globalConfig, ...options }
        const fullUrl = `${config.http}${url.startsWith('/') ? url : `/${url}`}`
        //获取token
        const token = localStorage.getItem("token") || ""

        const headers: Record<string, string> = {
            "Content-Type": "application/json",
            ...finalConfig.headers,
            "Token": token
        }

        const hasBody = method !== "GET" && method !== "DELETE" && body !== null && body !== undefined

        try {
            const response = await fetch(fullUrl, {
                method,
                headers: new Headers(headers),
                body: hasBody ? JSON.stringify(body) : undefined
            })

            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`)
            }

            const res: ApiResponse<T> = await response.json()

            if (res.code === 200 || res.code === 0) {
                if (finalConfig.showSuccess) {
                    ElNotification({
                        type: "success",
                        title: "成功",
                        message: res.message || finalConfig.successMsg || "操作成功"
                    })
                }

                finalConfig.onSuccess?.(res.data)
                return res.data
            }
            // 处理 401 未授权
            if (res.code === 401) {
                localStorage.removeItem("token")
                localStorage.removeItem("userInfo")

                ElNotification({
                    type: "warning",
                    title: "提示",
                    message: "登录已失效，请重新登录"
                })

                // 直接使用导入的 router 实例
                await router.push({name: "Login"})

                // 抛出错误，阻止后续处理
                throw new Error("未授权")
            }
            else {
                throw new Error(res.message || finalConfig.errorMsg || "请求失败")
            }
        } catch (error: any) {
            // 主动取消的请求不报错
            if (error.code === "ERR_CANCELED") {
                throw error
            }

            this.handleError(error, finalConfig)
            finalConfig.onFail?.(error)

            if (finalConfig.rethrow || this.globalConfig.rethrow) {
                throw error
            }

            return null as T
        }
    }

    private handleError(error: any, config: GlobalConfig) {
        const errMsg = error instanceof Error ? error.message : "未知错误"
        const strategy = config.errorStrategy || this.globalConfig.errorStrategy

        switch (strategy) {
            case ErrorStrategy.throw:
                throw error
            case ErrorStrategy.console:
                console.error(errMsg, error)
                break
            case ErrorStrategy.notify:
                if (config.showError !== false) {
                    ElNotification({
                        type: "error",
                        title: "错误",
                        message: errMsg
                    })
                }
                break
            case ErrorStrategy.silent:
                // 静默处理
                break
        }
    }

    // 便捷方法
    get<T = any>(url: string, options?: RequestOptions) {
        return this.fetch<T>("GET", url, null, options)
    }

    post<T = any>(url: string, body?: any, options?: RequestOptions) {
        return this.fetch<T>("POST", url, body, options)
    }

    put<T = any>(url: string, body?: any, options?: RequestOptions) {
        return this.fetch<T>("PUT", url, body, options)
    }

    delete<T = any>(url: string, options?: RequestOptions) {
        return this.fetch<T>("DELETE", url, null, options)
    }

    // 通用请求方法（支持多种参数格式）
    async request<T = any>(
        methodOrUrl: Method | string,
        urlOrBody?: string | any,
        bodyOrOptions?: any | RequestOptions,
        optionsOrUndefined?: RequestOptions
    ): Promise<T> {
        let method: Method = "POST"
        let url: string = ""
        let body: any = null
        let options: RequestOptions = {}

        if (["GET", "POST", "PUT", "DELETE"].includes(methodOrUrl.toUpperCase())) {
            method = methodOrUrl.toUpperCase() as Method
            url = urlOrBody as string
            body = bodyOrOptions
            options = optionsOrUndefined || {}
        } else {
            url = methodOrUrl
            body = urlOrBody
            options = bodyOrOptions || {}
        }

        return this.fetch<T>(method, url, body, options)
    }
}

// 单例导出
export const request = new RequestManager()

// 带状态管理的请求 Hook
export const useRequest = <T extends Record<string, (...args: any[]) => Promise<any>>>(
    apiMap: T,
    defaultData?: Partial<{ [K in keyof T]: Awaited<ReturnType<T[K]>> }> & GlobalConfig
) => {
    const keys = Object.keys(apiMap) as (keyof T)[]
    const options: GlobalConfig = {
        errorStrategy: defaultData?.errorStrategy || ErrorStrategy.notify,
        rethrow: defaultData?.rethrow || false
    }

    // 数据存储
    const data = reactive(
        keys.reduce((acc, key) => {
            acc[key] = defaultData?.[key] ?? null
            return acc
        }, {} as any)
    ) as { [K in keyof T]: Awaited<ReturnType<T[K]>> | null }

    // loading 状态
    const loadingMap = reactive(
        keys.reduce((acc, key) => {
            acc[key] = false
            return acc
        }, {} as any)
    ) as { [K in keyof T]: boolean }

    const loading = computed(() => Object.values(loadingMap).some(Boolean))

    // 请求方法
    const run = async <K extends keyof T>(
        key: K,
        params?: Parameters<T[K]>[0],
        options?: RequestOptions
    ): Promise<Awaited<ReturnType<T[K]>> | null> => {
        loadingMap[key] = true

        try {
            const res = await apiMap[key](params)
            data[key] = res

            if (options?.onSuccess) {
                options.onSuccess(res)
            }

            if (options?.showSuccess) {
                ElNotification({
                    type: "success",
                    title: "成功",
                    message: options.successMsg || "操作成功"
                })
            }

            return res
        } catch (error: any) {
            if (error.code === "ERR_CANCELED") {
                throw error
            }

            data[key] = defaultData?.[key] ?? null

            if (options?.showError !== false) {
                const strategy = options?.errorStrategy || defaultData?.errorStrategy || ErrorStrategy.notify
                const errMsg = error instanceof Error ? error.message : "未知错误"

                switch (strategy) {
                    case ErrorStrategy.notify:
                        ElNotification({
                            type: "error",
                            title: "错误",
                            message: errMsg
                        })
                        break
                    case ErrorStrategy.console:
                        console.error(errMsg, error)
                        break
                }
            }

            if (options?.onFail) {
                options.onFail(error)
            }

            if (options?.rethrow || defaultData?.rethrow) {
                throw error
            }

            return null
        } finally {
            loadingMap[key] = false
        }
    }

    return {
        data,
        loadingMap,
        loading,
        run
    }
}