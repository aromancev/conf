import { createRouter, createWebHistory, RouteRecordRaw, RouteLocationNormalized } from "vue-router"

const routes: Array<RouteRecordRaw> = [
  {
    path: "/",
    name: "home",
    component: () => import("@/views/HomePage.vue"),
  },
  {
    path: "/login",
    name: "login",
    props(to: RouteLocationNormalized): Record<string, string> {
      return {
        token: to.query.token as string,
      }
    },
    component: () => import("@/views/LoginPage.vue"),
  },
  {
    path: "/:confa",
    name: "confaOverview",
    props(to: RouteLocationNormalized): Record<string, string> {
      return {
        handle: to.params.confa as string,
        tab: "overview",
      }
    },
    component: () => import("@/views/confa/Confa.vue"),
  },
  {
    path: "/:confa/edit",
    name: "confaEdit",
    props(to: RouteLocationNormalized): Record<string, string> {
      return {
        handle: to.params.confa as string,
        tab: "edit",
      }
    },
    component: () => import("@/views/confa/Confa.vue"),
  },
  {
    path: "/:confa/:talk",
    name: "talkOverview",
    props(to: RouteLocationNormalized): Record<string, string> {
      return {
        handle: to.params.talk as string,
        confaHandle: to.params.confa as string,
        tab: "overview",
      }
    },
    component: () => import("../views/Talk.vue"),
  },
  {
    path: "/t/:confa/:talk",
    name: "rtc",
    component: () => import("@/views/RTCExample.vue"),
  },
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
})

export default router
