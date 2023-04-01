import { Store } from "@/platform/store"

export type Theme = "light" | "dark"

export interface Style extends Object {
  theme: Theme
}

export class StyleStore extends Store<Style> {
  setTheme(theme: Theme) {
    this.reactive.theme = theme
    localStorage.setItem("theme", theme)
  }
}

export const styleStore = new StyleStore({
  theme: (localStorage.getItem("theme") as Theme) || "light",
})
