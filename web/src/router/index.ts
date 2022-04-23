import { createRouter, createWebHistory, RouteRecordRaw, RouteLocationNormalized, RouteLocationRaw } from "vue-router"

export const handleNew = "new"

export type ProfileTab = "overview" | "edit"
export type ConfaTab = "overview" | "edit"
export type TalkTab = "overview" | "edit" | "online"

export const route = {
  home(): RouteLocationRaw {
    return {
      name: "home",
    }
  },

  login(): RouteLocationRaw {
    return {
      name: "login",
    }
  },

  profile(profile: string, tab: ProfileTab): RouteLocationRaw {
    return {
      name: "profile." + tab,
      params: {
        profile: profile,
      },
    }
  },

  confa(confa: string, tab: ConfaTab): RouteLocationRaw {
    return {
      name: "confa." + tab,
      params: {
        confa: confa,
      },
    }
  },

  talk(confa: string, talk: string, tab: TalkTab): RouteLocationRaw {
    return {
      name: "talk." + tab,
      params: {
        confa: confa,
        talk: talk,
      },
    }
  },
}

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
    path: "/pro/:profile",
    name: "profile.overview",
    props(to: RouteLocationNormalized): Record<string, string> {
      return {
        handle: to.params.profile as string,
        tab: "overview",
      }
    },
    component: () => import("@/views/profile/ProfileRoot.vue"),
  },
  {
    path: "/pro/:profile/ed",
    name: "profile.edit",
    props(to: RouteLocationNormalized): Record<string, string> {
      return {
        handle: to.params.profile as string,
        tab: "edit",
      }
    },
    component: () => import("@/views/profile/ProfileRoot.vue"),
  },
  {
    path: "/:confa",
    name: "confa.overview",
    props(to: RouteLocationNormalized): Record<string, string> {
      return {
        handle: to.params.confa as string,
        tab: "overview",
      }
    },
    component: () => import("@/views/confa/ConfaRoot.vue"),
  },
  {
    path: "/:confa/ed",
    name: "confa.edit",
    props(to: RouteLocationNormalized): Record<string, string> {
      return {
        handle: to.params.confa as string,
        tab: "edit",
      }
    },
    component: () => import("@/views/confa/ConfaRoot.vue"),
  },
  {
    path: "/:confa/:talk",
    name: "talk.overview",
    props(to: RouteLocationNormalized): Record<string, string> {
      return {
        handle: to.params.talk as string,
        confaHandle: to.params.confa as string,
        tab: "overview",
      }
    },
    component: () => import("@/views/talk/TalkRoot.vue"),
  },
  {
    path: "/:confa/:talk/on",
    name: "talk.online",
    props(to: RouteLocationNormalized): Record<string, string> {
      return {
        handle: to.params.talk as string,
        confaHandle: to.params.confa as string,
        tab: "online",
      }
    },
    component: () => import("@/views/talk/TalkRoot.vue"),
  },
  {
    path: "/:confa/:talk/ed",
    name: "talk.edit",
    props(to: RouteLocationNormalized): Record<string, string> {
      return {
        handle: to.params.talk as string,
        confaHandle: to.params.confa as string,
        tab: "edit",
      }
    },
    component: () => import("@/views/talk/TalkRoot.vue"),
  },
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
})

export default router
