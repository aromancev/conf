import { Store } from "@/platform/store"
import { RegexValidator } from "@/platform/validator"
import { genAvatar, genName } from "@/platform/gen"

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
  hasAvatar: boolean
  avatarThumbnail: string
}

export class ProfileStore extends Store<Profile> {
  update(mask: ProfileMask): void {
    if (mask.id !== undefined) {
      this.reactive.id = mask.id
    }
    if (mask.ownerId !== undefined) {
      this.reactive.ownerId = mask.ownerId
    }
    if (mask.handle !== undefined) {
      this.reactive.handle = mask.handle
    }
    if (mask.displayName !== undefined) {
      this.reactive.displayName = mask.displayName || genName(this.reactive.ownerId)
    }
    if (mask.avatarThumbnail !== undefined) {
      this.reactive.avatarThumbnail = mask.avatarThumbnail
      if (mask.avatarThumbnail.length === 0) {
        genAvatar(this.reactive.ownerId, 128).then((thumbnail) => {
          // Check again in case while avatar was generating the state has changed.
          if (this.reactive.avatarThumbnail.length !== 0) {
            return
          }
          this.reactive.avatarThumbnail = thumbnail
        })
      }
    }
  }
}

export const profileStore = new ProfileStore({
  id: "",
  ownerId: "",
  handle: "",
  displayName: "",
  hasAvatar: false,
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
