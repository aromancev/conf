import { Store } from "@/platform/store"

export enum Account {
  Guest = 0,
  User = 1,
  Admin = 2,
}

export interface User {
  id: string
  account: Account
  allowedWrite: boolean
}

export class UserStore extends Store<User> {
  login(id: string, acc: Account): void {
    this._state.id = id
    this._state.account = acc
    this._state.allowedWrite = acc === Account.User || acc === Account.Admin
  }
}

export const userStore = new UserStore({
  id: "",
  account: Account.Guest,
  allowedWrite: false,
})
