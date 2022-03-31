import { ProfileStore } from "./profile"
import { UserStore } from "./user"

export * from "./user"
export * from "./profile"
export * from "./confa"
export * from "./talk"

export const userStore = new UserStore()
export const profileStore = new ProfileStore()

export const currentUser = userStore.state()
export const currentProfile = profileStore.state()
