import { config } from "./config.ts"

export interface ApiOptions {
    method?: "POST" | "GET" | "PUT" | "DELETE";
    url?: string | null;
    body?: any;
    headers?: Record<string, string>;
}

export const useApi = ({ method = "POST", url = "", body = null, headers = {} }: ApiOptions = {}) => {
    async function api<T = any>() {
        const fullUrl = url
            ? `${config.http}${url.startsWith('/') ? url : `/${url}`}`
            : config.http;

        const defaultHeaders: Record<string, string> = {
            "Content-Type": "application/json",
            ...headers
        };

        // 只有非 GET/DELETE 请求且有 body 时才设置 body
        const hasBody = method !== "GET" && method !== "DELETE" && body !== null;

        const response = await fetch(fullUrl, {
            method,
            headers: new Headers(defaultHeaders),
            body: hasBody ? JSON.stringify(body) : undefined
        });

        if (!response.ok) {
            throw new Error(`请求失败: ${response.status} ${response.statusText}`);
        }

        return await response.json() as T;
    }

    // 返回可复用的请求配置
    const withBody = (newBody: any) => {
        return useApi({ method, url, body: newBody, headers });
    };

    const withHeaders = (newHeaders: Record<string, string>) => {
        return useApi({ method, url, body, headers: { ...headers, ...newHeaders } });
    };

    const withMethod = (newMethod: "POST" | "GET" | "PUT" | "DELETE") => {
        return useApi({ method: newMethod, url, body, headers });
    };

    return {
        api,
        withBody,
        withHeaders,
        withMethod,
        config: { method, url, body, headers }
    };
}
