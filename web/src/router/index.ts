import { createRouter, createWebHistory, RouteRecordRaw } from "vue-router"
import Home from "../views/Home.vue"

const routes: Array<RouteRecordRaw> = [
  {
    path: "/",
    name: "Home",
    component: Home,
  },
  {
    path: "/login",
    name: "Login",
    component: () =>
      import(/* webpackChunkName: "login" */ "../views/Login.vue"),
  },
  {
    path: "/:confa",
    name: "Confa",
    component: () =>
      import(/* webpackChunkName: "confa" */ "../views/Confa.vue"),
  },
  {
    path: "/:confa/:talk",
    name: "Talk",
    component: () => import(/* webpackChunkName: "talk" */ "../views/Talk.vue"),
  },
  {
    path: "/stream",
    name: "Stream",
    component: () =>
      import(/* webpackChunkName: "stream" */ "../views/Stream.vue"),
  },
]

const router = createRouter({
  history: createWebHistory(process.env.BASE_URL),
  routes,
})

export default router
