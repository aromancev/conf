import { UserStore } from "./user"

export * from "./user"
export * from "./confa"
export * from "./talk"
export * from "./event"

export const userStore: UserStore = new UserStore()
