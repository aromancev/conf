import { createRouter, createWebHistory, RouteRecordRaw, RouteLocationNormalized, RouteLocationRaw } from "vue-router"

export const handleNew = "new"

export type ProfileTab = "overview" | "edit" | "settings"
export type ConfaTab = "overview" | "edit"
export type TalkTab = "overview" | "edit" | "watch"
export type LoginAction = "login" | "create-password" | "reset-password"

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

  contentHub(): RouteLocationRaw {
    return {
      name: "contentHub",
    }
  },

  disclaimer(): RouteLocationRaw {
    return {
      name: "disclaimer",
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
    path: "/acc/login",
    name: "login",
    props(to: RouteLocationNormalized): Record<string, string> {
      return {
        action: to.query.action as string,
        token: to.query.token as string,
      }
    },
    component: () => import("@/views/LoginPage.vue"),
  },
  {
    path: "/hub",
    name: "contentHub",
    component: () => import("@/views/ContentHub.vue"),
  },
  {
    path: "/dis",
    name: "disclaimer",
    component: () => import("@/views/DisclaimerPage.vue"),
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
    path: "/pro/:profile/edit",
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
    path: "/pro/:profile/settings",
    name: "profile.settings",
    props(to: RouteLocationNormalized): Record<string, string> {
      return {
        handle: to.params.profile as string,
        tab: "settings",
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
    path: "/:confa/edit",
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
    path: "/:confa/:talk/overview",
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
    path: "/:confa/:talk",
    name: "talk.watch",
    props(to: RouteLocationNormalized): Record<string, string> {
      return {
        handle: to.params.talk as string,
        confaHandle: to.params.confa as string,
        tab: "watch",
      }
    },
    component: () => import("@/views/talk/TalkRoot.vue"),
  },
  {
    path: "/:confa/:talk/edit",
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
