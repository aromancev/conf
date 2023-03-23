import { Store } from "@/platform/store"
import { RegexValidator } from "@/platform/validator"
import { genAvatar, genName } from "@/platform/gen"

export interface Profile extends Object {
  id: string
  ownerId: string
  handle: string
  givenName: string
  familyName: string
  avatarThumbnail: string
  avatarUrl: string
}

export class ProfileStore extends Store<Profile> {
  set(profile: Profile): void {
    super.set(profile)

    if (!profile.givenName) {
      this.reactive.givenName = genName(this.reactive.ownerId)
    }
    if (!profile.avatarThumbnail) {
      genAvatar(this.reactive.ownerId, 128).then((thumbnail) => {
        // Check again in case while avatar was generating the state has changed.
        if (this.reactive.avatarThumbnail) {
          return
        }
        this.reactive.avatarThumbnail = thumbnail
      })
    }
  }

  update(givenName: string, familyName: string, avatarThumbnail: string): void {
    this.reactive.givenName = givenName || this.reactive.givenName
    this.reactive.familyName = familyName || this.reactive.familyName
    this.reactive.avatarThumbnail = avatarThumbnail || this.reactive.avatarThumbnail
  }
}

export const profileStore = new ProfileStore({
  id: "",
  ownerId: "",
  handle: "",
  givenName: "",
  familyName: "",
  avatarThumbnail: "",
  avatarUrl: "",
})
export const handleValidator = new RegexValidator("^[a-z0-9-]{4,64}$", [
  "Must be from 4 to 64 characters long",
  "Can only contain lower case letters, numbers, and '-'",
])
export const nameValidator = new RegexValidator("^[a-zA-Z ]{0,64}$", [
  "Must be from 0 to 64 characters long",
  "Can only contain letters and spaces",
])
