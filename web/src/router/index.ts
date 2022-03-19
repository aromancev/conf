import { createRouter, createWebHistory, RouteRecordRaw, RouteLocationNormalized, RouteLocationRaw } from "vue-router"

export const handleNew = "new"

export enum Name {
  Home = "home",
  Login = "login",
  ConfaOverview = "confaOverview",
  ConfaEdit = "confaEdit",
  TalkOverview = "talkOverview",
  TalkOnline = "talkOnline",
  TalkEdit = "talkEdit",
}

export enum ConfaTab {
  Overview = "overview",
  Edit = "edit",
}

export enum TalkTab {
  Overview = "overview",
  Edit = "edit",
  Online = "online",
}

export const route = {
  login(): RouteLocationRaw {
    return {
      name: Name.Login,
    }
  },

  confa(confa: string, tab: ConfaTab): RouteLocationRaw {
    switch (tab) {
      case ConfaTab.Overview:
        return {
          name: Name.ConfaOverview,
          params: {
            confa: confa,
          },
        }
      case ConfaTab.Edit:
        return {
          name: Name.ConfaEdit,
          params: {
            confa: confa,
          },
        }
      default:
        return {
          name: Name.ConfaOverview,
          params: {
            confa: confa,
          },
        }
    }
  },

  talk(confa: string, talk: string, tab: TalkTab): RouteLocationRaw {
    switch (tab) {
      case TalkTab.Overview:
        return {
          name: Name.TalkOverview,
          params: {
            confa: confa,
            talk: talk,
          },
        }
      case TalkTab.Edit:
        return {
          name: Name.TalkEdit,
          params: {
            confa: confa,
            talk: talk,
          },
        }
      case TalkTab.Online:
        return {
          name: Name.TalkOnline,
          params: {
            confa: confa,
            talk: talk,
          },
        }
      default:
        return {
          name: Name.TalkOverview,
          params: {
            confa: confa,
            talk: talk,
          },
        }
    }
  },
}

const routes: Array<RouteRecordRaw> = [
  {
    path: "/",
    name: Name.Home,
    component: () => import("@/views/HomePage.vue"),
  },
  {
    path: "/login",
    name: Name.Login,
    props(to: RouteLocationNormalized): Record<string, string> {
      return {
        token: to.query.token as string,
      }
    },
    component: () => import("@/views/LoginPage.vue"),
  },
  {
    path: "/:confa",
    name: Name.ConfaOverview,
    props(to: RouteLocationNormalized): Record<string, string> {
      return {
        handle: to.params.confa as string,
        tab: ConfaTab.Overview,
      }
    },
    component: () => import("@/views/confa/Confa.vue"),
  },
  {
    path: "/:confa/ed",
    name: Name.ConfaEdit,
    props(to: RouteLocationNormalized): Record<string, string> {
      return {
        handle: to.params.confa as string,
        tab: ConfaTab.Edit,
      }
    },
    component: () => import("@/views/confa/Confa.vue"),
  },
  {
    path: "/:confa/:talk",
    name: Name.TalkOverview,
    props(to: RouteLocationNormalized): Record<string, string> {
      return {
        handle: to.params.talk as string,
        confaHandle: to.params.confa as string,
        tab: TalkTab.Overview,
      }
    },
    component: () => import("@/views/talk/Talk.vue"),
  },
  {
    path: "/:confa/:talk/on",
    name: Name.TalkOnline,
    props(to: RouteLocationNormalized): Record<string, string> {
      return {
        handle: to.params.talk as string,
        confaHandle: to.params.confa as string,
        tab: TalkTab.Online,
      }
    },
    component: () => import("@/views/talk/Talk.vue"),
  },
  {
    path: "/:confa/:talk/ed",
    name: Name.TalkEdit,
    props(to: RouteLocationNormalized): Record<string, string> {
      return {
        handle: to.params.talk as string,
        confaHandle: to.params.confa as string,
        tab: TalkTab.Edit,
      }
    },
    component: () => import("@/views/talk/Talk.vue"),
  },
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
})

export default router
