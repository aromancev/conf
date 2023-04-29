import { Store } from "@/platform/store"

const CHARS_PER_SECOND = 987 / 60

export type Level = "info" | "error"

export interface Notification {
  message: string
  level: Level
}

export class NotificationStore extends Store<Notification> {
  private timeoutId = 0

  info(message: string): void {
    this.message(message, "info")
  }

  error(message: string): void {
    this.message(message, "error")
  }

  message(message: string, level: Level): void {
    clearTimeout(this.timeoutId)
    this.reactive.message = message
    this.reactive.level = level
    this.timeoutId = window.setTimeout(() => {
      this.reactive.message = ""
      this.reactive.level = "info"
    }, (message.length / CHARS_PER_SECOND) * 1000)
  }
}

export const notificationStore = new NotificationStore({
  message: "",
  level: "info",
})
