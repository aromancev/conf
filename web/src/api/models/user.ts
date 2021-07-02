import { Store } from "@/platform/store"

export interface User extends Object {
  id: string
}

export class UserStore extends Store<User> {
  protected data(): User {
    return {
      id: "",
    }
  }

  set id(id: string) {
    this.state.id = id
  }
}
