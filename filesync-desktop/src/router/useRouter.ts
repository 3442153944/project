// router/index.ts
import {createRouter, createWebHashHistory} from "vue-router"

export const router = createRouter({
    history: createWebHashHistory(),
    routes: [
        {
            path: "/",
            name: "Home",
            component: () => import("../competent/home.vue"),
            redirect: "/dashboard", // 默认重定向
            children: [
                {
                    path: "dashboard",
                    name: "Dashboard",
                    component: () => import("../views/Dashboard.vue")
                },
                {
                    path: "file/list",
                    name: "FileList",
                    component: () => import("../views/file/List.vue")
                },
                {
                    path: "file/upload",
                    name: "FileUpload",
                    component: () => import("../views/file/Upload.vue")
                },
                {
                    path: "monitor/system",
                    name: "MonitorSystem",
                    component: () => import("../views/monitor/System.vue")
                },
                {
                    path: "monitor/network",
                    name: "MonitorNetwork",
                    component: () => import("../views/monitor/Network.vue")
                }
            ]
        },
        {
            path: "/login",
            name: "Login",
            component: () => import("../competent/login/login.vue")
        },

        // 捕获所有未匹配路由，重定向到首页
        {
            path: "/:pathMatch(.*)*",
            name: "NotFound",
            redirect: "/"
        }
    ]
})

//路由守卫增强
router.beforeEach(async (to, from, next) => {
    const token = localStorage.getItem("token")

    // 如果去登录页
    if (to.path === "/login") {
        if (token) {
            // 已登录，跳转首页
            next("/")
        } else {
            next()
        }
        return
    }

    // 需要登录的页面
    if (!token) {
        next("/login")
        return
    }

    next()
})

// 路由错误处理
router.onError(async (error) => {
    console.error("路由错误:", error)
    // 路由加载失败，回到首页
    await router.push("/")
})