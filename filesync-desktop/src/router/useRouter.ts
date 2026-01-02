import {createRouter, createWebHashHistory} from "vue-router"

export const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    {
      path: "/",
      name: "Home",
      component: () => import("../competent/home.vue"),
    },
      {
          path:"/login",
          name:"Login",
          component:()=>import("../competent/login/login.vue")
      }
  ],
});