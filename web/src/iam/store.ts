import { Store } from "@/platform/store"

export interface User extends Object {
  loggedIn: boolean
}

class UserStore extends Store<User> {
  protected data(): User {
    return {
      loggedIn: false,
    }
  }

  login() {
    this.state.loggedIn = true
  }
}

export const userStore: UserStore = new UserStore()
