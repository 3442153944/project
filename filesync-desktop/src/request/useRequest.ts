// useRequest.ts
import {config} from "./config.ts"
import {reactive, computed} from "vue"
import {router} from "../router/useRouter.ts"
import {createDiscreteApi} from 'naive-ui'

// 1. 初始化 Naive UI 的离散 API (用于非组件环境弹窗)
const {message: naiveMessage, notification: naiveNotification} = createDiscreteApi(
    ['message', 'notification']
)

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
    rethrow?: boolean
    errorStrategy?: ErrorStrategy
}

export enum ErrorStrategy {
    notify = "notify",
    message = "message", // 新增 message 策略 (更轻量)
    console = "console",
    throw = "throw",
    silent = "silent"
}

export interface GlobalConfig extends RequestOptions {
    errorStrategy?: ErrorStrategy
    rethrow?: boolean
}

// 自定义错误类，方便携带更多信息
class ApiError extends Error {
    code?: number;
    data?: any;

    constructor(message: string, code?: number, data?: any) {
        super(message);
        this.name = 'ApiError';
        this.code = code;
        this.data = data;
    }
}

class RequestManager {
    private globalConfig: GlobalConfig = {
        showSuccess: false,
        showError: true,
        errorStrategy: ErrorStrategy.message, // 默认改为 message，notification 有点太重
        rethrow: false
    }

    config(cfg: Partial<GlobalConfig>) {
        this.globalConfig = {...this.globalConfig, ...cfg}
        return this
    }

    private async fetch<T = any>(
        method: Method,
        url: string,
        body?: any,
        options: RequestOptions = {}
    ): Promise<T> {
        const finalConfig = {...this.globalConfig, ...options}
        const fullUrl = `${config.http}${url.startsWith('/') ? url : `/${url}`}`
        const token = localStorage.getItem("token") || ""

        const headers: Record<string, string> = {
            "Content-Type": "application/json",
            ...finalConfig.headers,
            ...(token ? {"Token": token} : {}) // 只有 token 存在时才附加
        }

        const hasBody = method !== "GET" && method !== "DELETE" && body !== null && body !== undefined

        try {
            const response = await fetch(fullUrl, {
                method,
                headers: new Headers(headers),
                body: hasBody ? JSON.stringify(body) : undefined
            })

            //刷新token
            if (response.headers.get("Token-Refreshed") === "true") {
                const newToken = response.headers.get("New-Token")
                if (newToken) {
                    localStorage.setItem("token", newToken)
                    naiveMessage.success("Token已自动刷新")
                }
            }

            // 1. 尝试解析响应体 (无论是成功还是失败状态码)
            let resData: any = null;
            const contentType = response.headers.get("content-type");
            if (contentType && contentType.includes("application/json")) {
                resData = await response.json().catch(() => null);
            }

            // 2. 处理 HTTP 层面错误 (如 404, 500)
            if (!response.ok) {
                // 如果后端返回了标准的 JSON 错误信息，优先使用
                const errMsg = resData?.message || `HTTP 请求错误: ${response.status} ${response.statusText}`;
                throw new ApiError(errMsg, response.status, resData);
            }

            // 如果不是 JSON，直接返回原始文本或 Blob 等 (根据需要扩展，目前默认按 JSON 处理)
            if (!resData) {
                throw new ApiError("无效的响应格式", response.status);
            }

            const res = resData as ApiResponse<T>;

            // 3. 处理业务状态码
            if (res.code === 200 || res.code === 0) {
                if (finalConfig.showSuccess) {
                    naiveMessage.success(res.message || finalConfig.successMsg || "操作成功")
                }
                finalConfig.onSuccess?.(res.data)
                return res.data
            }

            // 处理 401 未授权
            if (res.code === 401) {
                localStorage.removeItem("token")
                localStorage.removeItem("userInfo")

                naiveMessage.warning("登录已失效，请重新登录")
                await router.push({name: "Login"})

                throw new ApiError("未授权", 401); // 抛出异常中断当前请求流
            }

            // 其他业务错误
            throw new ApiError(res.message || finalConfig.errorMsg || "请求失败", res.code, res.data);

        } catch (error: any) {
            // 取消请求不处理
            if (error.name === "AbortError" || error.code === "ERR_CANCELED") {
                throw error
            }

            this.handleError(error, finalConfig)
            finalConfig.onFail?.(error)

            if (finalConfig.rethrow) {
                throw error
            }

            return Promise.reject(error) // 之前你 return null，这会导致外部 try-catch 无法捕获且类型不对，应该 reject
        }
    }

    private handleError(error: any, config: RequestOptions) {
        if (config.showError === false) return;

        const errMsg = error instanceof Error ? error.message : "未知错误"
        const strategy = config.errorStrategy || this.globalConfig.errorStrategy

        switch (strategy) {
            case ErrorStrategy.throw:
                // 这里不处理，交给外层 catch
                break
            case ErrorStrategy.console:
                console.error("请求异常:", errMsg, error)
                break
            case ErrorStrategy.notify:
                naiveNotification.error({
                    title: "请求错误",
                    content: errMsg,
                    duration: 3000
                })
                break
            case ErrorStrategy.message:
                naiveMessage.error(errMsg)
                break
            case ErrorStrategy.silent:
                break
        }
    }

    // ... (get, post, put, delete 方法保持不变)
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
}

export const request = new RequestManager()

// --- useRequest Hook 保持原有设计思路，稍作调整 ---
export const useRequest = <T extends Record<string, (...args: any[]) => Promise<any>>>(
    apiMap: T,
    defaultData?: Partial<{ [K in keyof T]: Awaited<ReturnType<T[K]>> }> & GlobalConfig
) => {
    const keys = Object.keys(apiMap) as (keyof T)[]

    const data = reactive(
        keys.reduce((acc, key) => {
            acc[key] = defaultData?.[key as string] ?? null
            return acc
        }, {} as any)
    ) as { [K in keyof T]: Awaited<ReturnType<T[K]>> | null }

    const loadingMap = reactive(
        keys.reduce((acc, key) => {
            acc[key] = false
            return acc
        }, {} as any)
    ) as { [K in keyof T]: boolean }

    const loading = computed(() => Object.values(loadingMap).some(Boolean))

    const run = async <K extends keyof T>(
        key: K,
        params?: Parameters<T[K]>[0],
        options?: RequestOptions
    ): Promise<Awaited<ReturnType<T[K]>> | null> => {
        loadingMap[key] = true

        try {
            const res = await apiMap[key](params)
            data[key] = res

            if (options?.showSuccess) {
                naiveMessage.success(options.successMsg || "操作成功")
            }
            options?.onSuccess?.(res)

            return res
        } catch (error: any) {
            data[key] = defaultData?.[key as string] ?? null
            options?.onFail?.(error)

            // 错误提示已在 RequestManager 中处理，这里只需处理是否重新抛出
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