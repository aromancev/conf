import { Store } from "@/platform/store"

interface ProfileMask {
  id?: string
  ownerId?: string
  handle?: string
  displayName?: string
  avatarThumbnail?: string
}

export interface Profile extends Object {
  id: string
  ownerId: string
  handle: string
  displayName: string
  avatarThumbnail: string
}

export class ProfileStore extends Store<Profile> {
  protected data(): Profile {
    return {
      id: "",
      ownerId: "",
      handle: "",
      displayName: "",
      avatarThumbnail: "",
    }
  }

  update(mask: ProfileMask): void {
    if (mask.id !== undefined) this._state.id = mask.id
    if (mask.ownerId !== undefined) this._state.ownerId = mask.ownerId
    if (mask.handle !== undefined) this._state.handle = mask.handle
    if (mask.displayName !== undefined) this._state.displayName = mask.displayName
    if (mask.avatarThumbnail !== undefined) this._state.avatarThumbnail = mask.avatarThumbnail
  }
}
