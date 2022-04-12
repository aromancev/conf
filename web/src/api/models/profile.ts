import { Store } from "@/platform/store"

export interface Profile extends Object {
  id: string
  ownerId: string
  handle: string
  displayName: string
}

export class ProfileStore extends Store<Profile> {
  protected data(): Profile {
    return {
      id: "",
      ownerId: "",
      handle: "",
      displayName: "",
    }
  }

  update(prof: Profile): void {
    this._state.id = prof.id
    this._state.ownerId = prof.ownerId
    this._state.handle = prof.handle
    this._state.displayName = prof.displayName
  }
}
