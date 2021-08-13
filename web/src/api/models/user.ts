import { Store } from "@/platform/store"

export enum Account {
  Guest = 0,
  User = 1,
  Admin = 2,
}

export interface User extends Object {
  id: string
  account: Account
  allowedWrite: boolean
}

export class UserStore extends Store<User> {
  protected data(): User {
    return {
      id: "",
      account: Account.Guest,
      allowedWrite: false,
    }
  }

  login(id: string, acc: Account): void {
    this.state.id = id
    this.state.account = acc
    this.state.allowedWrite = acc === Account.User || acc === Account.Admin
  }
}
