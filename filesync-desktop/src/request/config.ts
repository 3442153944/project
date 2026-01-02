export const config = {
    title: "云梯",
    version: "1.0.0",
    description: "云梯",
    author: "sunyuanling",
    server: "localhost:9999",

    get websocket() {
        const protocol = this.isSecure ? "wss" : "ws";
        return `${protocol}://${this.server}/api/ws/connect`;
    },

    get http() {
        const protocol = this.isSecure ? "https" : "http";
        return `${protocol}://${this.server}/api`;
    },

    get isSecure() {
        // 生产环境默认使用安全协议，开发环境可配置
        return false;
    },

    get apiUrl() {
        return this.http;
    },

    get wsUrl() {
        return this.websocket;
    }
};