import { Store } from "@/platform/store"
import { RegexValidator } from "@/platform/validator"

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
  update(mask: ProfileMask): void {
    if (mask.id !== undefined) this._state.id = mask.id
    if (mask.ownerId !== undefined) this._state.ownerId = mask.ownerId
    if (mask.handle !== undefined) this._state.handle = mask.handle
    if (mask.displayName !== undefined) this._state.displayName = mask.displayName
    if (mask.avatarThumbnail !== undefined) this._state.avatarThumbnail = mask.avatarThumbnail
  }
}

export const profileStore = new ProfileStore({
  id: "",
  ownerId: "",
  handle: "",
  displayName: "",
  avatarThumbnail: "",
})
export const handleValidator = new RegexValidator("^[a-z0-9-]{4,64}$", [
  "Must be from 4 to 64 characters long",
  "Can only contain lower case letters, numbers, and '-'",
])
export const displayNameValidator = new RegexValidator("^[a-zA-Z ]{0,64}$", [
  "Must be from 0 to 64 characters long",
  "Can only contain letters and spaces",
])
