import { createRouter, createWebHistory, RouteRecordRaw, RouteLocationNormalized } from "vue-router"
import Home from "../views/Home.vue"

const routes: Array<RouteRecordRaw> = [
  {
    path: "/",
    name: "home",
    component: Home,
  },
  {
    path: "/login",
    name: "login",
    component: () =>
      import(/* webpackChunkName: "login" */ "../views/Login.vue"),
  },
  {
    path: "/:confa",
    name: "confaOverview",
    props(to: RouteLocationNormalized): Record<string, any> {
      return {
        handle: to.params.confa,
        tab: 'overview',
      }
    },
    component: () =>
      import(/* webpackChunkName: "confaOverview" */ "../views/confa/Confa.vue"),
  },
  {
    path: "/:confa/edit",
    name: "confaEdit",
    props(to: RouteLocationNormalized): Record<string, any> {
      return {
        handle: to.params.confa,
        tab: 'edit',
      }
    },
    component: () =>
      import(/* webpackChunkName: "confaEdit" */ "../views/confa/Confa.vue"),
  },
  {
    path: "/:confa/:talk",
    name: "talk",
    component: () => import(/* webpackChunkName: "talk" */ "../views/Talk.vue"),
  },
  {
    path: "/t/:confa/:talk",
    name: "rtc",
    component: () =>
      import(/* webpackChunkName: "rtc" */ "../views/RTCExample.vue"),
  },
  {
    path: "/stream",
    name: "stream",
    component: () =>
      import(/* webpackChunkName: "stream" */ "../views/Stream.vue"),
  },
]

const router = createRouter({
  history: createWebHistory(process.env.BASE_URL),
  routes,
})

export default router
