import { api } from "@/api"
import { profileStore } from "@/api/models/profile"
import { ProfileClient } from "@/api/profile"
import { config } from "@/config"

declare global {
  interface Window {
    // https://developers.google.com/identity/gsi/web/guides/overview
    google: any // eslint-disable-line @typescript-eslint/no-explicit-any
  }
}

type PromptResponse = {
  credential: string
}

let script: Promise<HTMLScriptElement> | undefined = undefined

export type Theme = "outline" | "filled_blue" | "filled_black"

export async function gsiPromptOneTap() {
  if (!config.gsi.clientId) {
    return
  }
  await load(config.gsi.clientId)

  window.google.accounts.id.prompt()
}

export async function gsiRenderButton(el: HTMLElement, theme: Theme) {
  if (!config.gsi.clientId) {
    return
  }
  await load(config.gsi.clientId)

  window.google.accounts.id.renderButton(el, { theme: theme, size: "medium", type: "standard" })
}

type Token = {
  family_name: string
  given_name: string
  picture: string
}

async function load(clientId: string): Promise<void> {
  if (script) {
    await script
    return
  }
  script = new Promise((res) => {
    const el = document.createElement("script")
    el.src = "https://accounts.google.com/gsi/client"
    el.type = "text/javascript"
    el.async = true
    el.defer = true
    el.onload = () => {
      res(el)
    }
    document.head.appendChild(el)
  })
  await script
  if (!window.google) {
    throw new Error("Failed to initialize google prompt.")
  }
  window.google.accounts.id.initialize({
    client_id: clientId,
    callback: async (resp: PromptResponse) => {
      const token = resp.credential
      await api.createSessionWithGSI(token)
      // Make sure we have an updated profile.
      // Request is doubled from `main.ts` but graphql is caching it in memory.
      await new ProfileClient(api).refreshProfile()
      if (!profileStore.state.id) {
        const p = JSON.parse(window.atob(token.split(".")[1])) as Token
        profileStore.update(p.given_name, p.family_name, p.picture)
      }
    },
    auto_select: false,
    cancel_on_tap_outside: true,
    context: "signin",
  })
}
